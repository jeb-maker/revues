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

// ErrInvalidAssignee is returned when assignee cannot see the subject (RBAC visibility).
var ErrInvalidAssignee = errors.New("invalid assignee")

// ErrRunItemConflict is returned when the caller's expected updated_at no longer matches the row
// (optimistic lock: the item was modified in between).
var ErrRunItemConflict = errors.New("run item modified concurrently")

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

// UpdateRunItemStatus changes status and comment on an in-progress run item (no optimistic lock).
func (s *Store) UpdateRunItemStatus(ctx context.Context, runID, itemID, userID int64, status, comment string) error {
	return s.UpdateRunItemStatusChecked(ctx, runID, itemID, userID, status, comment, "")
}

// UpdateRunItemStatusChecked is UpdateRunItemStatus with an optimistic lock: when expectedUpdatedAt
// is non-empty the UPDATE only applies if the row's updated_at still equals it, otherwise
// ErrRunItemConflict is returned and nothing is written. Empty expectedUpdatedAt = unconditional.
func (s *Store) UpdateRunItemStatusChecked(
	ctx context.Context,
	runID, itemID, userID int64,
	status, comment, expectedUpdatedAt string,
) error {
	run, err := s.RunByID(ctx, runID)
	if err != nil {
		return err
	}
	if run.Status != RunStatusInProgress {
		return ErrRunNotEditable
	}

	now := time.Now().UTC().Format(time.RFC3339)
	comment = strings.TrimSpace(comment)

	return withSQLiteBusyRetry(ctx, func() error {
		return s.updateRunItemStatusOnce(ctx, runID, itemID, userID, status, comment, now, expectedUpdatedAt)
	})
}

func (s *Store) updateRunItemStatusOnce(
	ctx context.Context,
	runID, itemID, userID int64,
	status, comment, now, expectedUpdatedAt string,
) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	oldStatus, err := loadRunItemStatusTx(ctx, tx, runID, itemID)
	if err != nil {
		return err
	}

	var checkedBy sql.NullInt64
	var checkedAt sql.NullString
	if status != RunItemStatusPending {
		checkedBy = sql.NullInt64{Int64: userID, Valid: true}
		checkedAt = sql.NullString{String: now, Valid: true}
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE run_items
		SET status = ?, comment = ?, checked_by = ?, checked_at = ?, updated_at = ?
		WHERE id = ? AND run_id = ? AND (? = '' OR updated_at = ?)
	`, status, comment, checkedBy, checkedAt, now, itemID, runID, expectedUpdatedAt, expectedUpdatedAt)
	if err != nil {
		return fmt.Errorf("update run item: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update run item rows: %w", err)
	}
	if n == 0 {
		return runItemNoRowsError(ctx, tx, runID, itemID, expectedUpdatedAt)
	}

	if oldStatus != status {
		if err := insertRunItemEventTx(ctx, tx, itemID, userID, oldStatus, status, comment, now); err != nil {
			return err
		}
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("commit update run item: %w", commitErr)
	}
	return nil
}

// AssignRunItem sets or clears assignee on an in-progress run item (no optimistic lock).
func (s *Store) AssignRunItem(ctx context.Context, runID, itemID int64, assigneeID *int64) error {
	return s.AssignRunItemChecked(ctx, runID, itemID, assigneeID, "")
}

// AssignRunItemChecked is AssignRunItem with an optimistic lock on updated_at
// (same contract as UpdateRunItemStatusChecked).
func (s *Store) AssignRunItemChecked(
	ctx context.Context,
	runID, itemID int64,
	assigneeID *int64,
	expectedUpdatedAt string,
) error {
	run, err := s.RunByID(ctx, runID)
	if err != nil {
		return err
	}
	if run.Status != RunStatusInProgress {
		return ErrRunNotEditable
	}

	var assignedTo sql.NullInt64
	if assigneeID != nil {
		assignee, userErr := s.UserByID(ctx, *assigneeID)
		if errors.Is(userErr, ErrUserNotFound) {
			return ErrInvalidAssignee
		}
		if userErr != nil {
			return fmt.Errorf("load assignee: %w", userErr)
		}
		access, accessErr := s.ResolveSubjectAccess(ctx, *assigneeID, run.SubjectID, assignee.Role)
		if accessErr != nil {
			return fmt.Errorf("resolve assignee access: %w", accessErr)
		}
		if !access.Visible {
			return ErrInvalidAssignee
		}
		assignedTo = sql.NullInt64{Int64: *assigneeID, Valid: true}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	return withSQLiteBusyRetry(ctx, func() error {
		return s.assignRunItemOnce(ctx, runID, itemID, assignedTo, now, expectedUpdatedAt)
	})
}

func (s *Store) assignRunItemOnce(
	ctx context.Context,
	runID, itemID int64,
	assignedTo sql.NullInt64,
	now, expectedUpdatedAt string,
) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	res, err := tx.ExecContext(ctx, `
		UPDATE run_items SET assigned_to = ?, updated_at = ?
		WHERE id = ? AND run_id = ? AND (? = '' OR updated_at = ?)
	`, assignedTo, now, itemID, runID, expectedUpdatedAt, expectedUpdatedAt)
	if err != nil {
		return fmt.Errorf("assign run item: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("assign run item rows: %w", err)
	}
	if n == 0 {
		return runItemNoRowsError(ctx, tx, runID, itemID, expectedUpdatedAt)
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("commit assign run item: %w", commitErr)
	}
	return nil
}

// runItemNoRowsError disambiguates a zero-row UPDATE: missing item vs stale expectedUpdatedAt.
// Read in the same transaction as the UPDATE so the verdict matches what the guard saw.
func runItemNoRowsError(ctx context.Context, tx *sql.Tx, runID, itemID int64, expectedUpdatedAt string) error {
	if expectedUpdatedAt == "" {
		return ErrRunItemNotFound
	}
	var current string
	err := tx.QueryRowContext(ctx, `
		SELECT updated_at FROM run_items WHERE id = ? AND run_id = ?
	`, itemID, runID).Scan(&current)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrRunItemNotFound
	}
	if err != nil {
		return fmt.Errorf("load run item updated_at: %w", err)
	}
	return ErrRunItemConflict
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
// Scoped to the active organization (orgctx) and non-archived subjects/runs.
// IDOR: only rows where assigned_to = userID. Private/gated subjects are hidden
// unless the assignee still has ResolveSubjectAccess visibility (or is global admin).
func (s *Store) ListAssignedRunItems(ctx context.Context, userID int64, status, query string) ([]AssignedRunItemSummary, error) {
	orgID, err := organizationIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	user, err := s.UserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	globalAdmin := user.Role == "admin"

	sqlQuery := `
		SELECT ri.id, ri.run_id, ri.source_item_id, ri.section, ri.position, ri.label, ri.help_text, ri.required,
		       ri.status, ri.comment, ri.assigned_to, u.login, ri.updated_at,
		       t.name, p.name, cr.created_at, cr.id, cr.subject_id
		FROM run_items ri
		INNER JOIN checklist_runs cr ON cr.id = ri.run_id
		INNER JOIN subjects p ON p.id = cr.subject_id
		INNER JOIN template_versions tv ON tv.id = cr.template_version_id
		INNER JOIN checklist_templates t ON t.id = tv.template_id
		LEFT JOIN users u ON u.id = ri.assigned_to
	`
	args := []any{}
	if !globalAdmin {
		sqlQuery += `
		INNER JOIN organization_members om ON om.organization_id = p.organization_id AND om.user_id = ?
		`
		args = append(args, userID)
	}
	sqlQuery += `
		WHERE ri.assigned_to = ? AND cr.status != ? AND p.organization_id = ? AND p.archived_at IS NULL
	`
	args = append(args, userID, RunStatusArchived, orgID)
	if !globalAdmin {
		sqlQuery += subjectVisibleToOrgMemberSQL("p")
		args = append(args, userID, userID, orgID)
	}

	if status != "" {
		sqlQuery += " AND ri.status = ?"
		args = append(args, status)
	}
	for _, term := range searchTerms(query) {
		pattern := likeContainsPattern(term)
		sqlQuery += ` AND (
			p.name LIKE ? ESCAPE '\'
			OR t.name LIKE ? ESCAPE '\'
			OR ri.label LIKE ? ESCAPE '\'
			OR ri.section LIKE ? ESCAPE '\'
		)`
		args = append(args, pattern, pattern, pattern, pattern)
	}
	sqlQuery += " ORDER BY p.name, t.name, cr.created_at, ri.position"

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("list assigned run items: %w", err)
	}
	defer rows.Close()

	var tasks []AssignedRunItemSummary
	for rows.Next() {
		var summary AssignedRunItemSummary
		var required int
		var assignedLogin sql.NullString
		var templateName, subjectName, createdAt string
		var runID int64
		err := rows.Scan(
			&summary.ID, &summary.RunID, &summary.SourceItemID, &summary.Section, &summary.Position,
			&summary.Label, &summary.HelpText, &required, &summary.Status, &summary.Comment,
			&summary.AssignedTo, &assignedLogin, &summary.UpdatedAt,
			&templateName, &subjectName, &createdAt, &runID, &summary.SubjectID,
		)
		if err != nil {
			return nil, fmt.Errorf("scan assigned run item: %w", err)
		}
		summary.RunTitle = RunDisplayLabel(templateName, subjectName, createdAt, runID)
		summary.SubjectName = subjectName
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
