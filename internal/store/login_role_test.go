package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

func TestResolveLoginRole(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)
	bootstrap := "admin@example.com"

	t.Run("bootstrap admin", func(t *testing.T) {
		role, err := st.ResolveLoginRole(ctx, "Admin@Example.com", bootstrap)
		if err != nil {
			t.Fatalf("ResolveLoginRole() error = %v", err)
		}
		if role != auth.RoleAdmin {
			t.Errorf("role = %q, want admin", role)
		}
	})

	t.Run("self-service when not invited", func(t *testing.T) {
		role, err := st.ResolveLoginRole(ctx, "other@example.com", "")
		if err != nil {
			t.Fatalf("ResolveLoginRole() error = %v", err)
		}
		if role != auth.RoleEditor {
			t.Errorf("role = %q, want editor", role)
		}
	})

	t.Run("strict rejects unknown email", func(t *testing.T) {
		_, err := st.ResolveLoginRoleStrict(ctx, "stranger@example.com", bootstrap, true)
		if !errors.Is(err, store.ErrEmailNotAllowed) {
			t.Fatalf("error = %v, want ErrEmailNotAllowed", err)
		}
	})

	t.Run("pending invitation allows login", func(t *testing.T) {
		owner, err := st.UpsertGitHubUser(ctx, 10, "owner-inv", "owner-inv@example.com", "Owner", "", auth.RoleEditor)
		if err != nil {
			t.Fatalf("UpsertGitHubUser: %v", err)
		}
		org, err := st.CreateOrganization(ctx, "Invite Org", "invite-org-login", owner.ID)
		if err != nil {
			t.Fatalf("CreateOrganization: %v", err)
		}
		if err = st.AddOrganizationMember(ctx, org.ID, owner.ID, store.OrgRoleOwner); err != nil {
			t.Fatalf("AddOrganizationMember: %v", err)
		}
		if err = st.CreateOrganizationInvitation(ctx, "invitee@example.com", org.ID, store.OrgRoleMember); err != nil {
			t.Fatalf("CreateOrganizationInvitation: %v", err)
		}

		role, err := st.ResolveLoginRoleStrict(ctx, "invitee@example.com", bootstrap, true)
		if err != nil {
			t.Fatalf("ResolveLoginRoleStrict(): %v", err)
		}
		if role != auth.RoleEditor {
			t.Errorf("role = %q, want editor", role)
		}
	})

	t.Run("existing org member keeps role", func(t *testing.T) {
		user, err := st.UpsertGitHubUser(ctx, 11, "member-keep", "member-keep@example.com", "Member", "", auth.RoleReader)
		if err != nil {
			t.Fatalf("UpsertGitHubUser: %v", err)
		}
		org, err := st.CreateOrganization(ctx, "Keep Org", "keep-org-login", user.ID)
		if err != nil {
			t.Fatalf("CreateOrganization: %v", err)
		}
		if err = st.AddOrganizationMember(ctx, org.ID, user.ID, store.OrgRoleMember); err != nil {
			t.Fatalf("AddOrganizationMember: %v", err)
		}

		role, err := st.ResolveLoginRoleStrict(ctx, user.Email, bootstrap, true)
		if err != nil {
			t.Fatalf("ResolveLoginRoleStrict(): %v", err)
		}
		if role != auth.RoleReader {
			t.Errorf("role = %q, want reader", role)
		}
	})
}

func TestEnsureBootstrapOrgOwner(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)

	user, err := st.UpsertGitHubUser(ctx, 1, "bootstrap", "admin@example.com", "Admin", "", auth.RoleAdmin)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	if err = st.EnsureBootstrapOrgOwner(ctx, user.ID, "admin@example.com", "admin@example.com"); err != nil {
		t.Fatalf("EnsureBootstrapOrgOwner(): %v", err)
	}

	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}
	role, ok, err := st.OrganizationMemberRole(ctx, defaultOrg.ID, user.ID)
	if err != nil {
		t.Fatalf("OrganizationMemberRole(): %v", err)
	}
	if !ok || role != store.OrgRoleOwner {
		t.Fatalf("role = %q, member = %v, want owner", role, ok)
	}
}
