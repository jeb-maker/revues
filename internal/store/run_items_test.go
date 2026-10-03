package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jeb-maker/revues/internal/testutil"

	"github.com/jeb-maker/revues/internal/auth"
	runs "github.com/jeb-maker/revues/internal/features/runs"
	"github.com/jeb-maker/revues/internal/store"
)

func setupInProgressRun(t *testing.T) (context.Context, *store.Store, *store.ChecklistRun, int64) {
	t.Helper()
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 1, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "P", "", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, "Modèle", lead.ID, nil, []store.TemplateItemInput{
		{Section: "S", Label: "Point 1", Required: true},
		{Section: "S", Label: "Point 2", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, lead.ID)
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}
	if err = st.StartRun(ctx, run.ID); err != nil {
		t.Fatalf("StartRun(): %v", err)
	}
	runItems, err := st.ListRunItems(ctx, run.ID)
	if err != nil || len(runItems) == 0 {
		t.Fatalf("ListRunItems() = %v, %v", runItems, err)
	}
	return ctx, st, run, runItems[0].ID
}

func TestUpdateRunItemStatusStoresNokWithComment(t *testing.T) {
	ctx, st, run, itemID := setupInProgressRun(t)

	if err := runs.ValidateUpdate(runs.StatusNOK, ""); !errors.Is(err, runs.ErrCommentRequired) {
		t.Fatalf("ValidateUpdate() should require comment")
	}

	err := st.UpdateRunItemStatus(ctx, run.ID, itemID, 1, runs.StatusNOK, "Détail du problème")
	if err != nil {
		t.Fatalf("UpdateRunItemStatus(): %v", err)
	}

	item, err := st.RunItemByID(ctx, run.ID, itemID)
	if err != nil {
		t.Fatalf("RunItemByID(): %v", err)
	}
	if item.Status != runs.StatusNOK || item.Comment != "Détail du problème" {
		t.Fatalf("item = %+v", item)
	}
}

func TestUpdateRunItemStatusChecked_OptimisticLock(t *testing.T) {
	const stale = "2000-01-01T00:00:00Z"
	tests := []struct {
		name       string
		expected   func(current string) string
		wantErr    error
		wantStatus string
	}{
		{"matching updated_at applies", func(c string) string { return c }, nil, runs.StatusOK},
		{"stale updated_at conflicts", func(string) string { return stale }, store.ErrRunItemConflict, runs.StatusPending},
		{"empty updated_at skips the check", func(string) string { return "" }, nil, runs.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, st, run, itemID := setupInProgressRun(t)
			before, err := st.RunItemByID(ctx, run.ID, itemID)
			if err != nil {
				t.Fatalf("RunItemByID(): %v", err)
			}

			err = st.UpdateRunItemStatusChecked(ctx, run.ID, itemID, 1, runs.StatusOK, "", tt.expected(before.UpdatedAt))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("UpdateRunItemStatusChecked() error = %v, want %v", err, tt.wantErr)
			}

			after, err := st.RunItemByID(ctx, run.ID, itemID)
			if err != nil {
				t.Fatalf("RunItemByID(after): %v", err)
			}
			if after.Status != tt.wantStatus {
				t.Fatalf("status = %q, want %q", after.Status, tt.wantStatus)
			}
			if tt.wantErr != nil && after.UpdatedAt != before.UpdatedAt {
				t.Fatalf("updated_at changed on conflict: %q → %q", before.UpdatedAt, after.UpdatedAt)
			}
		})
	}
}

func TestUpdateRunItemStatusChecked_MissingItem(t *testing.T) {
	ctx, st, run, _ := setupInProgressRun(t)
	err := st.UpdateRunItemStatusChecked(ctx, run.ID, 999999, 1, runs.StatusOK, "", "2000-01-01T00:00:00Z")
	if !errors.Is(err, store.ErrRunItemNotFound) {
		t.Fatalf("error = %v, want ErrRunItemNotFound", err)
	}
}

func TestCompleteRunStoresClosingNote(t *testing.T) {
	ctx, st, run, itemID := setupInProgressRun(t)

	if err := st.UpdateRunItemStatus(ctx, run.ID, itemID, 1, runs.StatusNOK, "Bloqué"); err != nil {
		t.Fatalf("UpdateRunItemStatus(): %v", err)
	}

	nokItems, err := st.ListNokRunItems(ctx, run.ID)
	if err != nil || len(nokItems) != 1 {
		t.Fatalf("ListNokRunItems() = %v, %v", nokItems, err)
	}

	if err = st.CompleteRun(ctx, run.ID, "Revue clôturée avec un point nok"); err != nil {
		t.Fatalf("CompleteRun(): %v", err)
	}

	updated, err := st.RunByID(ctx, run.ID)
	if err != nil {
		t.Fatalf("RunByID(): %v", err)
	}
	if updated.Status != store.RunStatusDone || updated.ClosingNote != "Revue clôturée avec un point nok" {
		t.Fatalf("run = %+v", updated)
	}
}
