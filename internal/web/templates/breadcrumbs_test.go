package templates_test

import (
	"testing"

	"github.com/jeb-maker/revues/internal/web/templates"
)

func TestApplyPageMeta_SetsTitleFromLastCrumb(t *testing.T) {
	data := templates.ApplyPageMeta(templates.PageData{}, templates.BCSubjectNew(templates.DefaultUILabels().Subject))
	if data.Title != "Nouveau sujet" {
		t.Fatalf("Title = %q, want Nouveau sujet", data.Title)
	}
	if len(data.Breadcrumbs) != 2 {
		t.Fatalf("len(Breadcrumbs) = %d, want 2", len(data.Breadcrumbs))
	}
}

func TestBCRunWizardTemplates_Links(t *testing.T) {
	crumbs := templates.BCRunWizardTemplates("Alpha", 3, templates.DefaultUILabels().Run, false)
	if len(crumbs) != 3 {
		t.Fatalf("len = %d, want 3", len(crumbs))
	}
	if crumbs[1].URL != "/subjects/3" {
		t.Fatalf("subject URL = %q", crumbs[1].URL)
	}
	if crumbs[2].Label != "Choisir un modèle" || crumbs[2].URL != "" {
		t.Fatalf("last crumb = %+v", crumbs[2])
	}
	listUI := templates.BCRunWizardTemplates("Alpha", 3, templates.DefaultUILabels().Run, true)
	if listUI[2].Label != "Choisir une liste" {
		t.Fatalf("listUI last crumb = %+v", listUI[2])
	}
}

func TestBreadcrumbCurrent_Empty(t *testing.T) {
	if got := templates.BreadcrumbCurrent(nil); got != "" {
		t.Fatalf("BreadcrumbCurrent(nil) = %q", got)
	}
}

func TestBreadcrumbAncestors(t *testing.T) {
	if got := templates.BreadcrumbAncestors(nil); got != nil {
		t.Fatalf("nil crumbs = %v", got)
	}
	run := templates.DefaultUILabels().Run
	one := templates.BCRevues(run)
	if got := templates.BreadcrumbAncestors(one); got != nil {
		t.Fatalf("single crumb = %v, want nil", got)
	}
	deep := templates.BCRunWizardTemplates("Alpha", 3, run, false)
	got := templates.BreadcrumbAncestors(deep)
	if len(got) != 2 || got[0].Label != "Revues" || got[1].Label != "Alpha" {
		t.Fatalf("ancestors = %+v", got)
	}
}

func TestBCTemplatesNewWizard_Labels(t *testing.T) {
	list := templates.BCTemplatesNewWizard(true)
	if len(list) != 2 {
		t.Fatalf("listUI len = %d, want 2", len(list))
	}
	if list[0].Label != "Listes" || list[0].URL != templates.PathTemplates {
		t.Fatalf("listUI parent = %+v", list[0])
	}
	if list[1].Label != "Nouvelle liste" || list[1].URL != "" {
		t.Fatalf("listUI current = %+v", list[1])
	}
	org := templates.BCTemplatesNewWizard(false)
	if len(org) != 2 {
		t.Fatalf("org len = %d, want 2", len(org))
	}
	if org[0].Label != "Modèles" || org[1].Label != "Nouveau modèle" {
		t.Fatalf("org crumbs = %+v", org)
	}
}

func TestBCAdminIntegrations_UsesOrganisationHub(t *testing.T) {
	crumbs := templates.BCAdminIntegrations()
	if len(crumbs) != 2 {
		t.Fatalf("len = %d, want 2", len(crumbs))
	}
	if crumbs[0].Label != "Organisation" || crumbs[0].URL != templates.PathAdminOrg {
		t.Fatalf("parent = %+v", crumbs[0])
	}
	if crumbs[1].Label != "Intégrations" || crumbs[1].URL != "" {
		t.Fatalf("current = %+v", crumbs[1])
	}
	jira := templates.BCAdminJira()
	if len(jira) != 3 || jira[0].URL != templates.PathAdminOrg || jira[1].URL != templates.PathAdmin {
		t.Fatalf("jira crumbs = %+v", jira)
	}
}
