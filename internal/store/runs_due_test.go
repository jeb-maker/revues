package store_test

import (
	"context"
	"database/sql"
	"github.com/jeb-maker/revues/internal/testutil"
	"testing"
	"time"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

func TestListRunsDueOn(t *testing.T) {
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
	template, _, err := st.CreateChecklistTemplate(ctx, "Modèle", lead.ID, nil, nil)
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}

	runDue, err := st.CreateChecklistRun(ctx, project.ID, template.ID, "Due tomorrow", lead.ID, sql.NullString{})
	if err != nil {
		t.Fatalf("CreateChecklistRun(due): %v", err)
	}
	runOther, err := st.CreateChecklistRun(ctx, project.ID, template.ID, "Due later", lead.ID, sql.NullString{})
	if err != nil {
		t.Fatalf("CreateChecklistRun(other): %v", err)
	}
	runDraft, err := st.CreateChecklistRun(ctx, project.ID, template.ID, "Draft", lead.ID, sql.NullString{})
	if err != nil {
		t.Fatalf("CreateChecklistRun(draft): %v", err)
	}

	tomorrow := time.Now().UTC().Add(24 * time.Hour).Format("2006-01-02")
	nextWeek := time.Now().UTC().Add(7 * 24 * time.Hour).Format("2006-01-02")

	if err = st.SetRunDueDate(ctx, runDue.ID, sql.NullString{String: tomorrow + "T00:00:00Z", Valid: true}); err != nil {
		t.Fatalf("SetRunDueDate(due): %v", err)
	}
	if err = st.SetRunDueDate(ctx, runOther.ID, sql.NullString{String: nextWeek + "T00:00:00Z", Valid: true}); err != nil {
		t.Fatalf("SetRunDueDate(other): %v", err)
	}
	if err = st.SetRunDueDate(ctx, runDraft.ID, sql.NullString{String: tomorrow + "T00:00:00Z", Valid: true}); err != nil {
		t.Fatalf("SetRunDueDate(draft): %v", err)
	}

	for _, id := range []int64{runDue.ID, runOther.ID} {
		if err = st.StartRun(ctx, id); err != nil {
			t.Fatalf("StartRun(%d): %v", id, err)
		}
	}

	got, err := st.ListRunsDueOn(ctx, tomorrow)
	if err != nil {
		t.Fatalf("ListRunsDueOn(): %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("ListRunsDueOn() len = %d, want 1", len(got))
	}
	if got[0].ID != runDue.ID {
		t.Fatalf("ListRunsDueOn()[0].ID = %d, want %d", got[0].ID, runDue.ID)
	}
}
