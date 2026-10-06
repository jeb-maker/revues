package store_test

import (
	"context"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
)

func TestSetRunConfluenceURL(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 91001, "conflead", "conflead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateSubject(ctx, "P Conf", "", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, "Modèle Conf", lead.ID, nil, []store.TemplateItemInput{
		{Section: "S1", Label: "Point 1", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, lead.ID)
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}

	url := "https://example.atlassian.net/wiki/spaces/REV/pages/1"
	if setErr := st.SetRunConfluenceURL(ctx, run.ID, url); setErr != nil {
		t.Fatalf("SetRunConfluenceURL(): %v", setErr)
	}
	got, err := st.RunByID(ctx, run.ID)
	if err != nil {
		t.Fatalf("RunByID(): %v", err)
	}
	if got.ConfluenceURL != url {
		t.Fatalf("ConfluenceURL = %q want %q", got.ConfluenceURL, url)
	}
}
