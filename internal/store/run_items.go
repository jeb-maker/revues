package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrRunItemNotFound is returned when a run item id does not exist.
var ErrRunItemNotFound = errors.New("run item not found")

// ErrRunNotEditable is returned when a run cannot accept item updates.
var ErrRunNotEditable = errors.New("run not editable")

// ErrInvalidAssignee is returned when assignee is not a project member.
var ErrInvalidAssignee = errors.New("invalid assignee")

const (
	RunItemStatusPending = "pending"
	RunItemStatusOK      = "ok"
	RunItemStatusNOK     = "nok"
	RunItemStatusNA      = "na"
)

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRunItemRow(row rowScanner) (RunItem, error) {
	var item RunItem
	var required int
	var assignedLogin sql.NullString
	err := row.Scan(
		&item.ID, &item.RunID, &item.SourceItemID, &item.Section, &item.Position,
		&item.Label, &item.HelpText, &required, &item.Status, &item.Comment,
		&item.AssignedTo, &assignedLogin, &item.UpdatedAt,
	)
	if err != nil {
		return RunItem{}, err
	}
	item.Required = required == 1
	if assignedLogin.Valid {
		item.AssignedLogin = assignedLogin.String
	}
	return item, nil
}

// RunItemByID loads an item scoped to a run.
func (s *Store) RunItemByID(ctx context.Context, runID, itemID int64) (*RunItem, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT ri.id, ri.run_id, ri.source_item_id, ri.section, ri.position, ri.label, ri.help_text, ri.required,
		       ri.status, ri.comment, ri.assigned_to, u.login, ri.updated_at
		FROM run_items ri
		LEFT JOIN users u ON u.id = ri.assigned_to
		WHERE ri.id = ? AND ri.run_id = ?
	`, itemID, runID)
	item, err := scanRunItemRow(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRunItemNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("run item by id: %w", err)
	}
	return &item, nil
}

// UpdateRunItemStatus changes status and comment on an in-progress run item.
func (s *Store) UpdateRunItemStatus(ctx context.Context, runID, itemID, userID int64, status, comment string) error {
	run, err := s.RunByID(ctx, runID)
	if err != nil {
		return err
	}
	if run.Status != RunStatusInProgress {
		return ErrRunNotEditable
	}

	now := time.Now().UTC().Format(time.RFC3339)
	comment = strings.TrimSpace(comment)

	var checkedBy sql.NullInt64
	var checkedAt sql.NullString
	if status != RunItemStatusPending {
		checkedBy = sql.NullInt64{Int64: userID, Valid: true}
		checkedAt = sql.NullString{String: now, Valid: true}
	}

	res, err := s.db.ExecContext(ctx, `
		UPDATE run_items
		SET status = ?, comment = ?, checked_by = ?, checked_at = ?, updated_at = ?
		WHERE id = ? AND run_id = ?
	`, status, comment, checkedBy, checkedAt, now, itemID, runID)
	if err != nil {
		return fmt.Errorf("update run item: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update run item rows: %w", err)
	}
	if n == 0 {
		return ErrRunItemNotFound
	}
	return nil
}

// AssignRunItem sets or clears assignee on an in-progress run item.
func (s *Store) AssignRunItem(ctx context.Context, runID, itemID int64, assigneeID *int64) error {
	run, err := s.RunByID(ctx, runID)
	if err != nil {
		return err
	}
	if run.Status != RunStatusInProgress {
		return ErrRunNotEditable
	}

	var assignedTo sql.NullInt64
	if assigneeID != nil {
		_, isMember, memberErr := s.MemberRole(ctx, run.ProjectID, *assigneeID)
		if memberErr != nil {
			return fmt.Errorf("member role for assignee: %w", memberErr)
		}
		if !isMember {
			return ErrInvalidAssignee
		}
		assignedTo = sql.NullInt64{Int64: *assigneeID, Valid: true}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx, `
		UPDATE run_items SET assigned_to = ?, updated_at = ?
		WHERE id = ? AND run_id = ?
	`, assignedTo, now, itemID, runID)
	if err != nil {
		return fmt.Errorf("assign run item: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("assign run item rows: %w", err)
	}
	if n == 0 {
		return ErrRunItemNotFound
	}
	return nil
}

// ListNokRunItems returns items marked nok for a run.
func (s *Store) ListNokRunItems(ctx context.Context, runID int64) ([]RunItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT ri.id, ri.run_id, ri.source_item_id, ri.section, ri.position, ri.label, ri.help_text, ri.required,
		       ri.status, ri.comment, ri.assigned_to, u.login, ri.updated_at
		FROM run_items ri
		LEFT JOIN users u ON u.id = ri.assigned_to
		WHERE ri.run_id = ? AND ri.status = ?
		ORDER BY ri.position
	`, runID, RunItemStatusNOK)
	if err != nil {
		return nil, fmt.Errorf("list nok run items: %w", err)
	}
	defer rows.Close()

	var nokItems []RunItem
	for rows.Next() {
		item, err := scanRunItemRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scan nok run item: %w", err)
		}
		nokItems = append(nokItems, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate nok run items: %w", err)
	}

	return nokItems, nil
}

// ListAssignedRunItems returns tasks assigned to a user with optional filters.
func (s *Store) ListAssignedRunItems(ctx context.Context, userID, projectID int64, status string) ([]AssignedRunItemSummary, error) {
	query := `
		SELECT ri.id, ri.run_id, ri.source_item_id, ri.section, ri.position, ri.label, ri.help_text, ri.required,
		       ri.status, ri.comment, ri.assigned_to, u.login, ri.updated_at,
		       cr.title, cr.project_id, p.name
		FROM run_items ri
		INNER JOIN checklist_runs cr ON cr.id = ri.run_id
		INNER JOIN projects p ON p.id = cr.project_id
		LEFT JOIN users u ON u.id = ri.assigned_to
		WHERE ri.assigned_to = ? AND cr.status != ?
	`
	args := []any{userID, RunStatusArchived}

	if projectID > 0 {
		query += " AND cr.project_id = ?"
		args = append(args, projectID)
	}
	if status != "" {
		query += " AND ri.status = ?"
		args = append(args, status)
	}
	query += " ORDER BY p.name, cr.title, ri.position"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list assigned run items: %w", err)
	}
	defer rows.Close()

	var tasks []AssignedRunItemSummary
	for rows.Next() {
		var summary AssignedRunItemSummary
		var required int
		var assignedLogin sql.NullString
		err := rows.Scan(
			&summary.ID, &summary.RunID, &summary.SourceItemID, &summary.Section, &summary.Position,
			&summary.Label, &summary.HelpText, &required, &summary.Status, &summary.Comment,
			&summary.AssignedTo, &assignedLogin, &summary.UpdatedAt,
			&summary.RunTitle, &summary.ProjectID, &summary.ProjectName,
		)
		if err != nil {
			return nil, fmt.Errorf("scan assigned run item: %w", err)
		}
		summary.Required = required == 1
		if assignedLogin.Valid {
			summary.AssignedLogin = assignedLogin.String
		}
		tasks = append(tasks, summary)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate assigned run items: %w", err)
	}

	return tasks, nil
}
