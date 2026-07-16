package runs_test

import (
	"testing"

	"github.com/jeb-maker/revues/internal/features/runs"
)

func TestRunDisplayLabel(t *testing.T) {
	got := runs.RunDisplayLabel("Modèle QA", "Portail", "2026-07-14T10:00:00Z", 42)
	want := "Modèle QA · Portail · 14/07/2026 · #42"
	if got != want {
		t.Fatalf("RunDisplayLabel() = %q, want %q", got, want)
	}
}

func TestRunDisplayLabel_NoID(t *testing.T) {
	got := runs.RunDisplayLabel("Modèle", "Sujet", "2026-01-02T00:00:00Z", 0)
	want := "Modèle · Sujet · 02/01/2026"
	if got != want {
		t.Fatalf("RunDisplayLabel() = %q, want %q", got, want)
	}
}
