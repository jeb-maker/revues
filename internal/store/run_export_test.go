package store_test

import (
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	runs "github.com/jeb-maker/revues/internal/features/runs"
	"github.com/jeb-maker/revues/internal/store"
)

func TestListRunExportRows(t *testing.T) {
	ctx, st, run, itemID := setupInProgressRun(t)

	checker, err := st.UpsertGitHubUser(ctx, 2, "checker", "checker@example.com", "Checker", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	if err = st.UpdateRunItemStatus(ctx, run.ID, itemID, checker.ID, runs.StatusOK, ""); err != nil {
		t.Fatalf("UpdateRunItemStatus(): %v", err)
	}
	if err = st.CompleteRun(ctx, run.ID, "Terminé"); err != nil {
		t.Fatalf("CompleteRun(): %v", err)
	}

	rows, err := st.ListRunExportRows(ctx, run.ID)
	if err != nil {
		t.Fatalf("ListRunExportRows(): %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("len(rows) = %d, want 2", len(rows))
	}
	if rows[0].ProjectName != "P" || rows[0].RunTitle != "Revue" {
		t.Fatalf("first row metadata = %+v", rows[0])
	}
	if rows[0].PointLabel != "Point 1" || rows[0].Status != runs.StatusOK || rows[0].AuthorLogin != "checker" {
		t.Fatalf("first row item = %+v", rows[0])
	}
	if rows[0].RunDate == "" {
		t.Fatal("expected completed_at on export row")
	}
	if rows[1].PointLabel != "Point 2" || rows[1].Status != store.RunItemStatusPending {
		t.Fatalf("second row item = %+v", rows[1])
	}
}
