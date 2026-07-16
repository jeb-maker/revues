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
// Global admin may always manage. Org admin and subject leads require editor+ global role.
// Org admin reader cannot edit subjects (visibility ≠ write).
func CanManageAccess(user *User, access store.SubjectAccess) bool {
	if !access.Visible {
		return false
	}
	if auth.HasMinRole(user.Role, auth.RoleAdmin) {
		return true
	}
	if !auth.HasMinRole(user.Role, auth.RoleEditor) {
		return false
	}
	if access.HasSource(store.AccessSourceOrgAdmin) {
		return true
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
// Org admin with editor+ may contribute (no subject role required); reader cannot.
func CanContributeAccess(user *User, access store.SubjectAccess) bool {
	if !access.Visible {
		return false
	}
	if !auth.HasMinRole(user.Role, auth.RoleEditor) {
		return false
	}
	if auth.HasMinRole(user.Role, auth.RoleAdmin) || access.HasSource(store.AccessSourceOrgAdmin) {
		return true
	}
	return access.RoleAtLeast(store.SubjectRoleContributor)
}

// CanLeadAccess reports whether the user may assign/complete (lead-level) on a subject.
// Org admin visibility is not an implicit lead: assign/complete require subject lead
// (or legacy ungated path). Global admin keeps full lead capability.
func CanLeadAccess(user *User, access store.SubjectAccess) bool {
	if !access.Visible {
		return false
	}
	if !auth.HasMinRole(user.Role, auth.RoleEditor) {
		return false
	}
	if auth.HasMinRole(user.Role, auth.RoleAdmin) {
		return true
	}
	if access.HasSource(store.AccessSourceOrgMemberLegacy) {
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

// LeadsMayAssignTeams reports whether subject leads may grant team access.
// Stubbed true until Issue 9 (org policies) lands.
func LeadsMayAssignTeams() bool {
	return true
}

// CanAssignSubjectTeams reports whether the user may add/remove teams on a subject.
// Org owner/admin and global admin always may; subject leads require LeadsMayAssignTeams.
// Org admin visibility alone is enough (no subject lead required — unlike CanLeadAccess).
func CanAssignSubjectTeams(user *User, access store.SubjectAccess) bool {
	if !access.Visible {
		return false
	}
	if auth.HasMinRole(user.Role, auth.RoleAdmin) {
		return true
	}
	if access.HasSource(store.AccessSourceOrgAdmin) {
		return true
	}
	if !LeadsMayAssignTeams() {
		return false
	}
	return CanLeadAccess(user, access)
}

const (
	LocalRoleLead        = "lead"
	LocalRoleContributor = "contributor"
	LocalRoleViewer      = "viewer"
)

// DisplayRole returns a UI role label from resolved access.
// Supervisors without a subject role are not shown as lead.
func DisplayRole(access store.SubjectAccess) string {
	if access.Role != "" {
		return access.Role
	}
	if access.HasSource(store.AccessSourceOrgMemberLegacy) {
		return LocalRoleContributor
	}
	return ""
}
