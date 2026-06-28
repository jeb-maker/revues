package store_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jeb-maker/revues/internal/store"
)

func TestIntegrationLinkUpsertAndList(t *testing.T) {
	ctx := context.Background()
	st, _ := testStore(t)

	if err := st.UpsertIntegrationByType(ctx, store.IntegrationTypeJira, true, []byte("cfg")); err != nil {
		t.Fatalf("UpsertIntegrationByType(): %v", err)
	}
	integration, err := st.GetIntegrationByType(ctx, store.IntegrationTypeJira)
	if err != nil {
		t.Fatalf("GetIntegrationByType(): %v", err)
	}

	user, err := st.UpsertGitHubUser(ctx, 1, "lead", "lead@example.com", "Lead", "", "editor")
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "P", "", user.ID)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, project.ID, "T", user.ID, []store.TemplateItemInput{
		{Label: "Point", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, "Run", user.ID, sql.NullString{})
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}
	items, err := st.ListRunItems(ctx, run.ID)
	if err != nil || len(items) != 1 {
		t.Fatalf("ListRunItems() = %v, %v", items, err)
	}

	link, err := st.UpsertIntegrationLink(ctx, items[0].ID, integration.ID, "REV-1", "https://jira.example.com/browse/REV-1")
	if err != nil {
		t.Fatalf("UpsertIntegrationLink(): %v", err)
	}
	if link.ExternalKey != "REV-1" {
		t.Fatalf("ExternalKey = %q", link.ExternalKey)
	}

	got, err := st.IntegrationLinkByRunItemAndType(ctx, items[0].ID, store.IntegrationTypeJira)
	if err != nil {
		t.Fatalf("IntegrationLinkByRunItemAndType(): %v", err)
	}
	if got.ExternalURL != link.ExternalURL {
		t.Fatalf("ExternalURL = %q, want %q", got.ExternalURL, link.ExternalURL)
	}

	links, err := st.ListIntegrationLinksByRunItemIDs(ctx, []int64{items[0].ID}, store.IntegrationTypeJira)
	if err != nil {
		t.Fatalf("ListIntegrationLinksByRunItemIDs(): %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("links len = %d, want 1", len(links))
	}

	updated, err := st.UpsertIntegrationLink(ctx, items[0].ID, integration.ID, "REV-2", "https://jira.example.com/browse/REV-2")
	if err != nil {
		t.Fatalf("UpsertIntegrationLink(update): %v", err)
	}
	if updated.ID != link.ID || updated.ExternalKey != "REV-2" {
		t.Fatalf("updated = %+v, want same id with REV-2", updated)
	}
}
