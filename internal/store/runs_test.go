package store_test

import (
	"context"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

func TestCreateChecklistRunSnapshotsItems(t *testing.T) {
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

	template, version, err := st.CreateChecklistTemplate(ctx, project.ID, "Modèle", lead.ID, []store.TemplateItemInput{
		{Section: "S1", Label: "Point 1", Required: true},
		{Section: "S2", Label: "Point 2", Required: false},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}

	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, "Revue Q1", lead.ID)
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}
	if run.Status != store.RunStatusDraft {
		t.Fatalf("status = %q, want draft", run.Status)
	}
	if run.TemplateVersionID != version.ID {
		t.Fatalf("template_version_id = %d, want %d", run.TemplateVersionID, version.ID)
	}

	items, err := st.ListRunItems(ctx, run.ID)
	if err != nil {
		t.Fatalf("ListRunItems(): %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("items len = %d, want 2", len(items))
	}
	if items[0].Label != "Point 1" || items[0].Status != "pending" || !items[0].SourceItemID.Valid {
		t.Fatalf("first item = %+v", items[0])
	}
}

func TestRunStatusTransitions(t *testing.T) {
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
		{Label: "Point", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}

	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, "Revue", lead.ID)
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}

	if err = st.StartRun(ctx, run.ID); err != nil {
		t.Fatalf("StartRun(): %v", err)
	}
	run, err = st.RunByID(ctx, run.ID)
	if err != nil {
		t.Fatalf("RunByID(): %v", err)
	}
	if run.Status != store.RunStatusInProgress || !run.StartedAt.Valid {
		t.Fatalf("after start: %+v", run)
	}

	if err = st.CompleteRun(ctx, run.ID, "Note de clôture"); err != nil {
		t.Fatalf("CompleteRun(): %v", err)
	}
	run, err = st.RunByID(ctx, run.ID)
	if err != nil {
		t.Fatalf("RunByID(): %v", err)
	}
	if run.Status != store.RunStatusDone || !run.CompletedAt.Valid {
		t.Fatalf("after complete: %+v", run)
	}
	if run.ClosingNote != "Note de clôture" {
		t.Fatalf("closing_note = %q", run.ClosingNote)
	}
}
