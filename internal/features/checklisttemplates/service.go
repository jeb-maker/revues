package checklisttemplates

import (
	"github.com/jeb-maker/revues/internal/features/subjects"
	"github.com/jeb-maker/revues/internal/store"
)

// CanView reports whether the user may view checklist templates on a subject.
func CanView(user *store.User, orgMember bool) bool {
	return subjects.CanViewSubject(user, orgMember)
}

// CanManageGlobal reports whether the user may create, edit or archive global
// checklist templates (org admin or editor).
func CanManageGlobal(user *store.User) bool {
	return subjects.CanCreateSubject(user)
}
