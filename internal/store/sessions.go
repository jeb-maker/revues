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

// CreateSession stores a new session for userID in organizationID.
// When organizationID is zero the column is stored as NULL (onboarding pending).
func (s *Store) CreateSession(ctx context.Context, userID, organizationID int64, tokenHash string) error {
	now := time.Now().UTC()
	expires := now.Add(sessionTTL).Format(time.RFC3339)

	var orgID any
	if organizationID > 0 {
		orgID = organizationID
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO sessions (token_hash, user_id, organization_id, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, tokenHash, userID, orgID, expires, now.Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}

	return nil
}

// SessionByTokenHash resolves an active session to user and organization ids.
// organizationID is zero when the session has no active organization yet.
func (s *Store) SessionByTokenHash(ctx context.Context, tokenHash string) (userID, organizationID int64, err error) {
	var orgID sql.NullInt64
	err = s.db.QueryRowContext(ctx, `
		SELECT user_id, organization_id FROM sessions
		WHERE token_hash = ? AND expires_at > ?
	`, tokenHash, time.Now().UTC().Format(time.RFC3339)).Scan(&userID, &orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, ErrSessionNotFound
	}
	if err != nil {
		return 0, 0, fmt.Errorf("lookup session: %w", err)
	}
	if orgID.Valid {
		organizationID = orgID.Int64
	}

	return userID, organizationID, nil
}

// UserIDByTokenHash resolves an active session to a user id.
func (s *Store) UserIDByTokenHash(ctx context.Context, tokenHash string) (int64, error) {
	userID, _, err := s.SessionByTokenHash(ctx, tokenHash)
	return userID, err
}

// ResolveSessionOrganizationID picks the organization stored in a new session.
// When preferredOrganizationID is negative the session is created without an org.
// When preferredOrganizationID is positive it is returned as-is; otherwise the
// sole membership is used, or the first membership when several exist. Users
// with no membership are added to the default organization as member (bootstrap).
func (s *Store) ResolveSessionOrganizationID(ctx context.Context, userID, preferredOrganizationID int64) (int64, error) {
	if preferredOrganizationID < 0 {
		return 0, nil
	}
	if preferredOrganizationID > 0 {
		return preferredOrganizationID, nil
	}

	memberships, err := s.ListUserOrganizations(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("list user organizations: %w", err)
	}
	switch len(memberships) {
	case 1:
		return memberships[0].Organization.ID, nil
	case 0:
		defaultOrg, err := s.OrganizationBySlug(ctx, "default")
		if err != nil {
			return 0, fmt.Errorf("default organization: %w", err)
		}
		if err := s.AddOrganizationMember(ctx, defaultOrg.ID, userID, OrgRoleMember); err != nil {
			return 0, fmt.Errorf("bootstrap default organization member: %w", err)
		}
		return defaultOrg.ID, nil
	default:
		return memberships[0].Organization.ID, nil
	}
}

// UpdateSessionOrganization sets the active organization on an existing session.
func (s *Store) UpdateSessionOrganization(ctx context.Context, tokenHash string, organizationID int64) error {
	if organizationID <= 0 {
		return fmt.Errorf("update session organization: invalid organization id")
	}

	res, err := s.db.ExecContext(ctx, `
		UPDATE sessions SET organization_id = ?
		WHERE token_hash = ? AND expires_at > ?
	`, organizationID, tokenHash, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("update session organization: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update session organization rows: %w", err)
	}
	if n == 0 {
		return ErrSessionNotFound
	}

	return nil
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
