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

// ValidLocalRole reports whether role is a project-local role.
func ValidLocalRole(role string) bool {
	_, ok := localRoles[role]
	return ok
}

// CanCreate reports whether the user may create a new project.
func CanCreate(user *store.User) bool {
	return auth.HasMinRole(user.Role, auth.RoleEditor)
}

// CanView reports whether the user may view a project (member or admin).
func CanView(user *store.User, isMember bool) bool {
	if auth.HasMinRole(user.Role, auth.RoleAdmin) {
		return true
	}
	return isMember
}

// CanManage reports whether the user may update or archive a project.
func CanManage(user *store.User, memberRole string) bool {
	if auth.HasMinRole(user.Role, auth.RoleAdmin) {
		return true
	}
	return memberRole == LocalRoleLead
}

// CanManageMembers reports whether the user may change project membership.
func CanManageMembers(user *store.User, memberRole string) bool {
	return CanManage(user, memberRole)
}
