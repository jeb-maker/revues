package store_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jeb-maker/revues/internal/testutil"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

func TestCreateChecklistRunSnapshotsItems(t *testing.T) {
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

	template, version, err := st.CreateChecklistTemplate(ctx, "Modèle", lead.ID, nil, []store.TemplateItemInput{
		{Section: "S1", Label: "Point 1", Required: true},
		{Section: "S2", Label: "Point 2", Required: false},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}

	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, lead.ID)
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}
	if run.Status != store.RunStatusInProgress {
		t.Fatalf("status = %q, want in_progress", run.Status)
	}
	if !run.StartedAt.Valid {
		t.Fatal("expected started_at on new run")
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

func TestCreateChecklistRunWithDueDate(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 2, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "P", "", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, "Modèle", lead.ID, nil, nil)
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}

	dueDate := sql.NullString{String: "2026-07-15T00:00:00Z", Valid: true}
	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, lead.ID)
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}
	if err = st.SetRunDueDate(ctx, run.ID, dueDate); err != nil {
		t.Fatalf("SetRunDueDate(): %v", err)
	}
	run, err = st.RunByID(ctx, run.ID)
	if err != nil {
		t.Fatalf("RunByID(): %v", err)
	}
	if !run.DueDate.Valid || run.DueDate.String != dueDate.String {
		t.Fatalf("due_date = %+v, want %q", run.DueDate, dueDate.String)
	}
}

func TestRunStatusTransitions(t *testing.T) {
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
		{Label: "Point", Required: true},
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

func TestCompleteRun_SealsEvidenceHash(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 1, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatal(err)
	}
	subject, err := st.CreateSubject(ctx, "P", "", lead.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, "Modèle", lead.ID, nil, []store.TemplateItemInput{
		{Label: "Point", Required: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	run, err := st.CreateChecklistRun(ctx, subject.ID, template.ID, lead.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = st.CompleteRun(ctx, run.ID, ""); err != nil {
		t.Fatalf("CompleteRun(): %v", err)
	}
	hash := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	if err = st.SealRunEvidenceHash(ctx, run.ID, hash); err != nil {
		t.Fatalf("SealRunEvidenceHash(): %v", err)
	}
	got, err := st.RunByID(ctx, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.EvidenceCSVSHA256 != hash {
		t.Fatalf("EvidenceCSVSHA256 = %q, want sealed hash", got.EvidenceCSVSHA256)
	}
}
