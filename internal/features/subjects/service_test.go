package subjects

import (
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

func TestCanManageOrgUsers(t *testing.T) {
	admin := &User{Role: auth.RoleAdmin}
	editor := &User{Role: auth.RoleEditor}

	tests := []struct {
		name      string
		user      *User
		orgRole   string
		orgMember bool
		want      bool
	}{
		{"global admin", admin, store.OrgRoleMember, false, true},
		{"org owner", editor, store.OrgRoleOwner, true, true},
		{"org admin", editor, store.OrgRoleAdmin, true, true},
		{"org member", editor, store.OrgRoleMember, true, false},
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
	editor := &User{Role: auth.RoleEditor}
	reader := &User{Role: auth.RoleReader}
	orgAdminVis := store.SubjectAccess{Visible: true, Sources: []string{store.AccessSourceOrgAdmin}}

	if !CanContributeAccess(editor, orgAdminVis) {
		t.Fatal("org admin editor should contribute without subject role")
	}
	if CanContributeAccess(reader, orgAdminVis) {
		t.Fatal("org admin reader must not contribute")
	}
}

func TestCanLeadAccess_NoOrgAdminBypass(t *testing.T) {
	admin := &User{Role: auth.RoleAdmin}
	editor := &User{Role: auth.RoleEditor}
	orgAdminOnly := store.SubjectAccess{Visible: true, Sources: []string{store.AccessSourceOrgAdmin}}
	orgAdminLead := store.SubjectAccess{
		Visible: true,
		Role:    store.SubjectRoleLead,
		Sources: []string{store.AccessSourceOrgAdmin, store.AccessSourceDirect},
	}
	legacy := store.SubjectAccess{
		Visible: true,
		Role:    store.SubjectRoleContributor,
		Sources: []string{store.AccessSourceOrgMemberLegacy},
	}

	if CanLeadAccess(editor, orgAdminOnly) {
		t.Fatal("org admin must not get implicit lead")
	}
	if !CanLeadAccess(editor, orgAdminLead) {
		t.Fatal("org admin who is also subject lead may assign/complete")
	}
	if !CanLeadAccess(editor, legacy) {
		t.Fatal("legacy ungated path still allows lead actions for editor")
	}
	if !CanLeadAccess(admin, orgAdminOnly) {
		t.Fatal("global admin should keep lead capability")
	}
}

func TestCanManageAccess_OrgAdminReaderDenied(t *testing.T) {
	editor := &User{Role: auth.RoleEditor}
	reader := &User{Role: auth.RoleReader}
	admin := &User{Role: auth.RoleAdmin}
	orgAdminVis := store.SubjectAccess{Visible: true, Sources: []string{store.AccessSourceOrgAdmin}}

	if !CanManageAccess(editor, orgAdminVis) {
		t.Fatal("org admin editor may manage subjects")
	}
	if CanManageAccess(reader, orgAdminVis) {
		t.Fatal("org admin reader must not manage subjects")
	}
	if !CanManageAccess(admin, orgAdminVis) {
		t.Fatal("global admin may manage subjects")
	}
}

func TestCanSetSubjectVisibility(t *testing.T) {
	admin := &User{Role: auth.RoleAdmin}
	editor := &User{Role: auth.RoleEditor}
	reader := &User{Role: auth.RoleReader}
	lead := store.SubjectAccess{Visible: true, Role: store.SubjectRoleLead, Sources: []string{store.AccessSourceDirect}}
	legacy := store.SubjectAccess{Visible: true, Role: store.SubjectRoleContributor, Sources: []string{store.AccessSourceOrgMemberLegacy}}
	empty := store.SubjectAccess{}

	if !CanSetSubjectVisibility(admin, store.OrgRoleMember, true, empty) {
		t.Fatal("global admin may set visibility on create")
	}
	if !CanSetSubjectVisibility(editor, store.OrgRoleAdmin, true, empty) {
		t.Fatal("org admin may set visibility on create")
	}
	if CanSetSubjectVisibility(editor, store.OrgRoleMember, true, empty) {
		t.Fatal("plain org member editor must not set visibility on create")
	}
	if !CanSetSubjectVisibility(editor, store.OrgRoleMember, true, lead) {
		t.Fatal("subject lead may set visibility on edit")
	}
	if CanSetSubjectVisibility(editor, store.OrgRoleMember, true, legacy) {
		t.Fatal("legacy ungated must not set visibility")
	}
	if CanSetSubjectVisibility(reader, store.OrgRoleAdmin, true, empty) {
		t.Fatal("reader must not set visibility")
	}
}

func TestCanAssignSubjectTeams(t *testing.T) {
	admin := &User{Role: auth.RoleAdmin}
	editor := &User{Role: auth.RoleEditor}
	reader := &User{Role: auth.RoleReader}

	allow := store.OrgLeadPolicies{LeadsMayAssignTeams: true, LeadsMayInviteMembers: true}
	deny := store.OrgLeadPolicies{LeadsMayAssignTeams: false, LeadsMayInviteMembers: true}

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
	hidden := store.SubjectAccess{}

	if !CanAssignSubjectTeams(admin, orgAdminOnly, deny) {
		t.Fatal("global admin may assign teams even when policy denies leads")
	}
	if !CanAssignSubjectTeams(editor, orgAdminOnly, deny) {
		t.Fatal("org admin may assign teams without subject lead")
	}
	if !CanAssignSubjectTeams(reader, orgAdminOnly, deny) {
		t.Fatal("org admin reader may assign teams (supervision, not métier write)")
	}
	if !CanAssignSubjectTeams(editor, leadDirect, allow) {
		t.Fatal("subject lead may assign teams when policy allows")
	}
	if CanAssignSubjectTeams(editor, leadDirect, deny) {
		t.Fatal("subject lead must not assign teams when policy denies")
	}
	if CanAssignSubjectTeams(editor, contributor, allow) {
		t.Fatal("contributor must not assign teams")
	}
	if CanAssignSubjectTeams(editor, hidden, allow) {
		t.Fatal("invisible subject: no team assign")
	}
}

func TestCanInviteSubjectMember(t *testing.T) {
	admin := &User{Role: auth.RoleAdmin}
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

	if !CanInviteSubjectMember(admin, orgAdminOnly, denyAll, true) {
		t.Fatal("global admin may invite org members")
	}
	if !CanInviteSubjectMember(editor, orgAdminOnly, denyAll, false) {
		t.Fatal("org admin may invite externals regardless of policy")
	}
	if !CanInviteSubjectMember(editor, leadDirect, allowMembers, true) {
		t.Fatal("lead may invite org members when policy allows")
	}
	if CanInviteSubjectMember(editor, leadDirect, allowMembers, false) {
		t.Fatal("lead must not invite externals when only members policy is on")
	}
	if !CanInviteSubjectMember(editor, leadDirect, allowExternals, false) {
		t.Fatal("lead may invite externals when policy allows")
	}
	if CanInviteSubjectMember(editor, leadDirect, allowExternals, true) {
		t.Fatal("lead must not invite org members when only externals policy is on")
	}
	if CanInviteSubjectMember(editor, leadDirect, denyAll, true) {
		t.Fatal("lead must not invite when both policies deny")
	}
	if CanInviteSubjectMember(editor, contributor, allowMembers, true) {
		t.Fatal("contributor must not invite")
	}
	if !CanManageSubjectMembers(editor, leadDirect, allowMembers) {
		t.Fatal("lead may manage members when invite members allowed")
	}
	if !CanManageSubjectMembers(editor, leadDirect, allowExternals) {
		t.Fatal("lead may manage members when invite externals allowed")
	}
	if CanManageSubjectMembers(editor, leadDirect, denyAll) {
		t.Fatal("lead must not manage members when both invite policies deny")
	}
}
