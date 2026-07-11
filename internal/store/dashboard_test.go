package store_test

import (
	"context"
	"github.com/jeb-maker/revues/internal/testutil"
	"testing"

	runs "github.com/jeb-maker/revues/internal/features/runs"
	"github.com/jeb-maker/revues/internal/store"
)

func TestDashboard_ActiveRunsAndNokItems(t *testing.T) {
	ctx, st, run, itemID := setupInProgressRun(t)

	if err := st.UpdateRunItemStatus(ctx, run.ID, itemID, 1, runs.StatusOK, ""); err != nil {
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
	if err = st.UpdateRunItemStatus(ctx, run.ID, secondID, 1, runs.StatusNOK, "bloquant"); err != nil {
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

func TestDashboard_RecentCompletedRuns(t *testing.T) {
	ctx, st, run, itemID := setupInProgressRun(t)

	if err := st.UpdateRunItemStatus(ctx, run.ID, itemID, 1, runs.StatusOK, ""); err != nil {
		t.Fatalf("UpdateRunItemStatus(): %v", err)
	}
	runItems, err := st.ListRunItems(ctx, run.ID)
	if err != nil || len(runItems) != 2 {
		t.Fatalf("ListRunItems() = %v, %v", runItems, err)
	}
	for _, item := range runItems {
		if item.ID != itemID {
			if err = st.UpdateRunItemStatus(ctx, run.ID, item.ID, 1, runs.StatusOK, ""); err != nil {
				t.Fatalf("UpdateRunItemStatus(): %v", err)
			}
			break
		}
	}
	if err = st.CompleteRun(ctx, run.ID, "done"); err != nil {
		t.Fatalf("CompleteRun(): %v", err)
	}

	completed, err := st.ListRecentCompletedRunSummaries(ctx, 1, true)
	if err != nil {
		t.Fatalf("ListRecentCompletedRunSummaries(): %v", err)
	}
	if len(completed) != 1 {
		t.Fatalf("len(completed) = %d, want 1", len(completed))
	}
	if completed[0].RunID != run.ID || !completed[0].CompletedAt.Valid || completed[0].Percent != 100 {
		t.Fatalf("ListRecentCompletedRunSummaries() = %+v", completed)
	}

	active, err := st.ListActiveRunSummaries(ctx, 1, true)
	if err != nil {
		t.Fatalf("ListActiveRunSummaries(): %v", err)
	}
	if len(active) != 0 {
		t.Fatalf("len(active) = %d, want 0 after completion", len(active))
	}
}

func TestDashboard_TemplateIndexListsAllTemplates(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	alice, err := st.UpsertGitHubUser(ctx, 70, "alice", "alice@example.com", "Alice", "", "editor")
	if err != nil {
		t.Fatalf("UpsertGitHubUser(alice): %v", err)
	}
	bob, err := st.UpsertGitHubUser(ctx, 71, "bob", "bob@example.com", "Bob", "", "editor")
	if err != nil {
		t.Fatalf("UpsertGitHubUser(bob): %v", err)
	}

	_, err = st.CreateProject(ctx, "Alpha", "", alice.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	_, err = st.CreateProject(ctx, "Beta", "", bob.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(bob): %v", err)
	}
	if _, _, err = st.CreateChecklistTemplate(ctx, "Modèle A", alice.ID, nil, nil); err != nil {
		t.Fatalf("CreateChecklistTemplate(A): %v", err)
	}
	if _, _, err = st.CreateChecklistTemplate(ctx, "Modèle B", bob.ID, nil, nil); err != nil {
		t.Fatalf("CreateChecklistTemplate(B): %v", err)
	}

	aliceRows, err := st.ListTemplateIndex(ctx, alice.ID, false)
	if err != nil {
		t.Fatalf("ListTemplateIndex(alice): %v", err)
	}
	if len(aliceRows) != 2 {
		t.Fatalf("ListTemplateIndex(alice) len = %d, want 2", len(aliceRows))
	}

	adminRows, err := st.ListTemplateIndex(ctx, alice.ID, true)
	if err != nil {
		t.Fatalf("ListTemplateIndex(admin): %v", err)
	}
	if len(adminRows) != 2 {
		t.Fatalf("ListTemplateIndex(admin) len = %d, want 2", len(adminRows))
	}
}
