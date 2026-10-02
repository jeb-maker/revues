package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
)

const orgContextKey contextKey = 2

// LoadActiveOrganization validates the session organization and injects it into context.
// HTML routes without an org redirect to org select/new.
// Authenticated API routes soft-load the org when present; handlers call requireOrg when needed.
func LoadActiveOrganization(st *store.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok || isOrganizationExemptPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			if isAPIPath(r.URL.Path) {
				loadAPIOrganization(r, st, user, next, w)
				return
			}

			token := SessionTokenFromContext(r)
			if token == "" {
				redirectPendingOrganization(w, r, st, user)
				return
			}

			_, orgID, err := st.SessionByTokenHash(r.Context(), auth.HashToken(token))
			if err != nil || orgID <= 0 {
				redirectPendingOrganization(w, r, st, user)
				return
			}

			org, err := st.OrganizationByID(r.Context(), orgID)
			if err != nil {
				redirectPendingOrganization(w, r, st, user)
				return
			}

			if _, member, err := st.OrganizationMemberRole(r.Context(), orgID, user.ID); err != nil || !member {
				redirectPendingOrganization(w, r, st, user)
				return
			}

			ctx := orgctx.WithOrganizationID(r.Context(), org.ID)
			ctx = context.WithValue(ctx, orgContextKey, org)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func loadAPIOrganization(r *http.Request, st *store.Store, user *store.User, next http.Handler, w http.ResponseWriter) {
	token := SessionTokenFromContext(r)
	if token == "" {
		next.ServeHTTP(w, r)
		return
	}

	_, orgID, err := st.SessionByTokenHash(r.Context(), auth.HashToken(token))
	if err != nil || orgID <= 0 {
		next.ServeHTTP(w, r)
		return
	}

	org, err := st.OrganizationByID(r.Context(), orgID)
	if err != nil {
		next.ServeHTTP(w, r)
		return
	}

	if _, member, err := st.OrganizationMemberRole(r.Context(), orgID, user.ID); err != nil || !member {
		next.ServeHTTP(w, r)
		return
	}

	ctx := orgctx.WithOrganizationID(r.Context(), org.ID)
	ctx = context.WithValue(ctx, orgContextKey, org)
	next.ServeHTTP(w, r.WithContext(ctx))
}

// OrganizationFromContext returns the active organization for the request, if any.
func OrganizationFromContext(ctx context.Context) (*store.Organization, bool) {
	org, ok := ctx.Value(orgContextKey).(*store.Organization)
	return org, ok
}

func isAPIPath(path string) bool {
	return strings.HasPrefix(path, "/api/")
}

func isOrganizationExemptPath(path string) bool {
	switch {
	case path == "/org/select", path == "/org/new":
		return true
	case strings.HasPrefix(path, "/org/invitations/"):
		return true
	case strings.HasPrefix(path, "/login"),
		path == "/register",
		strings.HasPrefix(path, "/auth/"),
		path == "/logout",
		path == "/healthz",
		strings.HasPrefix(path, "/static/"):
		return true
	case path == "/api/v1/health",
		path == "/api/v1/bootstrap",
		path == "/api/v1/me",
		strings.HasPrefix(path, "/api/v1/auth/"):
		return true
	case strings.HasPrefix(path, "/api/") && !strings.HasPrefix(path, "/api/v1/"):
		return true
	default:
		return false
	}
}

func redirectPendingOrganization(w http.ResponseWriter, r *http.Request, st *store.Store, user *store.User) {
	count, err := st.CountUserOrganizations(r.Context(), user.ID)
	if err != nil {
		http.Redirect(w, r, "/org/select", http.StatusFound)
		return
	}
	if count == 0 {
		http.Redirect(w, r, "/org/new", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/org/select", http.StatusFound)
}
