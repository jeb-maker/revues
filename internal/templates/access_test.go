package templates_test

import (
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/templates"
)

func TestCanManage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		role      string
		orgRole   string
		orgMember bool
		want      bool
	}{
		{"admin", auth.RoleAdmin, store.OrgRoleMember, false, true},
		{"editor org owner", auth.RoleEditor, store.OrgRoleOwner, true, true},
		{"editor org member", auth.RoleEditor, store.OrgRoleMember, true, true},
		{"reader org member", auth.RoleReader, store.OrgRoleMember, true, false},
		{"editor outsider", auth.RoleEditor, store.OrgRoleMember, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			user := &store.User{Role: tt.role}
			if got := templates.CanManage(user, tt.orgRole, tt.orgMember); got != tt.want {
				t.Errorf("CanManage() = %v, want %v", got, tt.want)
			}
		})
	}
}
