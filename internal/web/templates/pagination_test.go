package templates_test

import (
	"testing"

	"github.com/jeb-maker/revues/internal/web/templates"
)

func TestNewPagination(t *testing.T) {
	p := templates.NewPagination(2, 25, 60, func(page int) string {
		return templates.RunsListURL("", "", page)
	})
	if p.TotalPages != 3 || !p.HasPrev || !p.HasNext {
		t.Fatalf("pagination = %+v", p)
	}
	if p.PrevURL != "/revues" || p.NextURL != "/revues?page=3" {
		t.Fatalf("urls prev=%q next=%q", p.PrevURL, p.NextURL)
	}
}

func TestRunsListURL(t *testing.T) {
	if got := templates.RunsListURL("draft", "api", 1); got != "/revues?q=api&status=draft" {
		t.Fatalf("page1 = %q", got)
	}
	if got := templates.RunsListURL("", "", 2); got != "/revues?page=2" {
		t.Fatalf("page2 = %q", got)
	}
}
