package projects

import (
	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

const (
	LocalRoleLead        = "lead"
	LocalRoleContributor = "contributor"
	LocalRoleViewer      = "viewer"
)

var localRoles = map[string]struct{}{
	LocalRoleLead:        {},
	LocalRoleContributor: {},
	LocalRoleViewer:      {},
}

func ValidLocalRole(role string) bool {
	_, ok := localRoles[role]
	return ok
}

func CanCreate(user *User) bool {
	return auth.HasMinRole(user.Role, auth.RoleEditor)
}

func CanView(user *User, isMember bool) bool {
	if auth.HasMinRole(user.Role, auth.RoleAdmin) {
		return true
	}
	return isMember
}

func CanManage(user *User, memberRole string) bool {
	if auth.HasMinRole(user.Role, auth.RoleAdmin) {
		return true
	}
	return memberRole == LocalRoleLead
}

func CanManageMembers(user *User, memberRole string) bool {
	return CanManage(user, memberRole)
}

func isOrgPrivilegedRole(orgRole string) bool {
	return orgRole == store.OrgRoleOwner || orgRole == store.OrgRoleAdmin
}

// CanAddProjectMember is true for global admin, project lead, or org owner/admin.
func CanAddProjectMember(user *User, memberRole, orgRole string) bool {
	if auth.HasMinRole(user.Role, auth.RoleAdmin) {
		return true
	}
	if memberRole == LocalRoleLead {
		return true
	}
	return isOrgPrivilegedRole(orgRole)
}

// CanInviteExternalToOrg is true when the invitee is not yet in the organization.
func CanInviteExternalToOrg(user *User, memberRole, orgRole string) bool {
	return CanAddProjectMember(user, memberRole, orgRole)
}

func CanLaunch(user *User, memberRole string) bool {
	if auth.HasMinRole(user.Role, auth.RoleAdmin) {
		return true
	}
	if !auth.HasMinRole(user.Role, auth.RoleEditor) {
		return false
	}
	return memberRole == LocalRoleLead || memberRole == LocalRoleContributor
}
