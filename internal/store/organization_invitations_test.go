package store_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
)

func TestOrganizationInvitations(t *testing.T) {
	ctx := context.Background()
	db := openInvitationsDB(t)
	st := store.New(db)

	owner, err := st.UpsertGitHubUser(ctx, 1, "owner", "owner@example.com", "Owner", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	org, err := st.CreateOrganization(ctx, "Acme", "acme", owner.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, org.ID, owner.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}
	orgCtx := orgctx.WithOrganizationID(ctx, org.ID)
	if _, err = st.CreateSubject(orgCtx, "Portal", "desc", owner.ID, nil); err != nil {
		t.Fatalf("CreateSubject(): %v", err)
	}

	email := "invitee@example.com"
	if err = st.CreateOrganizationInvitation(ctx, email, org.ID); err != nil {
		t.Fatalf("CreateOrganizationInvitation(): %v", err)
	}

	invites, err := st.ListPendingInvitationsByEmail(ctx, email)
	if err != nil {
		t.Fatalf("ListPendingInvitationsByEmail(): %v", err)
	}
	if len(invites) != 1 {
		t.Fatalf("invites len = %d, want 1", len(invites))
	}
	if invites[0].OrganizationName != "Acme" {
		t.Fatalf("org name = %q, want Acme", invites[0].OrganizationName)
	}

	ok, err := st.HasPendingInvitationByEmail(ctx, email)
	if err != nil || !ok {
		t.Fatalf("HasPendingInvitationByEmail() = %v, %v, want true", ok, err)
	}

	loaded, err := st.OrganizationInvitationByID(ctx, invites[0].ID)
	if err != nil {
		t.Fatalf("OrganizationInvitationByID(): %v", err)
	}
	if loaded.Email != email {
		t.Fatalf("loaded email = %q", loaded.Email)
	}

	if err = st.DeleteOrganizationInvitation(ctx, loaded.ID); err != nil {
		t.Fatalf("DeleteOrganizationInvitation(): %v", err)
	}

	ok, err = st.HasPendingInvitationByEmail(ctx, email)
	if err != nil || ok {
		t.Fatalf("HasPendingInvitationByEmail() after delete = %v, %v, want false", ok, err)
	}
}

func openInvitationsDB(t *testing.T) *sql.DB {
	t.Helper()
	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/invitations.db", 0)
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	t.Cleanup(func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Errorf("Close(): %v", closeErr)
		}
	})
	if migrateErr := store.Migrate(ctx, db); migrateErr != nil {
		t.Fatalf("Migrate(): %v", migrateErr)
	}
	return db
}
