package auth

import (
	"fmt"
	"net/http"
	"time"
)

const guestCookieName = "revues_guest"

// EnsureGuestToken returns a guest cookie token and derived CSRF token for
// unauthenticated forms (login / register). Reuses an existing guest cookie when present.
func (m *SessionManager) EnsureGuestToken(w http.ResponseWriter, r *http.Request) (token, csrf string, err error) {
	if existing := GuestTokenFromRequest(r); existing != "" {
		return existing, CSRFToken(existing, m.SessionSecret), nil
	}

	raw, _, err := RandomToken(32)
	if err != nil {
		return "", "", fmt.Errorf("guest token: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     guestCookieName,
		Value:    raw,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   m.SecureCookies,
		MaxAge:   int((2 * time.Hour).Seconds()),
	})

	return raw, CSRFToken(raw, m.SessionSecret), nil
}

// ClearGuestCookie expires the guest CSRF cookie after a successful login.
func (m *SessionManager) ClearGuestCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     guestCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   m.SecureCookies,
		MaxAge:   -1,
	})
}

// GuestTokenFromRequest reads the guest CSRF material cookie.
func GuestTokenFromRequest(r *http.Request) string {
	c, err := r.Cookie(guestCookieName)
	if err != nil || c.Value == "" {
		return ""
	}
	return c.Value
}
