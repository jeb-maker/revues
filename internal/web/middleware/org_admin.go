package middleware

import (
	"context"
	"net/http"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

// RequireOrgAdmin ensures the user is owner or admin of the active organization.
// Global admins bypass the org role check but still require a valid active organization.
func RequireOrgAdmin(st *store.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			if !CanManageOrgUsers(r.Context(), st, user) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// CanManageOrgUsers reports whether user can manage the org email whitelist.
func CanManageOrgUsers(ctx context.Context, st *store.Store, user *store.User) bool {
	if auth.HasMinRole(user.Role, auth.RoleAdmin) {
		return true
	}
	org, ok := OrganizationFromContext(ctx)
	if !ok {
		return false
	}
	role, member, err := st.OrganizationMemberRole(ctx, org.ID, user.ID)
	if err != nil || !member {
		return false
	}
	return role == store.OrgRoleOwner || role == store.OrgRoleAdmin
}
