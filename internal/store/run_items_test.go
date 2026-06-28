package store_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/items"
	"github.com/jeb-maker/revues/internal/store"
)

func setupInProgressRun(t *testing.T) (*store.Store, *store.ChecklistRun, int64) {
	t.Helper()
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)

	lead, err := st.UpsertGitHubUser(ctx, 1, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "P", "", lead.ID)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, project.ID, "Modèle", lead.ID, []store.TemplateItemInput{
		{Section: "S", Label: "Point 1", Required: true},
		{Section: "S", Label: "Point 2", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, "Revue", lead.ID, sql.NullString{})
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
	return st, run, runItems[0].ID
}

func TestUpdateRunItemStatusStoresNokWithComment(t *testing.T) {
	ctx := context.Background()
	st, run, itemID := setupInProgressRun(t)

	if err := items.ValidateUpdate(items.StatusNOK, ""); !errors.Is(err, items.ErrCommentRequired) {
		t.Fatalf("ValidateUpdate() should require comment")
	}

	err := st.UpdateRunItemStatus(ctx, run.ID, itemID, 1, items.StatusNOK, "Détail du problème")
	if err != nil {
		t.Fatalf("UpdateRunItemStatus(): %v", err)
	}

	item, err := st.RunItemByID(ctx, run.ID, itemID)
	if err != nil {
		t.Fatalf("RunItemByID(): %v", err)
	}
	if item.Status != items.StatusNOK || item.Comment != "Détail du problème" {
		t.Fatalf("item = %+v", item)
	}
}

func TestCompleteRunStoresClosingNote(t *testing.T) {
	ctx := context.Background()
	st, run, itemID := setupInProgressRun(t)

	if err := st.UpdateRunItemStatus(ctx, run.ID, itemID, 1, items.StatusNOK, "Bloqué"); err != nil {
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
