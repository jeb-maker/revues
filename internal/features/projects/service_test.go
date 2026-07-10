package projects

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

func TestCanAddProjectMember(t *testing.T) {
	admin := &User{Role: auth.RoleAdmin}
	editor := &User{Role: auth.RoleEditor}

	tests := []struct {
		name       string
		user       *User
		memberRole string
		orgRole    string
		want       bool
	}{
		{"global admin", admin, LocalRoleViewer, store.OrgRoleMember, true},
		{"project lead", editor, LocalRoleLead, store.OrgRoleMember, true},
		{"org owner viewer", editor, LocalRoleViewer, store.OrgRoleOwner, true},
		{"org admin contributor", editor, LocalRoleContributor, store.OrgRoleAdmin, true},
		{"org member contributor", editor, LocalRoleContributor, store.OrgRoleMember, false},
		{"org member viewer", editor, LocalRoleViewer, store.OrgRoleMember, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanAddProjectMember(tt.user, tt.memberRole, tt.orgRole); got != tt.want {
				t.Errorf("CanAddProjectMember() = %v, want %v", got, tt.want)
			}
		})
	}
}
