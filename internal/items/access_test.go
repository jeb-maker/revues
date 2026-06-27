package items_test

import (
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/items"
	"github.com/jeb-maker/revues/internal/projects"
	"github.com/jeb-maker/revues/internal/store"
)

func TestCanAssign(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		role       string
		memberRole string
		want       bool
	}{
		{"admin", auth.RoleAdmin, "", true},
		{"editor lead", auth.RoleEditor, projects.LocalRoleLead, true},
		{"editor contributor", auth.RoleEditor, projects.LocalRoleContributor, false},
		{"editor viewer", auth.RoleEditor, projects.LocalRoleViewer, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			user := &store.User{Role: tt.role}
			if got := items.CanAssign(user, tt.memberRole); got != tt.want {
				t.Errorf("CanAssign() = %v, want %v", got, tt.want)
			}
		})
	}
}
