package checklisttemplates

import (
	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/features/projects"
	"github.com/jeb-maker/revues/internal/store"
)

// CanView reports whether the user may view checklist templates on a project.
func CanView(user *store.User, isMember bool) bool {
	return projects.CanView(user, isMember)
}

// CanManageGlobal reports whether the user may create, edit or archive global
// checklist templates (org admin or editor).
func CanManageGlobal(user *store.User) bool {
	return auth.HasMinRole(user.Role, auth.RoleEditor)
}
