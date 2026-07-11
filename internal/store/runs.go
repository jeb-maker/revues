package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	RunStatusDraft      = "draft"
	RunStatusInProgress = "in_progress"
	RunStatusDone       = "done"
	RunStatusArchived   = "archived"
)

// ErrRunNotFound is returned when a run id does not exist.
var ErrRunNotFound = errors.New("run not found")

// ErrInvalidRunStatus is returned when a status transition is not allowed.
var ErrInvalidRunStatus = errors.New("invalid run status transition")

// ChecklistRun is a review execution instance.
type ChecklistRun struct {
	ID                int64
	ProjectID         int64
	TemplateVersionID int64
	Title             string
	Status            string
	DueDate           sql.NullString
	ClosingNote       string
	CreatedBy         sql.NullInt64
	StartedAt         sql.NullString
	CompletedAt       sql.NullString
	NotionURL         string
	CreatedAt         string
}

// RunItem is a snapshot checklist point within a run.
type RunItem struct {
	ID            int64
	RunID         int64
	SourceItemID  sql.NullInt64
	Section       string
	Position      int
	Label         string
	HelpText      string
	Required      bool
	Status        string
	Comment       string
	AssignedTo    sql.NullInt64
	AssignedLogin string
	UpdatedAt     string
}

// AssignedRunItemSummary is a task row for the my tasks view.
type AssignedRunItemSummary struct {
	RunItem
	RunTitle    string
	ProjectID   int64
	ProjectName string
}

// CreateChecklistRun inserts a run and snapshots template items in one transaction.
func (s *Store) CreateChecklistRun(ctx context.Context, projectID, templateID int64, title string, createdBy int64, dueDate sql.NullString) (*ChecklistRun, error) {
	template, err := s.ChecklistTemplateByID(ctx, templateID)
	if err != nil {
		return nil, err
	}
	if template.ArchivedAt.Valid {
		return nil, ErrChecklistTemplateNotFound
	}
	matches, err := s.TemplateMatchesProject(ctx, projectID, templateID)
	if err != nil {
		return nil, fmt.Errorf("template matches project: %w", err)
	}
	if !matches {
		return nil, ErrChecklistTemplateNotFound
	}

	version, err := s.LatestTemplateVersion(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("latest template version: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	res, err := tx.ExecContext(ctx, `
		INSERT INTO checklist_runs (
			project_id, template_version_id, title, status, due_date, created_by, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`, projectID, version.ID, title, RunStatusDraft, dueDate, createdBy, now)
	if err != nil {
		return nil, fmt.Errorf("insert checklist run: %w", err)
	}

	runID, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("run id: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO run_items (
			run_id, source_item_id, section, position, label, help_text, required, status, comment, updated_at
		)
		SELECT ?, id, section, position, label, help_text, required, 'pending', '', ?
		FROM template_items
		WHERE version_id = ?
	`, runID, now, version.ID)
	if err != nil {
		return nil, fmt.Errorf("snapshot run items: %w", err)
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return nil, fmt.Errorf("commit create checklist run: %w", commitErr)
	}

	return s.RunByID(ctx, runID)
}

// RunByID loads a run by primary key.
func (s *Store) RunByID(ctx context.Context, id int64) (*ChecklistRun, error) {
	var run ChecklistRun
	err := s.db.QueryRowContext(ctx, `
		SELECT id, project_id, template_version_id, title, status, due_date, closing_note,
		       created_by, started_at, completed_at, notion_url, created_at
		FROM checklist_runs
		WHERE id = ?
	`, id).Scan(
		&run.ID, &run.ProjectID, &run.TemplateVersionID, &run.Title, &run.Status, &run.DueDate,
		&run.ClosingNote, &run.CreatedBy, &run.StartedAt, &run.CompletedAt, &run.NotionURL, &run.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRunNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("run by id: %w", err)
	}
	return &run, nil
}

// ListRunsByProject returns runs for a project ordered by recency.
func (s *Store) ListRunsByProject(ctx context.Context, projectID int64) ([]ChecklistRun, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, project_id, template_version_id, title, status, due_date, closing_note,
		       created_by, started_at, completed_at, notion_url, created_at
		FROM checklist_runs
		WHERE project_id = ? AND status != ?
		ORDER BY created_at DESC
	`, projectID, RunStatusArchived)
	if err != nil {
		return nil, fmt.Errorf("list runs by project: %w", err)
	}
	defer rows.Close()

	var runs []ChecklistRun
	for rows.Next() {
		var run ChecklistRun
		if err := rows.Scan(
			&run.ID, &run.ProjectID, &run.TemplateVersionID, &run.Title, &run.Status, &run.DueDate,
			&run.ClosingNote, &run.CreatedBy, &run.StartedAt, &run.CompletedAt, &run.NotionURL, &run.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan run: %w", err)
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate runs: %w", err)
	}

	return runs, nil
}

// ListRunItems returns ordered items for a run.
func (s *Store) ListRunItems(ctx context.Context, runID int64) ([]RunItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT ri.id, ri.run_id, ri.source_item_id, ri.section, ri.position, ri.label, ri.help_text, ri.required,
		       ri.status, ri.comment, ri.assigned_to, u.login, ri.updated_at
		FROM run_items ri
		LEFT JOIN users u ON u.id = ri.assigned_to
		WHERE ri.run_id = ?
		ORDER BY ri.position
	`, runID)
	if err != nil {
		return nil, fmt.Errorf("list run items: %w", err)
	}
	defer rows.Close()

	var items []RunItem
	for rows.Next() {
		item, err := scanRunItemRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scan run item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate run items: %w", err)
	}

	return items, nil
}

// StartRun moves a run from draft to in_progress.
func (s *Store) StartRun(ctx context.Context, id int64) error {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx, `
		UPDATE checklist_runs
		SET status = ?, started_at = ?
		WHERE id = ? AND status = ?
	`, RunStatusInProgress, now, id, RunStatusDraft)
	if err != nil {
		return fmt.Errorf("start run: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("start run rows: %w", err)
	}
	if n == 0 {
		run, loadErr := s.RunByID(ctx, id)
		if loadErr != nil {
			return loadErr
		}
		if run.Status != RunStatusDraft {
			return ErrInvalidRunStatus
		}
		return ErrRunNotFound
	}
	return nil
}

// ListRunsDueOn returns in-progress runs whose due_date starts with datePrefix (YYYY-MM-DD).
func (s *Store) ListRunsDueOn(ctx context.Context, datePrefix string) ([]ChecklistRun, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, project_id, template_version_id, title, status, due_date, closing_note,
		       created_by, started_at, completed_at, notion_url, created_at
		FROM checklist_runs
		WHERE status = ? AND due_date IS NOT NULL AND due_date LIKE ?
		ORDER BY due_date, id
	`, RunStatusInProgress, datePrefix+"%")
	if err != nil {
		return nil, fmt.Errorf("list runs due on: %w", err)
	}
	defer rows.Close()

	var runs []ChecklistRun
	for rows.Next() {
		var run ChecklistRun
		if err := rows.Scan(
			&run.ID, &run.ProjectID, &run.TemplateVersionID, &run.Title, &run.Status, &run.DueDate,
			&run.ClosingNote, &run.CreatedBy, &run.StartedAt, &run.CompletedAt, &run.NotionURL, &run.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan run due on: %w", err)
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate runs due on: %w", err)
	}

	return runs, nil
}

// SetRunDueDate updates due_date on a run (used in tests).
func (s *Store) SetRunDueDate(ctx context.Context, runID int64, dueDate sql.NullString) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE checklist_runs SET due_date = ? WHERE id = ?
	`, dueDate, runID)
	if err != nil {
		return fmt.Errorf("set run due date: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("set run due date rows: %w", err)
	}
	if n == 0 {
		return ErrRunNotFound
	}
	return nil
}

// CompleteRun moves a run from in_progress to done with a closing note.
func (s *Store) CompleteRun(ctx context.Context, id int64, closingNote string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx, `
		UPDATE checklist_runs
		SET status = ?, completed_at = ?, closing_note = ?
		WHERE id = ? AND status = ?
	`, RunStatusDone, now, closingNote, id, RunStatusInProgress)
	if err != nil {
		return fmt.Errorf("complete run: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("complete run rows: %w", err)
	}
	if n == 0 {
		run, loadErr := s.RunByID(ctx, id)
		if loadErr != nil {
			return loadErr
		}
		if run.Status != RunStatusInProgress {
			return ErrInvalidRunStatus
		}
		return ErrRunNotFound
	}
	return nil
}

// SetRunNotionURL stores the Notion page URL for an exported run.
func (s *Store) SetRunNotionURL(ctx context.Context, runID int64, notionURL string) error {
	res, err := s.db.ExecContext(ctx, `UPDATE checklist_runs SET notion_url = ? WHERE id = ?`, strings.TrimSpace(notionURL), runID)
	if err != nil {
		return fmt.Errorf("set run notion url: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("set run notion url rows: %w", err)
	}
	if n == 0 {
		return ErrRunNotFound
	}
	return nil
}
