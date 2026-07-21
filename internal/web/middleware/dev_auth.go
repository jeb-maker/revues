package middleware

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

const (
	devAuthGitHubID     int64      = 1
	devAuthLogin                   = "devadmin"
	peerAddrContextKey  contextKey = 5
	devAuthUIContextKey contextKey = 6
)

// CapturePeerAddr stores the TCP peer IP before RealIP rewrites RemoteAddr.
// Must run before chi middleware.RealIP so DevAuth cannot be spoofed via X-Forwarded-For.
func CapturePeerAddr(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.RemoteAddr
		if h, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			host = h
		}
		ctx := context.WithValue(r.Context(), peerAddrContextKey, host)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// EnsureDevAuth auto-authenticates a demo admin when REVUES_DEV_AUTH is enabled
// (development only, loopback only). It creates a real session cookie so CSRF keeps working.
func EnsureDevAuth(st *store.Store, sessions *auth.SessionManager, enabled bool, email string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if enabled && IsLocalDevRequest(r) {
				r = r.WithContext(context.WithValue(r.Context(), devAuthUIContextKey, true))
			}
			if !enabled || isDevAuthExemptPath(r.URL.Path) || !IsLocalDevRequest(r) {
				next.ServeHTTP(w, r)
				return
			}
			if _, ok := UserFromContext(r.Context()); ok {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			user, orgID, err := ensureDevUser(ctx, st, email)
			if err != nil {
				slog.Error("dev auth bootstrap", "err", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			token, _, err := sessions.CreateLoginSession(ctx, user.ID, orgID)
			if err != nil {
				slog.Error("dev auth session", "err", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			sessions.SetSessionCookie(w, token)
			ctx = context.WithValue(ctx, userContextKey, user)
			ctx = context.WithValue(ctx, sessionTokenContextKey, token)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// DevAuthUIActive is true when local DevAuth UI (user switcher) may be shown.
func DevAuthUIActive(ctx context.Context) bool {
	active, _ := ctx.Value(devAuthUIContextKey).(bool)
	return active
}

// IsLocalDevRequest is true when the client is loopback and the Host is localhost/127.0.0.1.
func IsLocalDevRequest(r *http.Request) bool {
	if !isLoopbackHostHeader(r.Host) {
		return false
	}
	peer, _ := r.Context().Value(peerAddrContextKey).(string)
	if peer == "" {
		peer = r.RemoteAddr
		if h, _, err := net.SplitHostPort(peer); err == nil {
			peer = h
		}
	}
	return isLoopbackIP(peer)
}

func ensureDevUser(ctx context.Context, st *store.Store, email string) (*store.User, int64, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		email = "admin@example.com"
	}

	user, err := st.UpsertGitHubUser(ctx, devAuthGitHubID, devAuthLogin, email, "Admin démo", "", auth.RoleAdmin)
	if err != nil {
		return nil, 0, err
	}

	org, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		return nil, 0, err
	}
	if err = st.AddOrganizationMember(ctx, org.ID, user.ID, store.OrgRoleOwner); err != nil {
		return nil, 0, err
	}

	return user, org.ID, nil
}

func isDevAuthExemptPath(path string) bool {
	switch {
	case path == "/healthz", path == "/sw.js", path == "/login":
		return true
	case strings.HasPrefix(path, "/auth/dev/"):
		return true
	case strings.HasPrefix(path, "/static/"):
		return true
	default:
		return false
	}
}

func isLoopbackHostHeader(host string) bool {
	h := host
	if name, _, err := net.SplitHostPort(host); err == nil {
		h = name
	}
	h = strings.Trim(strings.ToLower(h), "[]")
	switch h {
	case "127.0.0.1", "localhost", "::1":
		return true
	default:
		return false
	}
}

func isLoopbackIP(ip string) bool {
	parsed := net.ParseIP(strings.Trim(ip, "[]"))
	return parsed != nil && parsed.IsLoopback()
}
