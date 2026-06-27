package store_test

import (
	"context"
	"testing"

	"github.com/jeb-maker/revues/internal/items"
	"github.com/jeb-maker/revues/internal/store"
)

func TestDashboard_ActiveRunsAndNokItems(t *testing.T) {
	ctx := context.Background()
	st, run, itemID := setupInProgressRun(t)

	if err := st.UpdateRunItemStatus(ctx, run.ID, itemID, 1, items.StatusOK, ""); err != nil {
		t.Fatalf("UpdateRunItemStatus(ok): %v", err)
	}
	runItems, err := st.ListRunItems(ctx, run.ID)
	if err != nil || len(runItems) != 2 {
		t.Fatalf("ListRunItems() = %v, %v", runItems, err)
	}
	var secondID int64
	for _, item := range runItems {
		if item.ID != itemID {
			secondID = item.ID
			break
		}
	}
	if err := st.UpdateRunItemStatus(ctx, run.ID, secondID, 1, items.StatusNOK, "bloquant"); err != nil {
		t.Fatalf("UpdateRunItemStatus(nok): %v", err)
	}

	summaries, err := st.ListActiveRunSummaries(ctx, 1, true)
	if err != nil {
		t.Fatalf("ListActiveRunSummaries(): %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("len(summaries) = %d, want 1", len(summaries))
	}
	if summaries[0].Percent != 50 {
		t.Fatalf("percent = %d, want 50", summaries[0].Percent)
	}

	nokItems, err := st.ListProjectNokItems(ctx, run.ProjectID)
	if err != nil {
		t.Fatalf("ListProjectNokItems(): %v", err)
	}
	if len(nokItems) != 1 || nokItems[0].Comment != "bloquant" {
		t.Fatalf("ListProjectNokItems() = %+v", nokItems)
	}

	runs, err := st.ListRunsWithProgressByProject(ctx, run.ProjectID)
	if err != nil {
		t.Fatalf("ListRunsWithProgressByProject(): %v", err)
	}
	if len(runs) != 1 || runs[0].Percent != 50 {
		t.Fatalf("ListRunsWithProgressByProject() = %+v", runs)
	}
}

func TestDashboard_TemplateIndexRespectsMembership(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)

	alice, err := st.UpsertGitHubUser(ctx, 70, "alice", "alice@example.com", "Alice", "", "editor")
	if err != nil {
		t.Fatalf("UpsertGitHubUser(alice): %v", err)
	}
	bob, err := st.UpsertGitHubUser(ctx, 71, "bob", "bob@example.com", "Bob", "", "editor")
	if err != nil {
		t.Fatalf("UpsertGitHubUser(bob): %v", err)
	}

	projectA, err := st.CreateProject(ctx, "Alpha", "", alice.ID)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	projectB, err := st.CreateProject(ctx, "Beta", "", bob.ID)
	if err != nil {
		t.Fatalf("CreateProject(bob): %v", err)
	}
	if _, _, err = st.CreateChecklistTemplate(ctx, projectA.ID, "Modèle A", alice.ID, nil); err != nil {
		t.Fatalf("CreateChecklistTemplate(A): %v", err)
	}
	if _, _, err = st.CreateChecklistTemplate(ctx, projectB.ID, "Modèle B", bob.ID, nil); err != nil {
		t.Fatalf("CreateChecklistTemplate(B): %v", err)
	}

	aliceRows, err := st.ListTemplateIndex(ctx, alice.ID, false)
	if err != nil {
		t.Fatalf("ListTemplateIndex(alice): %v", err)
	}
	if len(aliceRows) != 1 || aliceRows[0].Name != "Modèle A" {
		t.Fatalf("ListTemplateIndex(alice) = %+v", aliceRows)
	}

	adminRows, err := st.ListTemplateIndex(ctx, alice.ID, true)
	if err != nil {
		t.Fatalf("ListTemplateIndex(admin): %v", err)
	}
	if len(adminRows) != 2 {
		t.Fatalf("ListTemplateIndex(admin) len = %d, want 2", len(adminRows))
	}
}
