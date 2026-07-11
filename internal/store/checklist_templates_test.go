package store_test

import (
	"context"
	"github.com/jeb-maker/revues/internal/testutil"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

func sampleItems() []store.TemplateItemInput {
	return []store.TemplateItemInput{
		{Section: "Général", Label: "Point A", HelpText: "Aide A", Required: true},
		{Section: "Général", Label: "Point B", HelpText: "", Required: false},
		{Section: "Sécurité", Label: "Point C", HelpText: "Aide C", Required: true},
	}
}

func TestCreateChecklistTemplateCreatesVersionOne(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 1, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "P", "", lead.ID)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}

	template, version, err := st.CreateChecklistTemplate(ctx, project.ID, "Modèle A", lead.ID, sampleItems())
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	if version.Version != 1 {
		t.Fatalf("version = %d, want 1", version.Version)
	}

	items, err := st.ListTemplateItems(ctx, version.ID)
	if err != nil {
		t.Fatalf("ListTemplateItems(): %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("items len = %d, want 3", len(items))
	}
	if items[0].Position != 1 || items[0].Section != "Général" || items[0].Label != "Point A" {
		t.Fatalf("first item = %+v", items[0])
	}
	if items[2].Section != "Sécurité" {
		t.Fatalf("third item section = %q", items[2].Section)
	}

	summaries, err := st.ListChecklistTemplates(ctx, project.ID)
	if err != nil {
		t.Fatalf("ListChecklistTemplates(): %v", err)
	}
	if len(summaries) != 1 || summaries[0].ID != template.ID || summaries[0].LatestVersion != 1 {
		t.Fatalf("summaries = %+v", summaries)
	}
}

func TestCreateTemplateVersionIncrements(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 1, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "P", "", lead.ID)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}

	template, v1, err := st.CreateChecklistTemplate(ctx, project.ID, "Modèle A", lead.ID, sampleItems())
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}

	updatedItems := []store.TemplateItemInput{
		{Section: "Général", Label: "Point A bis", HelpText: "", Required: true},
	}
	v2, err := st.CreateTemplateVersion(ctx, template.ID, lead.ID, updatedItems)
	if err != nil {
		t.Fatalf("CreateTemplateVersion(): %v", err)
	}
	if v2.Version != 2 {
		t.Fatalf("version = %d, want 2", v2.Version)
	}

	oldItems, err := st.ListTemplateItems(ctx, v1.ID)
	if err != nil {
		t.Fatalf("ListTemplateItems(v1): %v", err)
	}
	if len(oldItems) != 3 || oldItems[0].Label != "Point A" {
		t.Fatalf("old version mutated: %+v", oldItems)
	}

	newItems, err := st.ListTemplateItems(ctx, v2.ID)
	if err != nil {
		t.Fatalf("ListTemplateItems(v2): %v", err)
	}
	if len(newItems) != 1 || newItems[0].Label != "Point A bis" {
		t.Fatalf("new items = %+v", newItems)
	}
}
