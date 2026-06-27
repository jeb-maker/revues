package runs_test

import (
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/projects"
	"github.com/jeb-maker/revues/internal/runs"
	"github.com/jeb-maker/revues/internal/store"
)

func TestCanLaunch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		role       string
		memberRole string
		want       bool
	}{
		{"admin", auth.RoleAdmin, "", true},
		{"editor lead", auth.RoleEditor, projects.LocalRoleLead, true},
		{"editor contributor", auth.RoleEditor, projects.LocalRoleContributor, true},
		{"editor viewer", auth.RoleEditor, projects.LocalRoleViewer, false},
		{"reader contributor", auth.RoleReader, projects.LocalRoleContributor, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			user := &store.User{Role: tt.role}
			if got := runs.CanLaunch(user, tt.memberRole); got != tt.want {
				t.Errorf("CanLaunch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanComplete(t *testing.T) {
	t.Parallel()

	user := &store.User{Role: auth.RoleEditor}
	if runs.CanComplete(user, projects.LocalRoleContributor) {
		t.Fatal("contributor should not complete runs")
	}
	if !runs.CanComplete(user, projects.LocalRoleLead) {
		t.Fatal("lead should complete runs")
	}
}
