package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrSessionNotFound is returned when a session token is unknown or expired.
var ErrSessionNotFound = errors.New("session not found")

const sessionTTL = 7 * 24 * time.Hour

// CreateSession stores a new session for userID and returns the raw token.
func (s *Store) CreateSession(ctx context.Context, userID int64, tokenHash string) error {
	now := time.Now().UTC()
	expires := now.Add(sessionTTL).Format(time.RFC3339)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO sessions (token_hash, user_id, expires_at, created_at)
		VALUES (?, ?, ?, ?)
	`, tokenHash, userID, expires, now.Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}

	return nil
}

// UserIDByTokenHash resolves an active session to a user id.
func (s *Store) UserIDByTokenHash(ctx context.Context, tokenHash string) (int64, error) {
	var userID int64
	err := s.db.QueryRowContext(ctx, `
		SELECT user_id FROM sessions
		WHERE token_hash = ? AND expires_at > ?
	`, tokenHash, time.Now().UTC().Format(time.RFC3339)).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrSessionNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("lookup session: %w", err)
	}

	return userID, nil
}

// DeleteSession removes a session by token hash.
func (s *Store) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash = ?`, tokenHash)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	return nil
}

// DeleteUserSessions removes all sessions for a user (rotation on login).
func (s *Store) DeleteUserSessions(ctx context.Context, userID int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("delete user sessions: %w", err)
	}

	return nil
}
