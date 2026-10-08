package atlassian

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/crypto"
	"github.com/jeb-maker/revues/internal/store"
)

// ErrUserOAuthRequired is returned when create/link needs a connected Atlassian account.
var ErrUserOAuthRequired = errors.New("atlassian user oauth required")

// ErrEncryptionNotConfigured is returned when REVUES_ENCRYPTION_KEY is missing.
var ErrEncryptionNotConfigured = crypto.ErrEncryptionNotConfigured

const refreshSkew = 60 * time.Second

// TokenStore persists encrypted Atlassian OAuth tokens.
type TokenStore interface {
	GetAtlassianOAuthTokensByUserID(ctx context.Context, userID int64) (*store.AtlassianOAuthTokens, error)
	UpsertAtlassianOAuthTokens(ctx context.Context, row store.AtlassianOAuthTokens) error
	DeleteAtlassianOAuthTokensByUserID(ctx context.Context, userID int64) error
}

// TokenService loads and refreshes per-user Atlassian OAuth tokens.
type TokenService struct {
	Store         TokenStore
	EncryptionKey []byte
	OAuth         *auth.AtlassianOAuth
}

// AccessCredentials are returned by EnsureAccessToken.
type AccessCredentials struct {
	AccessToken  string
	CloudID      string
	SiteURL      string
	AccountEmail string
}

// EnsureAccessToken returns a valid access token, refreshing when it expires within 60s.
func (s *TokenService) EnsureAccessToken(ctx context.Context, userID int64) (AccessCredentials, error) {
	if len(s.EncryptionKey) != crypto.KeySize {
		return AccessCredentials{}, ErrEncryptionNotConfigured
	}

	row, err := s.Store.GetAtlassianOAuthTokensByUserID(ctx, userID)
	if errors.Is(err, store.ErrAtlassianOAuthNotFound) {
		return AccessCredentials{}, ErrUserOAuthRequired
	}
	if err != nil {
		return AccessCredentials{}, fmt.Errorf("load atlassian oauth: %w", err)
	}

	access, err := crypto.Decrypt(s.EncryptionKey, row.AccessTokenEncrypted)
	if err != nil {
		return AccessCredentials{}, fmt.Errorf("decrypt access token: %w", err)
	}
	refresh, err := crypto.Decrypt(s.EncryptionKey, row.RefreshTokenEncrypted)
	if err != nil {
		return AccessCredentials{}, fmt.Errorf("decrypt refresh token: %w", err)
	}

	expiresAt, err := time.Parse(time.RFC3339, row.ExpiresAt)
	if err != nil {
		expiresAt = time.Time{}
	}

	creds := AccessCredentials{
		AccessToken:  string(access),
		CloudID:      row.CloudID,
		SiteURL:      row.SiteURL,
		AccountEmail: row.AccountEmail,
	}

	if time.Until(expiresAt) > refreshSkew {
		return creds, nil
	}

	if s.OAuth == nil {
		return AccessCredentials{}, ErrUserOAuthRequired
	}

	token, err := s.OAuth.RefreshAccessToken(ctx, string(refresh))
	if err != nil {
		return AccessCredentials{}, fmt.Errorf("%w: refresh failed: %w", ErrUserOAuthRequired, err)
	}

	newRefresh := token.RefreshToken
	if newRefresh == "" {
		newRefresh = string(refresh)
	}
	accessEnc, err := crypto.Encrypt(s.EncryptionKey, []byte(token.AccessToken))
	if err != nil {
		return AccessCredentials{}, fmt.Errorf("encrypt access token: %w", err)
	}
	refreshEnc, err := crypto.Encrypt(s.EncryptionKey, []byte(newRefresh))
	if err != nil {
		return AccessCredentials{}, fmt.Errorf("encrypt refresh token: %w", err)
	}

	expiresIn := token.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 3600
	}
	scopes := token.Scope
	if scopes == "" {
		scopes = row.Scopes
	}
	updated := store.AtlassianOAuthTokens{
		UserID:                userID,
		CloudID:               row.CloudID,
		SiteURL:               row.SiteURL,
		AccountEmail:          row.AccountEmail,
		AccessTokenEncrypted:  accessEnc,
		RefreshTokenEncrypted: refreshEnc,
		ExpiresAt:             time.Now().UTC().Add(time.Duration(expiresIn) * time.Second).Format(time.RFC3339),
		Scopes:                scopes,
	}
	if err := s.Store.UpsertAtlassianOAuthTokens(ctx, updated); err != nil {
		return AccessCredentials{}, fmt.Errorf("save refreshed tokens: %w", err)
	}

	creds.AccessToken = token.AccessToken
	return creds, nil
}

// Connected reports whether the user has stored Atlassian OAuth tokens.
func (s *TokenService) Connected(ctx context.Context, userID int64) (bool, *store.AtlassianOAuthTokens, error) {
	row, err := s.Store.GetAtlassianOAuthTokensByUserID(ctx, userID)
	if errors.Is(err, store.ErrAtlassianOAuthNotFound) {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, err
	}
	return true, row, nil
}

// Disconnect removes stored Atlassian OAuth tokens for the user.
func (s *TokenService) Disconnect(ctx context.Context, userID int64) error {
	err := s.Store.DeleteAtlassianOAuthTokensByUserID(ctx, userID)
	if errors.Is(err, store.ErrAtlassianOAuthNotFound) {
		return nil
	}
	return err
}

// NormalizeSiteURL trims trailing slashes.
func NormalizeSiteURL(raw string) string {
	return strings.TrimRight(strings.TrimSpace(raw), "/")
}
