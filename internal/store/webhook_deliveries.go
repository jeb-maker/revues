package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Webhook delivery queue states.
const (
	WebhookDeliveryPending = "pending"
	WebhookDeliveryDone    = "done"
	WebhookDeliveryPoison  = "poison"
)

// WebhookDelivery is one durable outbound webhook attempt queue row.
type WebhookDelivery struct {
	ID            int64
	EventID       string
	EventType     string
	URL           string
	Payload       []byte
	StatusCode    sql.NullInt64
	Success       bool
	Attempts      int
	NextAttemptAt string
	ExpiresAt     string
	State         string
	LastError     string
	CreatedAt     string
}

// EnqueueWebhookDelivery inserts a pending delivery for durable retry.
func (s *Store) EnqueueWebhookDelivery(ctx context.Context, eventID, eventType, url string, payload []byte, nextAttemptAt, expiresAt time.Time) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO webhook_deliveries (
			event_id, event_type, url, status_code, success, created_at,
			payload, attempts, next_attempt_at, expires_at, state, last_error
		) VALUES (?, ?, ?, NULL, 0, ?, ?, 0, ?, ?, ?, NULL)
	`, eventID, eventType, url, now, payload, nextAttemptAt.UTC().Format(time.RFC3339), expiresAt.UTC().Format(time.RFC3339), WebhookDeliveryPending)
	if err != nil {
		return 0, fmt.Errorf("enqueue webhook delivery: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("enqueue webhook delivery id: %w", err)
	}
	return id, nil
}

// ListDueWebhookDeliveries returns pending deliveries ready to attempt.
func (s *Store) ListDueWebhookDeliveries(ctx context.Context, now time.Time, limit int) ([]WebhookDelivery, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, event_id, event_type, url, payload, status_code, success,
		       attempts, COALESCE(next_attempt_at, ''), COALESCE(expires_at, ''),
		       state, COALESCE(last_error, ''), created_at
		FROM webhook_deliveries
		WHERE state = ? AND next_attempt_at IS NOT NULL AND next_attempt_at <= ?
		ORDER BY next_attempt_at ASC
		LIMIT ?
	`, WebhookDeliveryPending, now.UTC().Format(time.RFC3339), limit)
	if err != nil {
		return nil, fmt.Errorf("list due webhook deliveries: %w", err)
	}
	defer rows.Close()

	var out []WebhookDelivery
	for rows.Next() {
		var d WebhookDelivery
		var successInt int
		var payload []byte
		if err := rows.Scan(
			&d.ID, &d.EventID, &d.EventType, &d.URL, &payload, &d.StatusCode, &successInt,
			&d.Attempts, &d.NextAttemptAt, &d.ExpiresAt, &d.State, &d.LastError, &d.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan webhook delivery: %w", err)
		}
		d.Payload = payload
		d.Success = successInt == 1
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate webhook deliveries: %w", err)
	}
	return out, nil
}

// UpdateWebhookDeliveryAttempt persists the result of one delivery attempt.
func (s *Store) UpdateWebhookDeliveryAttempt(ctx context.Context, id int64, statusCode int, success bool, attempts int, nextAttemptAt *time.Time, state, lastError string) error {
	var code any
	if statusCode > 0 {
		code = statusCode
	}
	successInt := 0
	if success {
		successInt = 1
	}
	var next any
	if nextAttemptAt != nil {
		next = nextAttemptAt.UTC().Format(time.RFC3339)
	}
	var errMsg any
	if lastError != "" {
		errMsg = lastError
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE webhook_deliveries
		SET status_code = ?, success = ?, attempts = ?, next_attempt_at = ?, state = ?, last_error = ?
		WHERE id = ?
	`, code, successInt, attempts, next, state, errMsg, id)
	if err != nil {
		return fmt.Errorf("update webhook delivery %d: %w", id, err)
	}
	return nil
}

// InsertWebhookDelivery records a one-shot delivery log row (legacy / immediate success path).
// Prefer EnqueueWebhookDelivery + UpdateWebhookDeliveryAttempt for durable retries.
func (s *Store) InsertWebhookDelivery(ctx context.Context, eventID, eventType, url string, statusCode int, success bool) error {
	now := time.Now().UTC().Format(time.RFC3339)
	var code any
	if statusCode > 0 {
		code = statusCode
	}
	successInt := 0
	if success {
		successInt = 1
	}
	state := WebhookDeliveryDone
	if !success {
		state = WebhookDeliveryPoison
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO webhook_deliveries (
			event_id, event_type, url, status_code, success, created_at,
			payload, attempts, next_attempt_at, expires_at, state, last_error
		) VALUES (?, ?, ?, ?, ?, ?, NULL, 1, NULL, NULL, ?, NULL)
	`, eventID, eventType, url, code, successInt, now, state)
	if err != nil {
		return fmt.Errorf("insert webhook delivery: %w", err)
	}
	return nil
}
