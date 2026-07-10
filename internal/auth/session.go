package auth

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const (
	sessionCookieName = "revues_session"
	oauthCookieName   = "revues_oauth"
	lastOrgCookieName = "revues_last_org"
)

// SessionOrgPending marks a login session waiting for organization selection.
const SessionOrgPending int64 = -1

// SessionManager handles browser session cookies and DB persistence.
type SessionManager struct {
	Store         SessionStore
	SessionSecret string
	SecureCookies bool
}

// CreateLoginSession rotates sessions and returns raw token + CSRF token.
// Pass SessionOrgPending to create a session without an active organization.
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

// SetActiveOrganization updates the organization on the current session.
func (m *SessionManager) SetActiveOrganization(ctx context.Context, sessionToken string, organizationID int64) error {
	if sessionToken == "" {
		return fmt.Errorf("session token: empty")
	}
	return m.Store.UpdateSessionOrganization(ctx, HashToken(sessionToken), organizationID)
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

// LastOrgIDFromRequest reads the remembered organization id cookie.
func LastOrgIDFromRequest(r *http.Request) int64 {
	c, err := r.Cookie(lastOrgCookieName)
	if err != nil {
		return 0
	}
	id, err := strconv.ParseInt(c.Value, 10, 64)
	if err != nil || id <= 0 {
		return 0
	}
	return id
}

// SetLastOrgCookie remembers the user's last selected organization.
func SetLastOrgCookie(w http.ResponseWriter, orgID int64, secure bool) {
	if orgID <= 0 {
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     lastOrgCookieName,
		Value:    strconv.FormatInt(orgID, 10),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		MaxAge:   int((365 * 24 * time.Hour).Seconds()),
	})
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
