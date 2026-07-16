package subjects

import (
	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

// User is the authenticated account checked against subject access rules.
type User = store.User

// CanViewSubject reports whether the user may view a subject in the active organization.
func CanViewSubject(user *User, orgMember bool) bool {
	if auth.HasMinRole(user.Role, auth.RoleAdmin) {
		return true
	}
	return orgMember
}

// CanManageSubject reports whether the user may create, edit or archive a subject.
func CanManageSubject(user *User, orgRole string, orgMember bool) bool {
	if auth.HasMinRole(user.Role, auth.RoleAdmin) {
		return true
	}
	if !orgMember {
		return false
	}
	if orgRole == store.OrgRoleOwner || orgRole == store.OrgRoleAdmin {
		return true
	}
	return auth.HasMinRole(user.Role, auth.RoleEditor)
}

// CanLaunchRun reports whether the user may create or start a run on a subject.
func CanLaunchRun(user *User, orgMember bool) bool {
	if !CanViewSubject(user, orgMember) {
		return false
	}
	return auth.HasMinRole(user.Role, auth.RoleEditor)
}

// CanCreateSubject reports whether the user may create a new subject.
func CanCreateSubject(user *User) bool {
	return auth.HasMinRole(user.Role, auth.RoleEditor)
}

// CanManageOrgUsers is true for global admin or org owner/admin.
func CanManageOrgUsers(user *User, orgRole string, orgMember bool) bool {
	if auth.HasMinRole(user.Role, auth.RoleAdmin) {
		return true
	}
	if !orgMember {
		return false
	}
	return orgRole == store.OrgRoleOwner || orgRole == store.OrgRoleAdmin
}

const (
	LocalRoleLead        = "lead"
	LocalRoleContributor = "contributor"
	LocalRoleViewer      = "viewer"
)
