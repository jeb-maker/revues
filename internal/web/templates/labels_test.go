package templates_test

import (
	"testing"

	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/templates"
)

func TestRunLabelsForPreset_ListesEnCours(t *testing.T) {
	got := templates.RunLabelsForPreset(store.UIRunLabelListesEnCours)
	if got.Nav != "Listes en cours" {
		t.Fatalf("Nav = %q", got.Nav)
	}
	if got.NavShort != "En cours" {
		t.Fatalf("NavShort = %q", got.NavShort)
	}
	if got.Singular != "liste" || got.Article != "une" {
		t.Fatalf("Singular/Article = %q/%q", got.Singular, got.Article)
	}
	if cta := templates.LaunchRunCTA(got); cta != "Lancer une liste" {
		t.Fatalf("LaunchRunCTA = %q", cta)
	}
}

func TestRunLabelsForPreset_DefaultRevues(t *testing.T) {
	got := templates.RunLabelsForPreset("")
	if got.Nav != "Revues" || got.NavShort != "Revues" {
		t.Fatalf("default = %+v", got)
	}
}
