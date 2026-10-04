package runs_test

import (
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/features/runs"
	"github.com/jeb-maker/revues/internal/store"
)

func TestCanLaunch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		role      string
		orgMember bool
		want      bool
	}{
		{"org member", auth.RoleEditor, true, true},
		{"outsider", auth.RoleEditor, false, false},
		{"member any technical role", auth.RoleReader, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			user := &store.User{Role: tt.role}
			if got := runs.CanLaunch(user, tt.orgMember); got != tt.want {
				t.Errorf("CanLaunch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanComplete(t *testing.T) {
	t.Parallel()

	editor := &store.User{Role: auth.RoleEditor}

	if !runs.CanComplete(editor, store.OrgRoleMember, true) {
		t.Fatal("org member should complete runs (coarse gate)")
	}
	if !runs.CanComplete(editor, store.OrgRoleOwner, true) {
		t.Fatal("org owner should complete runs")
	}
	if runs.CanComplete(editor, store.OrgRoleMember, false) {
		t.Fatal("non-member must not complete runs")
	}
}
