package store

import (
	"context"
	"fmt"
	"time"
)

// Email delivery queue states.
const (
	EmailDeliveryPending = "pending"
	EmailDeliveryDone    = "done"
	EmailDeliveryPoison  = "poison"
)

// EmailDelivery is one durable outbound email queue row.
type EmailDelivery struct {
	ID             int64
	OrganizationID int64
	ToAddress      string
	Subject        string
	Body           string
	Attempts       int
	NextAttemptAt  string
	ExpiresAt      string
	State          string
	LastError      string
	CreatedAt      string
}

// EnqueueEmailDelivery inserts a pending notification email for durable retry.
// organizationID may be 0 when the caller has no org context (still drainable).
func (s *Store) EnqueueEmailDelivery(ctx context.Context, organizationID int64, to, subject, body string, nextAttemptAt, expiresAt time.Time) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	var org any
	if organizationID > 0 {
		org = organizationID
	}
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO email_deliveries (
			organization_id, to_address, subject, body, attempts, next_attempt_at, expires_at, state, last_error, created_at
		) VALUES (?, ?, ?, ?, 0, ?, ?, ?, NULL, ?)
	`, org, to, subject, body, nextAttemptAt.UTC().Format(time.RFC3339), expiresAt.UTC().Format(time.RFC3339), EmailDeliveryPending, now)
	if err != nil {
		return 0, fmt.Errorf("enqueue email delivery: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("enqueue email delivery id: %w", err)
	}
	return id, nil
}

// ListDueEmailDeliveries returns pending emails ready to attempt.
func (s *Store) ListDueEmailDeliveries(ctx context.Context, now time.Time, limit int) ([]EmailDelivery, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, COALESCE(organization_id, 0), to_address, subject, body,
		       attempts, COALESCE(next_attempt_at, ''), COALESCE(expires_at, ''),
		       state, COALESCE(last_error, ''), created_at
		FROM email_deliveries
		WHERE state = ? AND next_attempt_at IS NOT NULL AND next_attempt_at <= ?
		ORDER BY next_attempt_at ASC
		LIMIT ?
	`, EmailDeliveryPending, now.UTC().Format(time.RFC3339), limit)
	if err != nil {
		return nil, fmt.Errorf("list due email deliveries: %w", err)
	}
	defer rows.Close()

	var out []EmailDelivery
	for rows.Next() {
		var d EmailDelivery
		if err := rows.Scan(
			&d.ID, &d.OrganizationID, &d.ToAddress, &d.Subject, &d.Body,
			&d.Attempts, &d.NextAttemptAt, &d.ExpiresAt, &d.State, &d.LastError, &d.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan email delivery: %w", err)
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate email deliveries: %w", err)
	}
	return out, nil
}

// UpdateEmailDeliveryAttempt persists the result of one send attempt.
func (s *Store) UpdateEmailDeliveryAttempt(ctx context.Context, id int64, attempts int, nextAttemptAt *time.Time, state, lastError string) error {
	var next any
	if nextAttemptAt != nil {
		next = nextAttemptAt.UTC().Format(time.RFC3339)
	}
	var errMsg any
	if lastError != "" {
		errMsg = lastError
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE email_deliveries
		SET attempts = ?, next_attempt_at = ?, state = ?, last_error = ?
		WHERE id = ?
	`, attempts, next, state, errMsg, id)
	if err != nil {
		return fmt.Errorf("update email delivery %d: %w", id, err)
	}
	return nil
}
