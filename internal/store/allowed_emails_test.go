package store_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
)

func TestResolveLoginRole(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)

	t.Run("bootstrap admin when whitelist empty", func(t *testing.T) {
		role, err := st.ResolveLoginRole(ctx, "Admin@Example.com", "admin@example.com")
		if err != nil {
			t.Fatalf("ResolveLoginRole() error = %v", err)
		}
		if role != auth.RoleAdmin {
			t.Errorf("role = %q, want admin", role)
		}
	})

	t.Run("refuse when not whitelisted", func(t *testing.T) {
		orgCtx := defaultOrgCtx(ctx, st)
		if err := st.InsertAllowedEmail(orgCtx, "allowed@example.com", auth.RoleReader); err != nil {
			t.Fatalf("InsertAllowedEmail(): %v", err)
		}
		_, err := st.ResolveLoginRole(ctx, "other@example.com", "")
		if !errors.Is(err, store.ErrEmailNotAllowed) {
			t.Fatalf("error = %v, want ErrEmailNotAllowed", err)
		}
	})

	t.Run("allow whitelisted email", func(t *testing.T) {
		role, err := st.ResolveLoginRole(ctx, "allowed@example.com", "")
		if err != nil {
			t.Fatalf("ResolveLoginRole() error = %v", err)
		}
		if role != auth.RoleReader {
			t.Errorf("role = %q, want reader", role)
		}
	})
}

func TestListAndDeleteAllowedEmail(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)
	ctx = defaultOrgCtx(ctx, st)

	if err := st.InsertAllowedEmail(ctx, "a@example.com", auth.RoleEditor); err != nil {
		t.Fatalf("InsertAllowedEmail(): %v", err)
	}
	if err := st.InsertAllowedEmail(ctx, "b@example.com", auth.RoleReader); err != nil {
		t.Fatalf("InsertAllowedEmail(): %v", err)
	}

	emails, err := st.ListAllowedEmails(ctx)
	if err != nil {
		t.Fatalf("ListAllowedEmails(): %v", err)
	}
	if len(emails) != 2 {
		t.Fatalf("len(emails) = %d, want 2", len(emails))
	}

	if err := st.DeleteAllowedEmail(ctx, "a@example.com"); err != nil {
		t.Fatalf("DeleteAllowedEmail(): %v", err)
	}
	if err := st.DeleteAllowedEmail(ctx, "missing@example.com"); !errors.Is(err, store.ErrAllowedEmailNotFound) {
		t.Fatalf("DeleteAllowedEmail missing error = %v", err)
	}
}

func openMemoryDB(t *testing.T) *sql.DB {
	t.Helper()

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
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("foreign_keys: %v", err)
	}
	if migrateErr := store.Migrate(ctx, db); migrateErr != nil {
		t.Fatalf("Migrate(): %v", migrateErr)
	}
	return db
}

func defaultOrgCtx(ctx context.Context, st *store.Store) context.Context {
	org, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		panic("default organization: " + err.Error())
	}
	return orgctx.WithOrganizationID(ctx, org.ID)
}
