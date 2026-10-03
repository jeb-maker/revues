package store_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/jeb-maker/revues/internal/testutil"

	"github.com/jeb-maker/revues/internal/auth"
	runs "github.com/jeb-maker/revues/internal/features/runs"
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
	project, err := st.CreateSubject(ctx, "P", "", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject(): %v", err)
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
	project, err := st.CreateSubject(ctx, "P", "", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, "Modèle", lead.ID, nil, nil)
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}

	tests := []struct {
		name    string
		dueDate sql.NullString
	}{
		{"with due date", sql.NullString{String: "2026-07-15T00:00:00Z", Valid: true}},
		{"without due date", sql.NullString{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run, err := st.CreateChecklistRunWithDueDate(ctx, project.ID, template.ID, lead.ID, tt.dueDate)
			if err != nil {
				t.Fatalf("CreateChecklistRunWithDueDate(): %v", err)
			}
			if run.Status != store.RunStatusInProgress {
				t.Fatalf("status = %q, want in_progress", run.Status)
			}
			if run.DueDate != tt.dueDate {
				t.Fatalf("due_date = %+v, want %+v", run.DueDate, tt.dueDate)
			}
			reloaded, err := st.RunByID(ctx, run.ID)
			if err != nil {
				t.Fatalf("RunByID(): %v", err)
			}
			if reloaded.DueDate != tt.dueDate {
				t.Fatalf("reloaded due_date = %+v, want %+v", reloaded.DueDate, tt.dueDate)
			}
		})
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
	project, err := st.CreateSubject(ctx, "P", "", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject(): %v", err)
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

func TestCompleteRunWithEvidence_RejectsPendingRequired(t *testing.T) {
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
	if err = st.StartRun(ctx, run.ID); err != nil {
		t.Fatal(err)
	}

	err = st.CompleteRunWithEvidence(ctx, run.ID, "note", "abc", "2026-10-03T12:00:00Z")
	if !errors.Is(err, store.ErrPendingRequiredItems) {
		t.Fatalf("CompleteRunWithEvidence() = %v, want ErrPendingRequiredItems", err)
	}
	got, err := st.RunByID(ctx, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != store.RunStatusInProgress {
		t.Fatalf("status = %q, want in_progress after rejected complete", got.Status)
	}
}

func TestCompleteRunWithEvidence_AtomicSeal(t *testing.T) {
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
	if err = st.StartRun(ctx, run.ID); err != nil {
		t.Fatal(err)
	}
	items, err := st.ListRunItems(ctx, run.ID)
	if err != nil || len(items) == 0 {
		t.Fatalf("ListRunItems: %v %v", items, err)
	}
	if err = st.UpdateRunItemStatus(ctx, run.ID, items[0].ID, lead.ID, store.RunItemStatusOK, ""); err != nil {
		t.Fatalf("UpdateRunItemStatus: %v", err)
	}

	hash := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	completedAt := "2026-10-03T12:00:00Z"
	if err = st.CompleteRunWithEvidence(ctx, run.ID, "done", hash, completedAt); err != nil {
		t.Fatalf("CompleteRunWithEvidence(): %v", err)
	}
	got, err := st.RunByID(ctx, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != store.RunStatusDone || got.EvidenceCSVSHA256 != hash {
		t.Fatalf("got status=%q hash=%q", got.Status, got.EvidenceCSVSHA256)
	}
	if !got.CompletedAt.Valid || got.CompletedAt.String != completedAt {
		t.Fatalf("completed_at = %+v, want %s", got.CompletedAt, completedAt)
	}
}

func TestCompleteRunWithEvidence_HashMatchesPostCompleteExport(t *testing.T) {
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
	if err = st.StartRun(ctx, run.ID); err != nil {
		t.Fatal(err)
	}
	items, err := st.ListRunItems(ctx, run.ID)
	if err != nil || len(items) == 0 {
		t.Fatalf("ListRunItems: %v %v", items, err)
	}
	if err = st.UpdateRunItemStatus(ctx, run.ID, items[0].ID, lead.ID, store.RunItemStatusOK, ""); err != nil {
		t.Fatal(err)
	}

	completedAt := "2026-10-03T15:30:00Z"
	csvRows, err := st.ListRunExportRows(ctx, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	for i := range csvRows {
		csvRows[i].RunDate = completedAt
	}
	csvData, err := runs.BuildRunCSV(csvRows)
	if err != nil {
		t.Fatal(err)
	}
	hash := runs.SHA256Hex(csvData)
	if err = st.CompleteRunWithEvidence(ctx, run.ID, "done", hash, completedAt); err != nil {
		t.Fatalf("CompleteRunWithEvidence(): %v", err)
	}

	afterRows, err := st.ListRunExportRows(ctx, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	afterCSV, err := runs.BuildRunCSV(afterRows)
	if err != nil {
		t.Fatal(err)
	}
	if got := runs.SHA256Hex(afterCSV); got != hash {
		t.Fatalf("post-complete CSV hash = %q, want sealed %q", got, hash)
	}
}
