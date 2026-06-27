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

const (
	RunItemStatusPending = "pending"
	RunItemStatusOK      = "ok"
	RunItemStatusNOK     = "nok"
	RunItemStatusNA      = "na"
)

// RunItemByID loads an item scoped to a run.
func (s *Store) RunItemByID(ctx context.Context, runID, itemID int64) (*RunItem, error) {
	var item RunItem
	var required int
	err := s.db.QueryRowContext(ctx, `
		SELECT id, run_id, source_item_id, section, position, label, help_text, required,
		       status, comment, updated_at
		FROM run_items
		WHERE id = ? AND run_id = ?
	`, itemID, runID).Scan(
		&item.ID, &item.RunID, &item.SourceItemID, &item.Section, &item.Position,
		&item.Label, &item.HelpText, &required, &item.Status, &item.Comment, &item.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRunItemNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("run item by id: %w", err)
	}
	item.Required = required == 1
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

// ListNokRunItems returns items marked nok for a run.
func (s *Store) ListNokRunItems(ctx context.Context, runID int64) ([]RunItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, run_id, source_item_id, section, position, label, help_text, required,
		       status, comment, updated_at
		FROM run_items
		WHERE run_id = ? AND status = ?
		ORDER BY position
	`, runID, RunItemStatusNOK)
	if err != nil {
		return nil, fmt.Errorf("list nok run items: %w", err)
	}
	defer rows.Close()

	var nokItems []RunItem
	for rows.Next() {
		var item RunItem
		var required int
		if err := rows.Scan(
			&item.ID, &item.RunID, &item.SourceItemID, &item.Section, &item.Position,
			&item.Label, &item.HelpText, &required, &item.Status, &item.Comment, &item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan nok run item: %w", err)
		}
		item.Required = required == 1
		nokItems = append(nokItems, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate nok run items: %w", err)
	}

	return nokItems, nil
}
