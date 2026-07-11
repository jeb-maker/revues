package store_test

import (
	"context"
	"database/sql"
	"github.com/jeb-maker/revues/internal/testutil"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

func TestReplaceAttachment_OnePerRunItem(t *testing.T) {
	ctx := context.Background()
	db := openAttachmentTestDB(t)
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	user, err := st.UpsertGitHubUser(ctx, 1, "u", "u@example.com", "U", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "P", "", user.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, "T", user.ID, nil, []store.TemplateItemInput{{Label: "X"}})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, "R", user.ID, sql.NullString{})
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}
	items, err := st.ListRunItems(ctx, run.ID)
	if err != nil || len(items) != 1 {
		t.Fatalf("ListRunItems(): %v", err)
	}
	itemID := items[0].ID

	if _, err = st.ReplaceAttachment(ctx, itemID, "a.pdf", "application/pdf", "uuid-1.pdf", 100); err != nil {
		t.Fatalf("ReplaceAttachment first: %v", err)
	}
	if _, err = st.ReplaceAttachment(ctx, itemID, "b.pdf", "application/pdf", "uuid-2.pdf", 200); err != nil {
		t.Fatalf("ReplaceAttachment second: %v", err)
	}
	got, err := st.AttachmentByRunItemID(ctx, itemID)
	if err != nil {
		t.Fatalf("AttachmentByRunItemID(): %v", err)
	}
	if got.Filename != "b.pdf" || got.StoragePath != "uuid-2.pdf" {
		t.Fatalf("attachment = %+v", got)
	}
}

func TestListAttachmentsByRunItemIDs(t *testing.T) {
	ctx := context.Background()
	db := openAttachmentTestDB(t)
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	user, err := st.UpsertGitHubUser(ctx, 2, "u2", "u2@example.com", "U2", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "P2", "", user.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, "T", user.ID, nil, []store.TemplateItemInput{
		{Label: "A"},
		{Label: "B"},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, "R", user.ID, sql.NullString{})
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}
	items, err := st.ListRunItems(ctx, run.ID)
	if err != nil || len(items) != 2 {
		t.Fatalf("ListRunItems(): %v", err)
	}
	if _, err = st.ReplaceAttachment(ctx, items[0].ID, "a.jpg", "image/jpeg", "a.jpg", 10); err != nil {
		t.Fatalf("ReplaceAttachment(): %v", err)
	}

	got, err := st.ListAttachmentsByRunItemIDs(ctx, []int64{items[0].ID, items[1].ID})
	if err != nil {
		t.Fatalf("ListAttachmentsByRunItemIDs(): %v", err)
	}
	if len(got) != 1 || got[items[0].ID].Filename != "a.jpg" {
		t.Fatalf("attachments = %+v", got)
	}
}

func openAttachmentTestDB(t *testing.T) *sql.DB {
	t.Helper()
	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/test.db", 0)
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := store.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate(): %v", err)
	}
	return db
}
