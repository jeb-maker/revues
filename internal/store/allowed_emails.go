package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrEmailNotAllowed is returned when an email is not on the whitelist.
var ErrEmailNotAllowed = errors.New("email not allowed")

// AllowedRole returns the role for email if whitelisted.
func (s *Store) AllowedRole(ctx context.Context, email string) (string, bool, error) {
	var role string
	err := s.db.QueryRowContext(ctx, `
		SELECT role FROM allowed_emails WHERE email = ?
	`, email).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("allowed role lookup: %w", err)
	}

	return role, true, nil
}

// InsertAllowedEmail adds an email to the whitelist.
func (s *Store) InsertAllowedEmail(ctx context.Context, email, role string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO allowed_emails (email, role, created_at)
		VALUES (?, ?, ?)
		ON CONFLICT(email) DO UPDATE SET role = excluded.role
	`, email, role, now)
	if err != nil {
		return fmt.Errorf("insert allowed email: %w", err)
	}

	return nil
}

// CountAllowedEmails returns whitelist size.
func (s *Store) CountAllowedEmails(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM allowed_emails`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count allowed emails: %w", err)
	}

	return count, nil
}

// ResolveLoginRole determines the role for a verified GitHub email at login.
func (s *Store) ResolveLoginRole(ctx context.Context, email, bootstrapAdmin string) (string, error) {
	if role, ok, err := s.AllowedRole(ctx, email); err != nil {
		return "", err
	} else if ok {
		return role, nil
	}

	count, err := s.CountAllowedEmails(ctx)
	if err != nil {
		return "", err
	}

	if count == 0 && bootstrapAdmin != "" && email == bootstrapAdmin {
		if err := s.InsertAllowedEmail(ctx, email, "admin"); err != nil {
			return "", err
		}
		return "admin", nil
	}

	return "", ErrEmailNotAllowed
}
