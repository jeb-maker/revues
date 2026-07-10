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
func LoadActiveOrganization(st *store.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok || isOrganizationExemptPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			token, err := auth.SessionTokenFromRequest(r)
			if err != nil {
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

// OrganizationFromContext returns the active organization for the request, if any.
func OrganizationFromContext(ctx context.Context) (*store.Organization, bool) {
	org, ok := ctx.Value(orgContextKey).(*store.Organization)
	return org, ok
}

func isOrganizationExemptPath(path string) bool {
	switch {
	case path == "/org/select", path == "/org/new":
		return true
	case strings.HasPrefix(path, "/org/invitations/"):
		return true
	case strings.HasPrefix(path, "/login"),
		strings.HasPrefix(path, "/auth/"),
		path == "/logout",
		path == "/healthz",
		strings.HasPrefix(path, "/static/"):
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
