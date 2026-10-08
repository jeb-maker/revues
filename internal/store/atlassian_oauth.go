package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrAtlassianOAuthNotFound is returned when the user has no Atlassian OAuth tokens.
var ErrAtlassianOAuthNotFound = errors.New("atlassian oauth tokens not found")

// AtlassianOAuthTokens holds encrypted Atlassian 3LO tokens for one user.
type AtlassianOAuthTokens struct {
	UserID                int64
	CloudID               string
	SiteURL               string
	AccountEmail          string
	AccessTokenEncrypted  []byte
	RefreshTokenEncrypted []byte
	ExpiresAt             string
	Scopes                string
	CreatedAt             string
	UpdatedAt             string
}

// UpsertAtlassianOAuthTokens inserts or replaces Atlassian OAuth tokens for a user.
func (s *Store) UpsertAtlassianOAuthTokens(ctx context.Context, row AtlassianOAuthTokens) error {
	if row.UserID <= 0 {
		return fmt.Errorf("upsert atlassian oauth: invalid user_id")
	}
	if len(row.AccessTokenEncrypted) == 0 || len(row.RefreshTokenEncrypted) == 0 {
		return fmt.Errorf("upsert atlassian oauth: tokens required")
	}
	if row.ExpiresAt == "" {
		return fmt.Errorf("upsert atlassian oauth: expires_at required")
	}

	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO atlassian_oauth_tokens (
			user_id, cloud_id, site_url, account_email,
			access_token_encrypted, refresh_token_encrypted,
			expires_at, scopes, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			cloud_id = excluded.cloud_id,
			site_url = excluded.site_url,
			account_email = excluded.account_email,
			access_token_encrypted = excluded.access_token_encrypted,
			refresh_token_encrypted = excluded.refresh_token_encrypted,
			expires_at = excluded.expires_at,
			scopes = excluded.scopes,
			updated_at = excluded.updated_at
	`, row.UserID, row.CloudID, row.SiteURL, row.AccountEmail,
		row.AccessTokenEncrypted, row.RefreshTokenEncrypted,
		row.ExpiresAt, row.Scopes, now, now)
	if err != nil {
		return fmt.Errorf("upsert atlassian oauth: %w", err)
	}
	return nil
}

// GetAtlassianOAuthTokensByUserID returns stored Atlassian tokens for the user.
func (s *Store) GetAtlassianOAuthTokensByUserID(ctx context.Context, userID int64) (*AtlassianOAuthTokens, error) {
	var row AtlassianOAuthTokens
	err := s.db.QueryRowContext(ctx, `
		SELECT user_id, cloud_id, site_url, account_email,
			access_token_encrypted, refresh_token_encrypted,
			expires_at, scopes, created_at, updated_at
		FROM atlassian_oauth_tokens
		WHERE user_id = ?
	`, userID).Scan(
		&row.UserID,
		&row.CloudID,
		&row.SiteURL,
		&row.AccountEmail,
		&row.AccessTokenEncrypted,
		&row.RefreshTokenEncrypted,
		&row.ExpiresAt,
		&row.Scopes,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrAtlassianOAuthNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get atlassian oauth: %w", err)
	}
	return &row, nil
}

// DeleteAtlassianOAuthTokensByUserID removes Atlassian OAuth tokens for the user.
func (s *Store) DeleteAtlassianOAuthTokensByUserID(ctx context.Context, userID int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM atlassian_oauth_tokens WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("delete atlassian oauth: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete atlassian oauth rows: %w", err)
	}
	if n == 0 {
		return ErrAtlassianOAuthNotFound
	}
	return nil
}
