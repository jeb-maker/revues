package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/features/subjects"
	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
)

func TestCreateSubjectAddsOrgMember(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)
	ctx = defaultOrgCtx(ctx, st)

	creator, err := st.UpsertGitHubUser(ctx, 1, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	subject, err := st.CreateSubject(ctx, "Alpha", "desc", creator.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject(): %v", err)
	}

	role, ok, err := st.MemberRole(ctx, subject.ID, creator.ID)
	if err != nil || !ok || role != subjects.LocalRoleLead {
		t.Fatalf("MemberRole() = %q, %v, %v", role, ok, err)
	}

	subjects, err := st.ListSubjects(ctx, creator.ID, false, "")
	if err != nil || len(subjects) != 1 {
		t.Fatalf("ListSubjects() = %v, %v", subjects, err)
	}
}

func TestListSubjectsAdminSeesAll(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)
	ctx = defaultOrgCtx(ctx, st)

	a, err := st.UpsertGitHubUser(ctx, 1, "a", "a@example.com", "A", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(a): %v", err)
	}
	b, err := st.UpsertGitHubUser(ctx, 2, "b", "b@example.com", "B", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(b): %v", err)
	}

	_, err = st.CreateSubject(ctx, "P1", "", a.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject(p1): %v", err)
	}
	_, err = st.CreateSubject(ctx, "P2", "", b.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject(p2): %v", err)
	}

	items, err := st.ListSubjects(ctx, 0, true, "")
	if err != nil {
		t.Fatalf("ListSubjects(admin): %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("admin list len = %d, want 2", len(items))
	}

	items, err = st.ListSubjects(ctx, a.ID, false, "")
	if err != nil {
		t.Fatalf("ListSubjects(a): %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("member list len = %d, want 2 (v1 org-scoped access)", len(items))
	}
}

func TestListSubjectsSearch(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)
	ctx = defaultOrgCtx(ctx, st)

	user, err := st.UpsertGitHubUser(ctx, 3, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	if _, err = st.CreateSubject(ctx, "Alpha Platform", "Core services", user.ID, nil); err != nil {
		t.Fatalf("CreateSubject(alpha): %v", err)
	}
	if _, err = st.CreateSubject(ctx, "Beta", "", user.ID, nil); err != nil {
		t.Fatalf("CreateSubject(beta): %v", err)
	}

	byName, err := st.ListSubjects(ctx, user.ID, true, "alpha")
	if err != nil {
		t.Fatalf("ListSubjects(alpha): %v", err)
	}
	if len(byName) != 1 || byName[0].Name != "Alpha Platform" {
		t.Fatalf("ListSubjects(alpha) = %+v", byName)
	}

	byDesc, err := st.ListSubjects(ctx, user.ID, true, "services")
	if err != nil {
		t.Fatalf("ListSubjects(services): %v", err)
	}
	if len(byDesc) != 1 {
		t.Fatalf("len(byDesc) = %d, want 1", len(byDesc))
	}

	missing, err := st.ListSubjects(ctx, user.ID, true, "gamma")
	if err != nil {
		t.Fatalf("ListSubjects(gamma): %v", err)
	}
	if len(missing) != 0 {
		t.Fatalf("len(missing) = %d, want 0", len(missing))
	}
}

func TestSubjectByIDCrossOrganizationIDOR(t *testing.T) {
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
	subject, err := st.CreateSubject(ctxA, "Secret", "hidden", alice.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject(): %v", err)
	}

	ctxB := orgctx.WithOrganizationID(ctx, orgB.ID)
	_, err = st.SubjectByID(ctxB, subject.ID)
	if !errors.Is(err, store.ErrSubjectNotFound) {
		t.Fatalf("SubjectByID() error = %v, want ErrSubjectNotFound", err)
	}
}

func TestSubjectDomainsAndTags(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)
	ctx = defaultOrgCtx(ctx, st)

	user, err := st.UpsertGitHubUser(ctx, 1, "u", "u@example.com", "U", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	subject, err := st.CreateSubject(ctx, "P", "", user.ID, []string{"k8s", "frontend"})
	if err != nil {
		t.Fatalf("CreateSubject(): %v", err)
	}

	domains, err := st.ListSubjectDomains(ctx, subject.ID)
	if err != nil {
		t.Fatalf("ListSubjectDomains(): %v", err)
	}
	if len(domains) != 2 {
		t.Fatalf("domains = %v, want 2", domains)
	}

	if err = st.SetSubjectTags(ctx, subject.ID, []string{"prod", "critical"}); err != nil {
		t.Fatalf("SetSubjectTags(): %v", err)
	}
	tags, err := st.ListSubjectTags(ctx, subject.ID)
	if err != nil {
		t.Fatalf("ListSubjectTags(): %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("tags = %v, want 2", tags)
	}
}
