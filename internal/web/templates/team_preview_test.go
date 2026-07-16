package templates_test

import (
	"testing"

	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/templates"
)

func TestTeamAssignPreview(t *testing.T) {
	got := templates.TeamAssignPreview("Squad", 1, store.SubjectRoleViewer)
	want := "Équipe Squad : 1 membre aura le rôle Observateur"
	if got != want {
		t.Fatalf("singular = %q, want %q", got, want)
	}
	got = templates.TeamAssignPreview("Squad", 3, store.SubjectRoleContributor)
	want = "Équipe Squad : 3 membres auront le rôle Contributeur"
	if got != want {
		t.Fatalf("plural = %q, want %q", got, want)
	}
}
