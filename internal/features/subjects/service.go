package subjects

import (
	"github.com/jeb-maker/revues/internal/store"
)

// User is the authenticated account checked against subject access rules.
type User = store.User

// CanViewSubject reports whether the user may view a subject via org membership alone.
// Prefer CanViewAccess with ResolveSubjectAccess for new code.
func CanViewSubject(_ *User, orgMember bool) bool {
	return orgMember
}

// CanViewAccess reports whether resolved subject access allows viewing.
func CanViewAccess(access store.SubjectAccess) bool {
	return access.Visible
}

// CanManageSubject reports whether the user may create, edit or archive a subject
// from org-level context (create path). Org owner/admin or any org member may create.
func CanManageSubject(_ *User, orgRole string, orgMember bool) bool {
	if !orgMember {
		return false
	}
	_ = orgRole
	return true
}

// CanManageAccess reports whether the user may edit/archive a subject under resolved access.
// Org owner/admin and subject leads may manage.
func CanManageAccess(_ *User, access store.SubjectAccess) bool {
	if !access.Visible {
		return false
	}
	if access.HasSource(store.AccessSourceOrgAdmin) {
		return true
	}
	return access.Role == store.SubjectRoleLead
}

// CanLaunchRun reports whether the user may create or start a run on a subject (org gate).
// Prefer CanContributeAccess with ResolveSubjectAccess for new code.
func CanLaunchRun(user *User, orgMember bool) bool {
	return CanViewSubject(user, orgMember)
}

// CanContributeAccess reports whether the user may launch/check on a subject.
// Org owner/admin may contribute without a subject role; otherwise contributor+.
func CanContributeAccess(_ *User, access store.SubjectAccess) bool {
	if !access.Visible {
		return false
	}
	if access.HasSource(store.AccessSourceOrgAdmin) {
		return true
	}
	return access.RoleAtLeast(store.SubjectRoleContributor)
}

// CanLeadAccess reports whether the user may assign/complete (lead-level) on a subject.
// Org owner/admin are not implicit leads for assign/complete; require subject lead.
func CanLeadAccess(_ *User, access store.SubjectAccess) bool {
	if !access.Visible {
		return false
	}
	return access.Role == store.SubjectRoleLead
}

// CanCreateSubject reports whether the user may create a new subject in the active org.
func CanCreateSubject(_ *User, orgMember bool) bool {
	return orgMember
}

// CanSetSubjectVisibility reports whether the user may set subjects.visibility.
// Org owner/admin may set it on create/edit; subject leads on edit.
func CanSetSubjectVisibility(_ *User, orgRole string, orgMember bool, access store.SubjectAccess) bool {
	if orgMember && (orgRole == store.OrgRoleOwner || orgRole == store.OrgRoleAdmin) {
		return true
	}
	return access.Visible && access.Role == store.SubjectRoleLead
}

// CanManageOrgUsers is true for org owner/admin.
func CanManageOrgUsers(_ *User, orgRole string, orgMember bool) bool {
	if !orgMember {
		return false
	}
	return orgRole == store.OrgRoleOwner || orgRole == store.OrgRoleAdmin
}

// PoliciesFromOrganization returns lead-delegation flags for the active org.
func PoliciesFromOrganization(org *store.Organization) store.OrgLeadPolicies {
	return org.LeadPolicies()
}

// CanInviteSubjectMember reports whether the user may add a direct subject member.
// inviteeIsOrgMember selects leads_may_invite_members vs leads_may_invite_externals.
// Org owner/admin always may.
func CanInviteSubjectMember(user *User, access store.SubjectAccess, policies store.OrgLeadPolicies, inviteeIsOrgMember bool) bool {
	if !access.Visible {
		return false
	}
	if access.HasSource(store.AccessSourceOrgAdmin) {
		return true
	}
	if !CanLeadAccess(user, access) {
		return false
	}
	if inviteeIsOrgMember {
		return policies.LeadsMayInviteMembers
	}
	return policies.LeadsMayInviteExternals
}

// CanManageSubjectMembers reports whether the user may manage direct subject members
// (invite form / remove). Org admin always; leads need at least one invite policy.
func CanManageSubjectMembers(user *User, access store.SubjectAccess, policies store.OrgLeadPolicies) bool {
	if !access.Visible {
		return false
	}
	if access.HasSource(store.AccessSourceOrgAdmin) {
		return true
	}
	if !CanLeadAccess(user, access) {
		return false
	}
	return policies.LeadsMayInviteMembers || policies.LeadsMayInviteExternals
}
