package store_test

import (
	"context"
	"errors"
	"github.com/jeb-maker/revues/internal/testutil"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
)

func TestProjectByIDCrossOrganizationIDOR(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	alice, err := st.UpsertGitHubUser(ctx, 1, "alice", "alice@example.com", "Alice", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(alice): %v", err)
	}
	bob, err := st.UpsertGitHubUser(ctx, 2, "bob", "bob@example.com", "Bob", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(bob): %v", err)
	}

	orgA, err := st.CreateOrganization(ctx, "Org A", "org-a", alice.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(org-a): %v", err)
	}
	orgB, err := st.CreateOrganization(ctx, "Org B", "org-b", bob.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(org-b): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, orgA.ID, alice.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(alice): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, orgB.ID, bob.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(bob): %v", err)
	}

	ctxA := orgctx.WithOrganizationID(ctx, orgA.ID)
	project, err := st.CreateProject(ctxA, "Secret", "hidden", alice.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}

	ctxB := orgctx.WithOrganizationID(ctx, orgB.ID)
	_, err = st.ProjectByID(ctxB, project.ID)
	if !errors.Is(err, store.ErrProjectNotFound) {
		t.Fatalf("ProjectByID() error = %v, want ErrProjectNotFound", err)
	}
}
