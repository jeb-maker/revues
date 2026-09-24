package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

func TestCreateLocalUser_AndCredentials(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/test.db", 0)
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err = store.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate(): %v", err)
	}
	st := store.New(db)

	hash, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword(): %v", err)
	}

	user, err := st.CreateLocalUser(ctx, "Local@Example.com", "local", "Local User", auth.RoleEditor, hash)
	if err != nil {
		t.Fatalf("CreateLocalUser(): %v", err)
	}
	if user.GitHubID != 0 {
		t.Fatalf("GitHubID = %d, want 0", user.GitHubID)
	}

	got, passwordHash, err := st.UserCredentialsByEmail(ctx, "local@example.com")
	if err != nil {
		t.Fatalf("UserCredentialsByEmail(): %v", err)
	}
	if got.ID != user.ID {
		t.Fatalf("id = %d, want %d", got.ID, user.ID)
	}
	if passwordHash != hash {
		t.Fatal("password hash mismatch")
	}

	_, err = st.CreateLocalUser(ctx, "local@example.com", "local2", "Dup", auth.RoleReader, hash)
	if !errors.Is(err, store.ErrEmailTaken) {
		t.Fatalf("duplicate email err = %v, want ErrEmailTaken", err)
	}
}

func TestUpsertGitHubUser_StillWorks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/test.db", 0)
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err = store.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate(): %v", err)
	}
	st := store.New(db)

	user, err := st.UpsertGitHubUser(ctx, 42, "octocat", "octo@example.com", "Octo", "https://avatar", auth.RoleAdmin)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	if user.GitHubID != 42 {
		t.Fatalf("GitHubID = %d", user.GitHubID)
	}

	again, err := st.UpsertGitHubUser(ctx, 42, "octocat", "octo@example.com", "Octocat", "https://avatar2", auth.RoleAdmin)
	if err != nil {
		t.Fatalf("UpsertGitHubUser update: %v", err)
	}
	if again.ID != user.ID {
		t.Fatalf("id changed on upsert")
	}
	if again.DisplayName != "Octocat" {
		t.Fatalf("DisplayName = %q", again.DisplayName)
	}
}
