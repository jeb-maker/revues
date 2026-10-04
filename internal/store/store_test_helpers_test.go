package store_test

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
)

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
