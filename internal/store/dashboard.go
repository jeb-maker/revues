package store

import (
	"context"
	"database/sql"
	"fmt"
)

const (
	recentCompletedRunsLimit = 10
	filteredRunsLimit        = 100
)

// ActiveRunSummary is a draft or in-progress run with completion stats for the dashboard.
type ActiveRunSummary struct {
	RunID       int64
	Title       string
	ProjectID   int64
	ProjectName string
	Status      string
	DueDate     sql.NullString
	Done        int
	Total       int
	Percent     int
}

// RunListSummary is a non-archived run row for the unified /revues table.
type RunListSummary struct {
	RunID          int64
	Title          string
	ProjectID      int64
	ProjectName    string
	Status         string
	DueDate        sql.NullString
	CreatedAt      string
	StartedAt      sql.NullString
	CompletedAt    sql.NullString
	CreatedByLogin sql.NullString
	Done           int
	Total          int
	Percent        int
}

// CompletedRunSummary is a recently closed run for the dashboard.
type CompletedRunSummary struct {
	RunID       int64
	Title       string
	ProjectID   int64
	ProjectName string
	CompletedAt sql.NullString
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

// TemplateIndexRow is a checklist template for the global index.
type TemplateIndexRow struct {
	ChecklistTemplateSummary
}

func progressPercent(done, total int) int {
	if total == 0 {
		return 0
	}
	return done * 100 / total
}

// ListActiveRunSummaries returns draft and in-progress runs visible to the user with completion stats.
func (s *Store) ListActiveRunSummaries(ctx context.Context, userID int64, admin bool) ([]ActiveRunSummary, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var (
		rows *sqlRowsWrapper
	)

	if admin {
		rows, err = s.queryRows(ctx, activeRunSummariesSQL+`
		WHERE r.status IN (?, ?) AND p.archived_at IS NULL AND p.organization_id = ?
		GROUP BY r.id, r.title, r.project_id, p.name, r.due_date, r.status
		ORDER BY CASE r.status WHEN ? THEN 0 ELSE 1 END, COALESCE(r.started_at, r.created_at) DESC
		`, RunStatusDraft, RunStatusInProgress, orgID, RunStatusInProgress)
	} else {
		rows, err = s.queryRows(ctx, activeRunSummariesSQL+`
		INNER JOIN project_members pm ON pm.project_id = p.id AND pm.user_id = ?
		WHERE r.status IN (?, ?) AND p.archived_at IS NULL AND p.organization_id = ?
		GROUP BY r.id, r.title, r.project_id, p.name, r.due_date, r.status
		ORDER BY CASE r.status WHEN ? THEN 0 ELSE 1 END, COALESCE(r.started_at, r.created_at) DESC
		`, userID, RunStatusDraft, RunStatusInProgress, orgID, RunStatusInProgress)
	}
	if err != nil {
		return nil, fmt.Errorf("list active run summaries: %w", err)
	}
	defer rows.Close()

	return scanActiveRunSummaries(rows)
}

// ListFilteredRunSummaries returns non-archived runs visible to the user with optional filters.
func (s *Store) ListFilteredRunSummaries(ctx context.Context, userID int64, admin bool, status, query string) ([]RunListSummary, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	sqlQuery := runListSummariesSQL
	var args []any

	if admin {
		sqlQuery += `
		WHERE r.status != ? AND p.archived_at IS NULL AND p.organization_id = ?`
		args = append(args, RunStatusArchived, orgID)
	} else {
		sqlQuery += `
		INNER JOIN project_members pm ON pm.project_id = p.id AND pm.user_id = ?
		WHERE r.status != ? AND p.archived_at IS NULL AND p.organization_id = ?`
		args = append(args, userID, RunStatusArchived, orgID)
	}

	if status != "" {
		sqlQuery += ` AND r.status = ?`
		args = append(args, status)
	} else {
		sqlQuery += ` AND r.status IN (?, ?, ?)`
		args = append(args, RunStatusDraft, RunStatusInProgress, RunStatusDone)
	}

	for _, term := range searchTerms(query) {
		pattern := likeContainsPattern(term)
		sqlQuery += ` AND (
			r.title LIKE ? ESCAPE '\'
			OR p.name LIKE ? ESCAPE '\'
			OR COALESCE(u.login, '') LIKE ? ESCAPE '\'
		)`
		args = append(args, pattern, pattern, pattern)
	}

	sqlQuery += `
		GROUP BY r.id, r.title, r.project_id, p.name, r.status, r.due_date,
		         r.created_at, r.started_at, r.completed_at, u.login
		ORDER BY COALESCE(r.completed_at, r.started_at, r.created_at) DESC, r.id DESC
		LIMIT ?`
	args = append(args, filteredRunsLimit)

	rows, err := s.queryRows(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("list filtered run summaries: %w", err)
	}
	defer rows.Close()

	return scanRunListSummaries(rows)
}

// ListRecentCompletedRunSummaries returns the most recently closed runs visible to the user.
func (s *Store) ListRecentCompletedRunSummaries(ctx context.Context, userID int64, admin bool) ([]CompletedRunSummary, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var rows *sqlRowsWrapper
	if admin {
		rows, err = s.queryRows(ctx, completedRunSummariesSQL+`
		WHERE r.status = ? AND p.archived_at IS NULL AND p.organization_id = ?
		GROUP BY r.id, r.title, r.project_id, p.name, r.completed_at
		ORDER BY r.completed_at DESC, r.id DESC
		LIMIT ?
		`, RunStatusDone, orgID, recentCompletedRunsLimit)
	} else {
		rows, err = s.queryRows(ctx, completedRunSummariesSQL+`
		INNER JOIN project_members pm ON pm.project_id = p.id AND pm.user_id = ?
		WHERE r.status = ? AND p.archived_at IS NULL AND p.organization_id = ?
		GROUP BY r.id, r.title, r.project_id, p.name, r.completed_at
		ORDER BY r.completed_at DESC, r.id DESC
		LIMIT ?
		`, userID, RunStatusDone, orgID, recentCompletedRunsLimit)
	}
	if err != nil {
		return nil, fmt.Errorf("list recent completed run summaries: %w", err)
	}
	defer rows.Close()

	return scanCompletedRunSummaries(rows)
}

const activeRunSummariesSQL = `
	SELECT r.id, r.title, r.project_id, p.name, r.due_date, r.status,
	       COUNT(ri.id) AS total,
	       SUM(CASE WHEN ri.status IN ('ok', 'na') THEN 1 ELSE 0 END) AS done
	FROM checklist_runs r
	INNER JOIN projects p ON p.id = r.project_id
	INNER JOIN run_items ri ON ri.run_id = r.id
`

const runListSummariesSQL = `
	SELECT r.id, r.title, r.project_id, p.name, r.status, r.due_date,
	       r.created_at, r.started_at, r.completed_at, u.login,
	       COUNT(ri.id) AS total,
	       SUM(CASE WHEN ri.status IN ('ok', 'na') THEN 1 ELSE 0 END) AS done
	FROM checklist_runs r
	INNER JOIN projects p ON p.id = r.project_id
	LEFT JOIN users u ON u.id = r.created_by
	LEFT JOIN run_items ri ON ri.run_id = r.id
`

const completedRunSummariesSQL = `
	SELECT r.id, r.title, r.project_id, p.name, r.completed_at,
	       COUNT(ri.id) AS total,
	       SUM(CASE WHEN ri.status IN ('ok', 'na') THEN 1 ELSE 0 END) AS done
	FROM checklist_runs r
	INNER JOIN projects p ON p.id = r.project_id
	LEFT JOIN run_items ri ON ri.run_id = r.id
`

func scanRunListSummaries(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]RunListSummary, error) {
	var summaries []RunListSummary
	for rows.Next() {
		var summary RunListSummary
		if err := rows.Scan(
			&summary.RunID, &summary.Title, &summary.ProjectID, &summary.ProjectName, &summary.Status, &summary.DueDate,
			&summary.CreatedAt, &summary.StartedAt, &summary.CompletedAt, &summary.CreatedByLogin,
			&summary.Total, &summary.Done,
		); err != nil {
			return nil, fmt.Errorf("scan run list summary: %w", err)
		}
		summary.Percent = progressPercent(summary.Done, summary.Total)
		summaries = append(summaries, summary)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate run list summaries: %w", err)
	}
	return summaries, nil
}

func scanActiveRunSummaries(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]ActiveRunSummary, error) {
	var summaries []ActiveRunSummary
	for rows.Next() {
		var summary ActiveRunSummary
		if err := rows.Scan(
			&summary.RunID, &summary.Title, &summary.ProjectID, &summary.ProjectName, &summary.DueDate, &summary.Status,
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

func scanCompletedRunSummaries(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]CompletedRunSummary, error) {
	var summaries []CompletedRunSummary
	for rows.Next() {
		var summary CompletedRunSummary
		if err := rows.Scan(
			&summary.RunID, &summary.Title, &summary.ProjectID, &summary.ProjectName, &summary.CompletedAt,
			&summary.Total, &summary.Done,
		); err != nil {
			return nil, fmt.Errorf("scan completed run summary: %w", err)
		}
		summary.Percent = progressPercent(summary.Done, summary.Total)
		summaries = append(summaries, summary)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate completed run summaries: %w", err)
	}
	return summaries, nil
}

// ListRunsWithProgressByProject returns non-archived runs with completion stats.
func (s *Store) ListRunsWithProgressByProject(ctx context.Context, projectID int64) ([]RunWithProgress, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT r.id, r.project_id, r.template_version_id, r.title, r.status, r.due_date, r.closing_note,
		       r.created_by, r.started_at, r.completed_at, r.notion_url, r.created_at,
		       COUNT(ri.id) AS total,
		       SUM(CASE WHEN ri.status IN ('ok', 'na') THEN 1 ELSE 0 END) AS done
		FROM checklist_runs r
		LEFT JOIN run_items ri ON ri.run_id = r.id
		WHERE r.project_id = ? AND r.status != ?
		GROUP BY r.id, r.project_id, r.template_version_id, r.title, r.status, r.due_date, r.closing_note,
		         r.created_by, r.started_at, r.completed_at, r.notion_url, r.created_at
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
			&run.CreatedBy, &run.StartedAt, &run.CompletedAt, &run.NotionURL, &run.CreatedAt,
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

// ListTemplateIndex returns all active templates for the global catalog.
func (s *Store) ListTemplateIndex(ctx context.Context, userID int64, admin bool, query string) ([]TemplateIndexRow, error) {
	_ = userID
	_ = admin

	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	sqlQuery := `
		SELECT
			t.id, t.organization_id, t.project_id, t.name, t.archived_at, t.created_at,
			v.version,
			COUNT(i.id) AS item_count
		FROM checklist_templates t
		INNER JOIN template_versions v ON v.template_id = t.id
		LEFT JOIN template_items i ON i.version_id = v.id
		WHERE t.archived_at IS NULL AND t.organization_id = ?
		  AND v.version = (
			SELECT MAX(v2.version) FROM template_versions v2 WHERE v2.template_id = t.id
		  )`
	args := []any{orgID}

	for _, term := range searchTerms(query) {
		pattern := likeContainsPattern(term)
		sqlQuery += ` AND (
			t.name LIKE ? ESCAPE '\'
			OR EXISTS (
				SELECT 1 FROM template_tags tt
				WHERE tt.template_id = t.id AND tt.tag LIKE ? ESCAPE '\'
			)
		)`
		args = append(args, pattern, pattern)
	}

	sqlQuery += `
		GROUP BY t.id, t.organization_id, t.project_id, t.name, t.archived_at, t.created_at, v.version
		ORDER BY t.name`

	rows, err := s.queryRows(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("list template index: %w", err)
	}
	defer rows.Close()

	var templates []TemplateIndexRow
	for rows.Next() {
		var row TemplateIndexRow
		if err := rows.Scan(
			&row.ID, &row.OrganizationID, &row.ProjectID, &row.Name, &row.ArchivedAt, &row.CreatedAt,
			&row.LatestVersion, &row.ItemCount,
		); err != nil {
			return nil, fmt.Errorf("scan template index row: %w", err)
		}
		templates = append(templates, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate template index: %w", err)
	}

	for i := range templates {
		tags, err := s.ListTemplateTags(ctx, templates[i].ID)
		if err != nil {
			return nil, err
		}
		templates[i].Tags = tags
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
