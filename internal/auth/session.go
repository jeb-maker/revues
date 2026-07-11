package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const (
	sessionCookieName = "revues_session"
	oauthCookieName   = "revues_oauth"
)

// SessionManager handles browser session cookies and DB persistence.
type SessionManager struct {
	Store         SessionStore
	SessionSecret string
	SecureCookies bool
}

// CreateLoginSession rotates sessions and returns raw token + CSRF token.
// When organizationID is zero the store resolves the active organization.
func (m *SessionManager) CreateLoginSession(ctx context.Context, userID, organizationID int64) (string, string, error) {
	orgID, err := m.Store.ResolveSessionOrganizationID(ctx, userID, organizationID)
	if err != nil {
		return "", "", err
	}

	if err := m.Store.DeleteUserSessions(ctx, userID); err != nil {
		return "", "", err
	}

	raw, hash, err := RandomToken(32)
	if err != nil {
		return "", "", err
	}

	if err := m.Store.CreateSession(ctx, userID, orgID, hash); err != nil {
		return "", "", err
	}

	return raw, CSRFToken(raw, m.SessionSecret), nil
}

// ClearSession removes the session from the database.
func (m *SessionManager) ClearSession(ctx context.Context, sessionToken string) error {
	if sessionToken == "" {
		return nil
	}

	return m.Store.DeleteSession(ctx, HashToken(sessionToken))
}

// SetSessionCookie writes the session cookie on w.
func (m *SessionManager) SetSessionCookie(w http.ResponseWriter, sessionToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   m.SecureCookies,
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
	})
}

// ClearSessionCookie expires the session cookie.
func (m *SessionManager) ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   m.SecureCookies,
		MaxAge:   -1,
	})
}

// SessionTokenFromRequest reads the session cookie.
func SessionTokenFromRequest(r *http.Request) (string, error) {
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return "", fmt.Errorf("session cookie: %w", err)
	}

	return c.Value, nil
}

// SetOAuthCookie stores signed OAuth state for PKCE.
func (m *SessionManager) SetOAuthCookie(w http.ResponseWriter, payload, signature string) {
	value := payload + "|" + signature
	http.SetCookie(w, &http.Cookie{
		Name:     oauthCookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   m.SecureCookies,
		MaxAge:   600,
	})
}

// ClearOAuthCookie removes the pending OAuth cookie.
func (m *SessionManager) ClearOAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     oauthCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   m.SecureCookies,
		MaxAge:   -1,
	})
}

// ParseOAuthCookie validates and returns state and verifier.
func (m *SessionManager) ParseOAuthCookie(r *http.Request) (state, verifier string, err error) {
	c, err := r.Cookie(oauthCookieName)
	if err != nil {
		return "", "", fmt.Errorf("oauth cookie: %w", err)
	}

	parts := splitOAuthCookie(c.Value)
	if len(parts) != 3 {
		return "", "", fmt.Errorf("invalid oauth cookie format")
	}

	payload := parts[0] + "|" + parts[1]
	if !VerifyOAuthPayload(payload, m.SessionSecret, parts[2]) {
		return "", "", fmt.Errorf("invalid oauth cookie signature")
	}

	return parts[0], parts[1], nil
}

// BuildOAuthCookiePayload returns payload and signature for state and verifier.
func (m *SessionManager) BuildOAuthCookiePayload(state, verifier string) (payload, signature string) {
	payload = state + "|" + verifier
	return payload, SignOAuthPayload(payload, m.SessionSecret)
}

func splitOAuthCookie(value string) []string {
	lastPipe := -1
	for i := len(value) - 1; i >= 0; i-- {
		if value[i] == '|' {
			lastPipe = i
			break
		}
	}
	if lastPipe <= 0 {
		return nil
	}

	midPipe := -1
	for i := lastPipe - 1; i >= 0; i-- {
		if value[i] == '|' {
			midPipe = i
			break
		}
	}
	if midPipe < 0 {
		return nil
	}

	return []string{
		value[:midPipe],
		value[midPipe+1 : lastPipe],
		value[lastPipe+1:],
	}
}
