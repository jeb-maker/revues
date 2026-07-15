package subjects_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/features/subjects"
	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
)

// TestIDOR_CrossSubject ensures org A members cannot access subject B in org B via store lookups.
func TestIDOR_CrossSubject(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/subjects-idor.db", 0)
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err = store.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate(): %v", err)
	}

	st := store.New(db)

	alice, err := st.UpsertGitHubUser(ctx, 1, "alice", "alice@example.com", "Alice", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(alice): %v", err)
	}
	bob, err := st.UpsertGitHubUser(ctx, 2, "bob", "bob@example.com", "Bob", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(bob): %v", err)
	}

	orgA, err := st.CreateOrganization(ctx, "Org A", "org-a-idor", alice.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(orgA): %v", err)
	}
	orgB, err := st.CreateOrganization(ctx, "Org B", "org-b-idor", bob.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(orgB): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, orgA.ID, alice.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(alice): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, orgB.ID, bob.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(bob): %v", err)
	}

	ctxA := orgctx.WithOrganizationID(ctx, orgA.ID)
	subjectA, err := st.CreateSubject(ctxA, "Secret A", "", alice.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject(): %v", err)
	}

	ctxB := orgctx.WithOrganizationID(ctx, orgB.ID)
	_, err = st.SubjectByID(ctxB, subjectA.ID)
	if !errors.Is(err, store.ErrSubjectNotFound) {
		t.Fatalf("SubjectByID(cross-org) error = %v, want ErrSubjectNotFound", err)
	}

	_, isMember, err := st.MemberRole(ctxB, subjectA.ID, bob.ID)
	if err != nil {
		t.Fatalf("MemberRole(): %v", err)
	}
	if isMember {
		t.Fatal("bob must not be member of subject A org via subject B session")
	}
	if subjects.CanViewSubject(bob, isMember) {
		t.Fatal("CanViewSubject(bob) must be false for cross-org subject")
	}
	if subjects.CanLaunchRun(bob, isMember) {
		t.Fatal("CanLaunchRun(bob) must be false for cross-org subject")
	}

	template, _, err := st.CreateChecklistTemplate(ctxA, "Modèle", alice.ID, nil, []store.TemplateItemInput{
		{Label: "Point", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctxA, subjectA.ID, template.ID, alice.ID)
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}
	if _, err = st.RunByID(ctxA, run.ID); err != nil {
		t.Fatalf("RunByID(org A) = %v", err)
	}
}
