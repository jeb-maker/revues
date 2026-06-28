package store

import (
	"context"
	"database/sql"
	"fmt"
)

// ActiveRunSummary is an in-progress run with completion stats for the dashboard.
type ActiveRunSummary struct {
	RunID       int64
	Title       string
	ProjectID   int64
	ProjectName string
	Done        int
	Total       int
	Percent     int
}

// RunWithProgress is a run row with item completion stats.
type RunWithProgress struct {
	ChecklistRun
	Done    int
	Total   int
	Percent int
}

// ProjectNokItemSummary is a blocking nok point on an active project run.
type ProjectNokItemSummary struct {
	RunItem
	RunID    int64
	RunTitle string
}

// TemplateIndexRow is a checklist template with project context for the global index.
type TemplateIndexRow struct {
	ChecklistTemplateSummary
	ProjectName string
}

func progressPercent(done, total int) int {
	if total == 0 {
		return 0
	}
	return done * 100 / total
}

// ListActiveRunSummaries returns in-progress runs visible to the user with completion stats.
func (s *Store) ListActiveRunSummaries(ctx context.Context, userID int64, admin bool) ([]ActiveRunSummary, error) {
	var (
		rows *sqlRowsWrapper
		err  error
	)

	if admin {
		rows, err = s.queryRows(ctx, activeRunSummariesSQL+`
		WHERE r.status = ? AND p.archived_at IS NULL
		GROUP BY r.id, r.title, r.project_id, p.name
		ORDER BY r.started_at DESC, r.created_at DESC
		`, RunStatusInProgress)
	} else {
		rows, err = s.queryRows(ctx, activeRunSummariesSQL+`
		INNER JOIN project_members pm ON pm.project_id = p.id AND pm.user_id = ?
		WHERE r.status = ? AND p.archived_at IS NULL
		GROUP BY r.id, r.title, r.project_id, p.name
		ORDER BY r.started_at DESC, r.created_at DESC
		`, userID, RunStatusInProgress)
	}
	if err != nil {
		return nil, fmt.Errorf("list active run summaries: %w", err)
	}
	defer rows.Close()

	return scanActiveRunSummaries(rows)
}

const activeRunSummariesSQL = `
	SELECT r.id, r.title, r.project_id, p.name,
	       COUNT(ri.id) AS total,
	       SUM(CASE WHEN ri.status IN ('ok', 'na') THEN 1 ELSE 0 END) AS done
	FROM checklist_runs r
	INNER JOIN projects p ON p.id = r.project_id
	INNER JOIN run_items ri ON ri.run_id = r.id
`

func scanActiveRunSummaries(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]ActiveRunSummary, error) {
	var summaries []ActiveRunSummary
	for rows.Next() {
		var summary ActiveRunSummary
		if err := rows.Scan(
			&summary.RunID, &summary.Title, &summary.ProjectID, &summary.ProjectName,
			&summary.Total, &summary.Done,
		); err != nil {
			return nil, fmt.Errorf("scan active run summary: %w", err)
		}
		summary.Percent = progressPercent(summary.Done, summary.Total)
		summaries = append(summaries, summary)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active run summaries: %w", err)
	}
	return summaries, nil
}

// ListRunsWithProgressByProject returns non-archived runs with completion stats.
func (s *Store) ListRunsWithProgressByProject(ctx context.Context, projectID int64) ([]RunWithProgress, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT r.id, r.project_id, r.template_version_id, r.title, r.status, r.due_date, r.closing_note,
		       r.created_by, r.started_at, r.completed_at, r.created_at,
		       COUNT(ri.id) AS total,
		       SUM(CASE WHEN ri.status IN ('ok', 'na') THEN 1 ELSE 0 END) AS done
		FROM checklist_runs r
		LEFT JOIN run_items ri ON ri.run_id = r.id
		WHERE r.project_id = ? AND r.status != ?
		GROUP BY r.id, r.project_id, r.template_version_id, r.title, r.status, r.due_date, r.closing_note,
		         r.created_by, r.started_at, r.completed_at, r.created_at
		ORDER BY r.created_at DESC
	`, projectID, RunStatusArchived)
	if err != nil {
		return nil, fmt.Errorf("list runs with progress: %w", err)
	}
	defer rows.Close()

	var runs []RunWithProgress
	for rows.Next() {
		var run RunWithProgress
		if err := rows.Scan(
			&run.ID, &run.ProjectID, &run.TemplateVersionID, &run.Title, &run.Status, &run.DueDate, &run.ClosingNote,
			&run.CreatedBy, &run.StartedAt, &run.CompletedAt, &run.CreatedAt,
			&run.Total, &run.Done,
		); err != nil {
			return nil, fmt.Errorf("scan run with progress: %w", err)
		}
		run.Percent = progressPercent(run.Done, run.Total)
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate runs with progress: %w", err)
	}
	return runs, nil
}

// ListProjectNokItems returns nok points on in-progress runs for a project.
func (s *Store) ListProjectNokItems(ctx context.Context, projectID int64) ([]ProjectNokItemSummary, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT ri.id, ri.run_id, ri.source_item_id, ri.section, ri.position, ri.label, ri.help_text, ri.required,
		       ri.status, ri.comment, ri.assigned_to, u.login, ri.updated_at,
		       r.title
		FROM run_items ri
		INNER JOIN checklist_runs r ON r.id = ri.run_id
		LEFT JOIN users u ON u.id = ri.assigned_to
		WHERE r.project_id = ? AND r.status = ? AND ri.status = 'nok'
		ORDER BY r.created_at DESC, ri.position
	`, projectID, RunStatusInProgress)
	if err != nil {
		return nil, fmt.Errorf("list project nok items: %w", err)
	}
	defer rows.Close()

	var items []ProjectNokItemSummary
	for rows.Next() {
		var item ProjectNokItemSummary
		var required int
		var assignedLogin sql.NullString
		if err := rows.Scan(
			&item.ID, &item.RunID, &item.SourceItemID, &item.Section, &item.Position,
			&item.Label, &item.HelpText, &required, &item.Status, &item.Comment,
			&item.AssignedTo, &assignedLogin, &item.UpdatedAt,
			&item.RunTitle,
		); err != nil {
			return nil, fmt.Errorf("scan project nok item: %w", err)
		}
		item.Required = required == 1
		if assignedLogin.Valid {
			item.AssignedLogin = assignedLogin.String
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate project nok items: %w", err)
	}
	return items, nil
}

// ListTemplateIndex returns active templates across projects visible to the user.
func (s *Store) ListTemplateIndex(ctx context.Context, userID int64, admin bool) ([]TemplateIndexRow, error) {
	baseSQL := `
		SELECT
			t.id, t.project_id, t.name, t.archived_at, t.created_at,
			v.version,
			COUNT(i.id) AS item_count,
			p.name
		FROM checklist_templates t
		INNER JOIN projects p ON p.id = t.project_id
		INNER JOIN template_versions v ON v.template_id = t.id
		LEFT JOIN template_items i ON i.version_id = v.id
		WHERE t.archived_at IS NULL AND p.archived_at IS NULL
		  AND v.version = (
			SELECT MAX(v2.version) FROM template_versions v2 WHERE v2.template_id = t.id
		  )
	`

	var rows *sqlRowsWrapper
	var err error
	if admin {
		rows, err = s.queryRows(ctx, baseSQL+`
		GROUP BY t.id, t.project_id, t.name, t.archived_at, t.created_at, v.version, p.name
		ORDER BY p.name, t.name
		`)
	} else {
		rows, err = s.queryRows(ctx, baseSQL+`
		  AND EXISTS (
			SELECT 1 FROM project_members pm
			WHERE pm.project_id = p.id AND pm.user_id = ?
		  )
		GROUP BY t.id, t.project_id, t.name, t.archived_at, t.created_at, v.version, p.name
		ORDER BY p.name, t.name
		`, userID)
	}
	if err != nil {
		return nil, fmt.Errorf("list template index: %w", err)
	}
	defer rows.Close()

	var templates []TemplateIndexRow
	for rows.Next() {
		var row TemplateIndexRow
		if err := rows.Scan(
			&row.ID, &row.ProjectID, &row.Name, &row.ArchivedAt, &row.CreatedAt,
			&row.LatestVersion, &row.ItemCount,
			&row.ProjectName,
		); err != nil {
			return nil, fmt.Errorf("scan template index row: %w", err)
		}
		templates = append(templates, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate template index: %w", err)
	}
	return templates, nil
}

type sqlRowsWrapper struct {
	rows interface {
		Close() error
		Next() bool
		Scan(dest ...any) error
		Err() error
	}
}

func (r *sqlRowsWrapper) Close() error           { return r.rows.Close() }
func (r *sqlRowsWrapper) Next() bool             { return r.rows.Next() }
func (r *sqlRowsWrapper) Scan(dest ...any) error { return r.rows.Scan(dest...) }
func (r *sqlRowsWrapper) Err() error             { return r.rows.Err() }

func (s *Store) queryRows(ctx context.Context, query string, args ...any) (*sqlRowsWrapper, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &sqlRowsWrapper{rows: rows}, nil
}
