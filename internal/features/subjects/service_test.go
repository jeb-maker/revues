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

func TestCanAssignSubjectTeams(t *testing.T) {
	admin := &User{Role: auth.RoleAdmin}
	editor := &User{Role: auth.RoleEditor}
	reader := &User{Role: auth.RoleReader}

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

	if !CanAssignSubjectTeams(admin, orgAdminOnly) {
		t.Fatal("global admin may assign teams")
	}
	if !CanAssignSubjectTeams(editor, orgAdminOnly) {
		t.Fatal("org admin may assign teams without subject lead")
	}
	if !CanAssignSubjectTeams(reader, orgAdminOnly) {
		t.Fatal("org admin reader may assign teams (supervision, not métier write)")
	}
	if !CanAssignSubjectTeams(editor, leadDirect) {
		t.Fatal("subject lead may assign teams when policy allows")
	}
	if CanAssignSubjectTeams(editor, contributor) {
		t.Fatal("contributor must not assign teams")
	}
	if CanAssignSubjectTeams(editor, hidden) {
		t.Fatal("invisible subject: no team assign")
	}
}
