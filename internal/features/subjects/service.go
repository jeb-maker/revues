package subjects

import (
	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

// User is the authenticated account checked against subject access rules.
type User = store.User

// CanViewSubject reports whether the user may view a subject (v1 org-member flag).
// Prefer CanViewAccess with ResolveSubjectAccess for new code.
func CanViewSubject(user *User, orgMember bool) bool {
	if auth.HasMinRole(user.Role, auth.RoleAdmin) {
		return true
	}
	return orgMember
}

// CanViewAccess reports whether resolved subject access allows viewing.
func CanViewAccess(access store.SubjectAccess) bool {
	return access.Visible
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

// CanManageAccess reports whether the user may edit/archive a subject under resolved access.
// Org/global supervisors, subject leads, and v1 legacy org members (editor+) may manage.
func CanManageAccess(user *User, access store.SubjectAccess) bool {
	if !access.Visible {
		return false
	}
	if auth.HasMinRole(user.Role, auth.RoleAdmin) || access.HasSource(store.AccessSourceOrgAdmin) {
		return true
	}
	if !auth.HasMinRole(user.Role, auth.RoleEditor) {
		return false
	}
	return access.Role == store.SubjectRoleLead || access.HasSource(store.AccessSourceOrgMemberLegacy)
}

// CanLaunchRun reports whether the user may create or start a run on a subject.
func CanLaunchRun(user *User, orgMember bool) bool {
	if !CanViewSubject(user, orgMember) {
		return false
	}
	return auth.HasMinRole(user.Role, auth.RoleEditor)
}

// CanContributeAccess reports whether the user may launch/check on a subject.
func CanContributeAccess(user *User, access store.SubjectAccess) bool {
	if !access.Visible {
		return false
	}
	if !auth.HasMinRole(user.Role, auth.RoleEditor) {
		return false
	}
	return access.RoleAtLeast(store.SubjectRoleContributor)
}

// CanLeadAccess reports whether the user may assign/complete (lead-level) on a subject.
func CanLeadAccess(user *User, access store.SubjectAccess) bool {
	if !access.Visible {
		return false
	}
	if !auth.HasMinRole(user.Role, auth.RoleEditor) {
		return false
	}
	if access.IsSupervisor() || access.HasSource(store.AccessSourceOrgMemberLegacy) {
		return true
	}
	return access.Role == store.SubjectRoleLead
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

// DisplayRole returns a UI role label from resolved access.
func DisplayRole(access store.SubjectAccess) string {
	if access.Role != "" {
		return access.Role
	}
	if access.Visible {
		return LocalRoleLead
	}
	return ""
}
