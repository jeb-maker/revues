package middleware

import (
	"context"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

// OrgRoleLookup looks up a user's role in an organization.
type OrgRoleLookup interface {
	OrganizationMemberRole(ctx context.Context, organizationID, userID int64) (string, bool, error)
}

// CanManageOrgUsers reports whether user can manage org settings (whitelist, integrations).
func CanManageOrgUsers(ctx context.Context, st OrgRoleLookup, user *store.User) bool {
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
