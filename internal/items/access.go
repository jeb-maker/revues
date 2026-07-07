package items

import (
	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/features/projects"
	"github.com/jeb-maker/revues/internal/store"
)

// CanUpdate reports whether the user may change run item statuses.
func CanUpdate(user *store.User, memberRole string) bool {
	if auth.HasMinRole(user.Role, auth.RoleAdmin) {
		return true
	}
	if !auth.HasMinRole(user.Role, auth.RoleEditor) {
		return false
	}
	return memberRole == projects.LocalRoleLead || memberRole == projects.LocalRoleContributor
}

// CanLinkJira reports whether the user may link Jira issues to run items.
func CanLinkJira(user *store.User, memberRole string) bool {
	return CanUpdate(user, memberRole)
}

// CanAssign reports whether the user may assign run items to members.
func CanAssign(user *store.User, memberRole string) bool {
	if auth.HasMinRole(user.Role, auth.RoleAdmin) {
		return true
	}
	if !auth.HasMinRole(user.Role, auth.RoleEditor) {
		return false
	}
	return memberRole == projects.LocalRoleLead
}
