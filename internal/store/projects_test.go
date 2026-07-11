package store_test

import (
	"context"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/features/projects"
	"github.com/jeb-maker/revues/internal/store"
)

func TestCreateProjectAddsLead(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)
	ctx = defaultOrgCtx(ctx, st)

	creator, err := st.UpsertGitHubUser(ctx, 1, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	project, err := st.CreateProject(ctx, "Alpha", "desc", creator.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}

	role, ok, err := st.MemberRole(ctx, project.ID, creator.ID)
	if err != nil || !ok || role != projects.LocalRoleLead {
		t.Fatalf("MemberRole() = %q, %v, %v", role, ok, err)
	}

	projects, err := st.ListProjects(ctx, creator.ID, false, "")
	if err != nil || len(projects) != 1 {
		t.Fatalf("ListProjects() = %v, %v", projects, err)
	}
}

func TestListProjectsAdminSeesAll(t *testing.T) {
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

	_, err = st.CreateProject(ctx, "P1", "", a.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(p1): %v", err)
	}
	_, err = st.CreateProject(ctx, "P2", "", b.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(p2): %v", err)
	}

	items, err := st.ListProjects(ctx, 0, true, "")
	if err != nil {
		t.Fatalf("ListProjects(admin): %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("admin list len = %d, want 2", len(items))
	}

	items, err = st.ListProjects(ctx, a.ID, false, "")
	if err != nil {
		t.Fatalf("ListProjects(a): %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("member list len = %d, want 1", len(items))
	}
}

func TestListProjectsSearch(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)
	ctx = defaultOrgCtx(ctx, st)

	user, err := st.UpsertGitHubUser(ctx, 3, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	if _, err = st.CreateProject(ctx, "Alpha Platform", "Core services", user.ID, nil); err != nil {
		t.Fatalf("CreateProject(alpha): %v", err)
	}
	if _, err = st.CreateProject(ctx, "Beta", "", user.ID, nil); err != nil {
		t.Fatalf("CreateProject(beta): %v", err)
	}

	byName, err := st.ListProjects(ctx, user.ID, true, "alpha")
	if err != nil {
		t.Fatalf("ListProjects(alpha): %v", err)
	}
	if len(byName) != 1 || byName[0].Name != "Alpha Platform" {
		t.Fatalf("ListProjects(alpha) = %+v", byName)
	}

	byDesc, err := st.ListProjects(ctx, user.ID, true, "services")
	if err != nil {
		t.Fatalf("ListProjects(services): %v", err)
	}
	if len(byDesc) != 1 {
		t.Fatalf("len(byDesc) = %d, want 1", len(byDesc))
	}

	missing, err := st.ListProjects(ctx, user.ID, true, "gamma")
	if err != nil {
		t.Fatalf("ListProjects(gamma): %v", err)
	}
	if len(missing) != 0 {
		t.Fatalf("len(missing) = %d, want 0", len(missing))
	}
}
