package store_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/pressly/goose/v3"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/migrations"
)

func TestNormalizeOrganizationSlug(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{name: "lowercase trim", input: "  Acme Corp  ", want: "acme-corp"},
		{name: "underscores", input: "my_org", want: "my-org"},
		{name: "already valid", input: "acme-42", want: "acme-42"},
		{name: "empty", input: "   ", wantErr: store.ErrInvalidOrganizationSlug},
		{name: "invalid only symbols", input: "!!!", wantErr: store.ErrInvalidOrganizationSlug},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.NormalizeOrganizationSlug(tt.input)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("NormalizeOrganizationSlug(%q) error = %v, want %v", tt.input, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeOrganizationSlug(%q) error = %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("NormalizeOrganizationSlug(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCreateOrganization(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)

	creator, err := st.UpsertGitHubUser(ctx, 1, "owner", "owner@example.com", "Owner", "", auth.RoleAdmin)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	tests := []struct {
		name    string
		orgName string
		slug    string
		wantErr error
	}{
		{name: "creates with normalized slug", orgName: "Acme Corp", slug: "Acme-Corp"},
		{name: "duplicate slug", orgName: "Other", slug: "acme-corp", wantErr: store.ErrOrganizationSlugTaken},
	}

	var firstID int64
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			org, err := st.CreateOrganization(ctx, tt.orgName, tt.slug, creator.ID)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("CreateOrganization() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("CreateOrganization() error = %v", err)
			}
			if org.Slug != "acme-corp" {
				t.Fatalf("slug = %q, want acme-corp", org.Slug)
			}
			if !org.CreatedBy.Valid || org.CreatedBy.Int64 != creator.ID {
				t.Fatalf("created_by = %v, want %d", org.CreatedBy, creator.ID)
			}

			bySlug, err := st.OrganizationBySlug(ctx, "ACME-CORP")
			if err != nil || bySlug.ID != org.ID {
				t.Fatalf("OrganizationBySlug() = %v, %v", bySlug, err)
			}
			byID, err := st.OrganizationByID(ctx, org.ID)
			if err != nil || byID.Slug != org.Slug {
				t.Fatalf("OrganizationByID() = %v, %v", byID, err)
			}
			firstID = org.ID
		})
	}

	if firstID == 0 {
		t.Fatal("expected first organization id")
	}
}

func TestOrganizationMemberships(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)

	owner, err := st.UpsertGitHubUser(ctx, 1, "owner", "owner@example.com", "Owner", "", auth.RoleAdmin)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(owner): %v", err)
	}
	member, err := st.UpsertGitHubUser(ctx, 2, "member", "member@example.com", "Member", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(member): %v", err)
	}

	org, err := st.CreateOrganization(ctx, "Team", "team", owner.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(): %v", err)
	}

	if err = st.AddOrganizationMember(ctx, org.ID, owner.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(owner): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, org.ID, member.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(member): %v", err)
	}

	role, ok, err := st.OrganizationMemberRole(ctx, org.ID, member.ID)
	if err != nil || !ok || role != store.OrgRoleMember {
		t.Fatalf("OrganizationMemberRole() = %q, %v, %v", role, ok, err)
	}

	count, err := st.CountUserOrganizations(ctx, member.ID)
	if err != nil || count != 1 {
		t.Fatalf("CountUserOrganizations() = %d, %v", count, err)
	}

	memberships, err := st.ListUserOrganizations(ctx, member.ID)
	if err != nil {
		t.Fatalf("ListUserOrganizations(): %v", err)
	}
	if len(memberships) != 1 || memberships[0].Organization.Slug != "team" || memberships[0].Role != store.OrgRoleMember {
		t.Fatalf("ListUserOrganizations() = %+v", memberships)
	}

	if err = st.RemoveOrganizationMember(ctx, org.ID, member.ID); err != nil {
		t.Fatalf("RemoveOrganizationMember(): %v", err)
	}
	if err = st.RemoveOrganizationMember(ctx, org.ID, member.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("RemoveOrganizationMember() missing error = %v", err)
	}

	count, err = st.CountUserOrganizations(ctx, member.ID)
	if err != nil || count != 0 {
		t.Fatalf("CountUserOrganizations() after remove = %d, %v", count, err)
	}
}

func TestDefaultOrganizationExistsAfterMigrate(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)

	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(default): %v", err)
	}
	if defaultOrg.Name != "Default" {
		t.Fatalf("default org name = %q, want Default", defaultOrg.Name)
	}
}

func TestMigrationBackfillExistingUsers(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	t.Cleanup(func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Errorf("Close(): %v", closeErr)
		}
	})
	if _, err = db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("foreign_keys: %v", err)
	}

	migrateToVersion(t, ctx, db, 7)

	now := "2026-01-01T00:00:00Z"
	if _, err = db.ExecContext(ctx, `
		INSERT INTO users (github_id, login, email, display_name, avatar_url, role, created_at, last_login_at)
		VALUES
			(1, 'admin', 'admin@example.com', 'Admin', '', 'admin', ?, ?),
			(2, 'editor', 'editor@example.com', 'Editor', '', 'editor', ?, ?),
			(3, 'late-admin', 'late@example.com', 'Late', '', 'admin', ?, ?)
	`, now, now, now, now, now, now); err != nil {
		t.Fatalf("seed users: %v", err)
	}

	migrateToVersion(t, ctx, db, 8)

	st := store.New(db)
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(default): %v", err)
	}

	tests := []struct {
		login    string
		wantRole string
	}{
		{login: "admin", wantRole: store.OrgRoleOwner},
		{login: "editor", wantRole: store.OrgRoleMember},
		{login: "late-admin", wantRole: store.OrgRoleMember},
	}

	for _, tt := range tests {
		t.Run(tt.login, func(t *testing.T) {
			var userID int64
			if err = db.QueryRowContext(ctx, `SELECT id FROM users WHERE login = ?`, tt.login).Scan(&userID); err != nil {
				t.Fatalf("load user: %v", err)
			}
			role, ok, memberErr := st.OrganizationMemberRole(ctx, defaultOrg.ID, userID)
			if memberErr != nil || !ok {
				t.Fatalf("OrganizationMemberRole() = %q, %v, %v", role, ok, memberErr)
			}
			if role != tt.wantRole {
				t.Fatalf("role = %q, want %q", role, tt.wantRole)
			}
		})
	}

	count, err := st.CountUserOrganizations(ctx, 0)
	if err != nil {
		t.Fatalf("CountUserOrganizations(0): %v", err)
	}
	if count != 0 {
		t.Fatalf("CountUserOrganizations(0) = %d, want 0", count)
	}
}

func migrateToVersion(t *testing.T, ctx context.Context, db *sql.DB, version int64) {
	t.Helper()

	goose.SetBaseFS(migrations.Files)
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("goose dialect: %v", err)
	}
	if err := goose.UpToContext(ctx, db, ".", version); err != nil {
		t.Fatalf("goose up to %d: %v", version, err)
	}
}
