package templates_test

import (
	"testing"

	"github.com/jeb-maker/revues/internal/web/templates"
)

func TestApplyPageMeta_SetsTitleFromLastCrumb(t *testing.T) {
	data := templates.ApplyPageMeta(templates.PageData{}, templates.BCProjectNew())
	if data.Title != "Nouveau" {
		t.Fatalf("Title = %q, want Nouveau", data.Title)
	}
	if len(data.Breadcrumbs) != 2 {
		t.Fatalf("len(Breadcrumbs) = %d, want 2", len(data.Breadcrumbs))
	}
}

func TestBCRunWizardLaunch_Links(t *testing.T) {
	crumbs := templates.BCRunWizardLaunch("Alpha", 3, "Checklist QA")
	if len(crumbs) != 4 {
		t.Fatalf("len = %d, want 4", len(crumbs))
	}
	if crumbs[0].URL != templates.PathRevues {
		t.Fatalf("root URL = %q", crumbs[0].URL)
	}
	if crumbs[3].Label != "Checklist QA" || crumbs[3].URL != "" {
		t.Fatalf("last crumb = %+v", crumbs[3])
	}
}

func TestBreadcrumbCurrent_Empty(t *testing.T) {
	if got := templates.BreadcrumbCurrent(nil); got != "" {
		t.Fatalf("BreadcrumbCurrent(nil) = %q", got)
	}
}
