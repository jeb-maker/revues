package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
)

const orgContextKey contextKey = 2

// LoadActiveOrganization validates the session organization and injects it into context.
// HTML routes without an org redirect to org select/new.
// Authenticated API routes (except auth/bootstrap/health) soft-load the org when
// present; missing org yields JSON 403 org_required (SPA handles onboarding).
func LoadActiveOrganization(st *store.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok || isOrganizationExemptPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			if isAPIPath(r.URL.Path) {
				loadAPIOrganization(w, r, st, user, next)
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

func loadAPIOrganization(w http.ResponseWriter, r *http.Request, st *store.Store, user *store.User, next http.Handler) {
	token := SessionTokenFromContext(r)
	if token == "" {
		writeOrgRequired(w)
		return
	}

	_, orgID, err := st.SessionByTokenHash(r.Context(), auth.HashToken(token))
	if err != nil || orgID <= 0 {
		writeOrgRequired(w)
		return
	}

	org, err := st.OrganizationByID(r.Context(), orgID)
	if err != nil {
		writeOrgRequired(w)
		return
	}

	if _, member, err := st.OrganizationMemberRole(r.Context(), orgID, user.ID); err != nil || !member {
		writeOrgRequired(w)
		return
	}

	ctx := orgctx.WithOrganizationID(r.Context(), org.ID)
	ctx = context.WithValue(ctx, orgContextKey, org)
	next.ServeHTTP(w, r.WithContext(ctx))
}

func writeOrgRequired(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    "org_required",
			"message": "Organisation active requise.",
		},
	})
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
		// Non-v1 API probes stay exempt.
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
