package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// RunItemEvent is an audit entry for a status change on a run item.
type RunItemEvent struct {
	ID        int64
	RunItemID int64
	UserID    sql.NullInt64
	UserLogin string
	OldStatus sql.NullString
	NewStatus string
	Comment   string
	CreatedAt string
}

// ListRunItemEvents returns audit history for a run item ordered by recency.
func (s *Store) ListRunItemEvents(ctx context.Context, runItemID int64) ([]RunItemEvent, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT e.id, e.run_item_id, e.user_id, u.login, e.old_status, e.new_status, e.comment, e.created_at
		FROM run_item_events e
		LEFT JOIN users u ON u.id = e.user_id
		WHERE e.run_item_id = ?
		ORDER BY e.created_at DESC
	`, runItemID)
	if err != nil {
		return nil, fmt.Errorf("list run item events: %w", err)
	}
	defer rows.Close()

	var events []RunItemEvent
	for rows.Next() {
		var event RunItemEvent
		var userLogin sql.NullString
		if err := rows.Scan(
			&event.ID, &event.RunItemID, &event.UserID, &userLogin,
			&event.OldStatus, &event.NewStatus, &event.Comment, &event.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan run item event: %w", err)
		}
		if userLogin.Valid {
			event.UserLogin = userLogin.String
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate run item events: %w", err)
	}

	return events, nil
}

func insertRunItemEventTx(ctx context.Context, tx *sql.Tx, runItemID, userID int64, oldStatus, newStatus, comment, now string) error {
	var old sql.NullString
	if oldStatus != "" {
		old = sql.NullString{String: oldStatus, Valid: true}
	}
	var user sql.NullInt64
	if userID > 0 {
		user = sql.NullInt64{Int64: userID, Valid: true}
	}

	_, err := tx.ExecContext(ctx, `
		INSERT INTO run_item_events (run_item_id, user_id, old_status, new_status, comment, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, runItemID, user, old, newStatus, comment, now)
	if err != nil {
		return fmt.Errorf("insert run item event: %w", err)
	}
	return nil
}

func loadRunItemStatusTx(ctx context.Context, tx *sql.Tx, runID, itemID int64) (string, error) {
	var status string
	err := tx.QueryRowContext(ctx, `
		SELECT status FROM run_items WHERE id = ? AND run_id = ?
	`, itemID, runID).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrRunItemNotFound
	}
	if err != nil {
		return "", fmt.Errorf("load run item status: %w", err)
	}
	return status, nil
}
