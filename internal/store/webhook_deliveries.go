package store

import (
	"context"
	"fmt"
	"time"
)

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
	_, err := s.db.ExecContext(ctx, `INSERT INTO webhook_deliveries (event_id, event_type, url, status_code, success, created_at) VALUES (?, ?, ?, ?, ?, ?)`, eventID, eventType, url, code, successInt, now)
	if err != nil {
		return fmt.Errorf("insert webhook delivery: %w", err)
	}
	return nil
}
