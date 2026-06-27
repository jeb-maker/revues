package templates_test

import (
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/projects"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/templates"
)

func TestCanManage(t *testing.T) {
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
		{"reader lead", auth.RoleReader, projects.LocalRoleLead, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			user := &store.User{Role: tt.role}
			if got := templates.CanManage(user, tt.memberRole); got != tt.want {
				t.Errorf("CanManage() = %v, want %v", got, tt.want)
			}
		})
	}
}
