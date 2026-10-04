package subjects

import (
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

func TestCanManageOrgUsers(t *testing.T) {
	editor := &User{Role: auth.RoleEditor}

	tests := []struct {
		name      string
		user      *User
		orgRole   string
		orgMember bool
		want      bool
	}{
		{"org owner", editor, store.OrgRoleOwner, true, true},
		{"org admin", editor, store.OrgRoleAdmin, true, true},
		{"org member", editor, store.OrgRoleMember, true, false},
		{"not member", editor, store.OrgRoleOwner, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanManageOrgUsers(tt.user, tt.orgRole, tt.orgMember); got != tt.want {
				t.Errorf("CanManageOrgUsers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanContributeAccess_OrgAdmin(t *testing.T) {
	user := &User{Role: auth.RoleEditor}
	orgAdminVis := store.SubjectAccess{Visible: true, Sources: []string{store.AccessSourceOrgAdmin}}
	viewerOnly := store.SubjectAccess{
		Visible: true,
		Role:    store.SubjectRoleViewer,
		Sources: []string{store.AccessSourceDirect},
	}

	if !CanContributeAccess(user, orgAdminVis) {
		t.Fatal("org admin should contribute without subject role")
	}
	if CanContributeAccess(user, viewerOnly) {
		t.Fatal("viewer must not contribute")
	}
}

func TestCanLeadAccess_NoOrgAdminBypass(t *testing.T) {
	editor := &User{Role: auth.RoleEditor}
	orgAdminOnly := store.SubjectAccess{Visible: true, Sources: []string{store.AccessSourceOrgAdmin}}
	orgAdminLead := store.SubjectAccess{
		Visible: true,
		Role:    store.SubjectRoleLead,
		Sources: []string{store.AccessSourceOrgAdmin, store.AccessSourceDirect},
	}

	if CanLeadAccess(editor, orgAdminOnly) {
		t.Fatal("org admin must not get implicit lead")
	}
	if !CanLeadAccess(editor, orgAdminLead) {
		t.Fatal("org admin who is also subject lead may assign/complete")
	}
}

func TestCanManageAccess(t *testing.T) {
	user := &User{Role: auth.RoleEditor}
	orgAdminVis := store.SubjectAccess{Visible: true, Sources: []string{store.AccessSourceOrgAdmin}}
	lead := store.SubjectAccess{Visible: true, Role: store.SubjectRoleLead, Sources: []string{store.AccessSourceDirect}}
	contributor := store.SubjectAccess{Visible: true, Role: store.SubjectRoleContributor, Sources: []string{store.AccessSourceDirect}}

	if !CanManageAccess(user, orgAdminVis) {
		t.Fatal("org admin may manage subjects")
	}
	if !CanManageAccess(user, lead) {
		t.Fatal("subject lead may manage subjects")
	}
	if CanManageAccess(user, contributor) {
		t.Fatal("contributor must not manage subject metadata")
	}
}

func TestCanSetSubjectVisibility(t *testing.T) {
	editor := &User{Role: auth.RoleEditor}
	lead := store.SubjectAccess{Visible: true, Role: store.SubjectRoleLead, Sources: []string{store.AccessSourceDirect}}
	empty := store.SubjectAccess{}

	if !CanSetSubjectVisibility(editor, store.OrgRoleAdmin, true, empty) {
		t.Fatal("org admin may set visibility on create")
	}
	if CanSetSubjectVisibility(editor, store.OrgRoleMember, true, empty) {
		t.Fatal("plain org member must not set visibility on create")
	}
	if !CanSetSubjectVisibility(editor, store.OrgRoleMember, true, lead) {
		t.Fatal("subject lead may set visibility on edit")
	}
}

func TestCanCreateSubject(t *testing.T) {
	user := &User{Role: auth.RoleEditor}
	if !CanCreateSubject(user, true) {
		t.Fatal("org member may create subject")
	}
	if CanCreateSubject(user, false) {
		t.Fatal("non-member must not create subject")
	}
}

func TestCanInviteSubjectMember(t *testing.T) {
	editor := &User{Role: auth.RoleEditor}

	allowMembers := store.OrgLeadPolicies{LeadsMayInviteMembers: true, LeadsMayInviteExternals: false}
	allowExternals := store.OrgLeadPolicies{LeadsMayInviteMembers: false, LeadsMayInviteExternals: true}
	denyAll := store.OrgLeadPolicies{}

	orgAdminOnly := store.SubjectAccess{Visible: true, Sources: []string{store.AccessSourceOrgAdmin}}
	leadDirect := store.SubjectAccess{
		Visible: true,
		Role:    store.SubjectRoleLead,
		Sources: []string{store.AccessSourceDirect},
	}
	contributor := store.SubjectAccess{
		Visible: true,
		Role:    store.SubjectRoleContributor,
		Sources: []string{store.AccessSourceDirect},
	}

	if !CanInviteSubjectMember(editor, orgAdminOnly, denyAll, false) {
		t.Fatal("org admin may invite externals regardless of policy")
	}
	if !CanInviteSubjectMember(editor, leadDirect, allowMembers, true) {
		t.Fatal("lead may invite org members when policy allows")
	}
	if !CanInviteSubjectMember(editor, leadDirect, allowExternals, false) {
		t.Fatal("lead may invite externals when policy allows")
	}
	if CanInviteSubjectMember(editor, leadDirect, denyAll, true) {
		t.Fatal("lead must respect deny policy for org members")
	}
	if CanInviteSubjectMember(editor, contributor, allowMembers, true) {
		t.Fatal("contributor must not invite")
	}
}
