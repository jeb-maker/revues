package templates

import (
	"github.com/jeb-maker/revues/internal/features/subjects"
	"github.com/jeb-maker/revues/internal/store"
)

// CanView reports whether the user may view checklist templates on a subject.
func CanView(user *store.User, orgMember bool) bool {
	return subjects.CanViewSubject(user, orgMember)
}

// CanManage reports whether the user may create, edit or archive checklist templates on a subject.
func CanManage(user *store.User, orgRole string, orgMember bool) bool {
	return subjects.CanManageSubject(user, orgRole, orgMember)
}
