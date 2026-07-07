package runs

import (
	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/features/projects"
	"github.com/jeb-maker/revues/internal/store"
)

// CanView reports whether the user may view a run on a project.
func CanView(user *store.User, isMember bool) bool {
	return projects.CanView(user, isMember)
}

// CanLaunch reports whether the user may create or start a run.
func CanLaunch(user *store.User, memberRole string) bool {
	if auth.HasMinRole(user.Role, auth.RoleAdmin) {
		return true
	}
	if !auth.HasMinRole(user.Role, auth.RoleEditor) {
		return false
	}
	return memberRole == projects.LocalRoleLead || memberRole == projects.LocalRoleContributor
}

// CanComplete reports whether the user may close a run (in_progress → done).
func CanComplete(user *store.User, memberRole string) bool {
	return projects.CanManage(user, memberRole)
}
