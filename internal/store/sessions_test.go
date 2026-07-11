package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

func TestCreateSessionWithOrganization(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)

	user, err := st.UpsertGitHubUser(ctx, 1, "alice", "alice@example.com", "Alice", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(default): %v", err)
	}
	if err := st.AddOrganizationMember(ctx, defaultOrg.ID, user.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}

	if err := st.CreateSession(ctx, user.ID, defaultOrg.ID, "hash-1"); err != nil {
		t.Fatalf("CreateSession(): %v", err)
	}

	gotUserID, gotOrgID, err := st.SessionByTokenHash(ctx, "hash-1")
	if err != nil {
		t.Fatalf("SessionByTokenHash(): %v", err)
	}
	if gotUserID != user.ID || gotOrgID != defaultOrg.ID {
		t.Fatalf("session = (%d, %d), want (%d, %d)", gotUserID, gotOrgID, user.ID, defaultOrg.ID)
	}
}

func TestResolveSessionOrganizationID(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)

	user, err := st.UpsertGitHubUser(ctx, 2, "bob", "bob@example.com", "Bob", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	org, err := st.CreateOrganization(ctx, "Acme", "acme", user.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(): %v", err)
	}
	if err := st.AddOrganizationMember(ctx, org.ID, user.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}

	tests := []struct {
		name       string
		preferred  int64
		wantOrgID  int64
		wantMember bool
	}{
		{name: "preferred organization", preferred: org.ID, wantOrgID: org.ID, wantMember: true},
		{name: "single membership", preferred: 0, wantOrgID: org.ID, wantMember: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := st.ResolveSessionOrganizationID(ctx, user.ID, tt.preferred)
			if err != nil {
				t.Fatalf("ResolveSessionOrganizationID() error = %v", err)
			}
			if got != tt.wantOrgID {
				t.Fatalf("org id = %d, want %d", got, tt.wantOrgID)
			}
			_, ok, err := st.OrganizationMemberRole(ctx, got, user.ID)
			if err != nil || ok != tt.wantMember {
				t.Fatalf("OrganizationMemberRole() = %v, %v", ok, err)
			}
		})
	}
}

func TestResolveSessionOrganizationIDBootstrapsDefault(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)

	user, err := st.UpsertGitHubUser(ctx, 3, "carol", "carol@example.com", "Carol", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(default): %v", err)
	}

	got, err := st.ResolveSessionOrganizationID(ctx, user.ID, 0)
	if err != nil {
		t.Fatalf("ResolveSessionOrganizationID() error = %v", err)
	}
	if got != defaultOrg.ID {
		t.Fatalf("org id = %d, want %d", got, defaultOrg.ID)
	}

	role, ok, err := st.OrganizationMemberRole(ctx, defaultOrg.ID, user.ID)
	if err != nil || !ok || role != store.OrgRoleMember {
		t.Fatalf("OrganizationMemberRole() = %q, %v, %v", role, ok, err)
	}
}

func TestSessionByTokenHashNotFound(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)

	_, _, err := st.SessionByTokenHash(ctx, "missing")
	if !errors.Is(err, store.ErrSessionNotFound) {
		t.Fatalf("SessionByTokenHash() error = %v, want %v", err, store.ErrSessionNotFound)
	}
}
