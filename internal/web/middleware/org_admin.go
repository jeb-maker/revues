package middleware

import (
	"context"

	"github.com/jeb-maker/revues/internal/store"
)

// OrgRoleLookup looks up a user's role in an organization.
type OrgRoleLookup interface {
	OrganizationMemberRole(ctx context.Context, organizationID, userID int64) (string, bool, error)
}

// CanManageOrgUsers reports whether user can manage org settings (members, invitations, integrations).
func CanManageOrgUsers(ctx context.Context, st OrgRoleLookup, user *store.User) bool {
	org, ok := OrganizationFromContext(ctx)
	if !ok {
		return false
	}
	role, member, err := st.OrganizationMemberRole(ctx, org.ID, user.ID)
	if err != nil || !member {
		return false
	}
	_ = user
	return role == store.OrgRoleOwner || role == store.OrgRoleAdmin
}
