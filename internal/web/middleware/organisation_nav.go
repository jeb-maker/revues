package middleware

import (
	"context"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

// showOrganisationNav reports whether the Organisation top-level nav tab is shown.
// Solo org (one org, one member) hides it; global admins always see it (integrations).
func showOrganisationNav(ctx context.Context, st *store.Store, user *store.User, hd HeaderData) bool {
	if !CanManageOrgUsers(ctx, st, user) {
		return false
	}
	if auth.HasMinRole(user.Role, auth.RoleAdmin) {
		return true
	}
	if len(hd.UserOrganizations) > 1 {
		return true
	}
	org, ok := OrganizationFromContext(ctx)
	if !ok {
		return true
	}
	n, err := st.CountOrganizationMembers(ctx, org.ID)
	if err != nil || n > 1 {
		return true
	}
	allowed, err := st.CountAllowedEmails(ctx)
	if err != nil || allowed > 1 {
		return true
	}
	return false
}
