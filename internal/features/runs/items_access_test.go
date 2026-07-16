package runs_test

import (
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/features/runs"
	"github.com/jeb-maker/revues/internal/store"
)

func TestCanAssign(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			user := &store.User{Role: tt.role}
			if got := runs.CanAssign(user, tt.orgRole, tt.orgMember); got != tt.want {
				t.Errorf("CanAssign() = %v, want %v", got, tt.want)
			}
		})
	}
}
