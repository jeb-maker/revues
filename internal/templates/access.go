package templates

import (
	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/projects"
	"github.com/jeb-maker/revues/internal/store"
)

// CanView reports whether the user may view checklist templates on a project.
func CanView(user *store.User, isMember bool) bool {
	return projects.CanView(user, isMember)
}

// CanManage reports whether the user may create, edit or archive checklist templates.
func CanManage(user *store.User, memberRole string) bool {
	if auth.HasMinRole(user.Role, auth.RoleAdmin) {
		return true
	}
	return auth.HasMinRole(user.Role, auth.RoleEditor) && memberRole == projects.LocalRoleLead
}
