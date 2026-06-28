package store_test

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/jeb-maker/revues/internal/store"
)

func TestMigrateUpDown(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("foreign_keys pragma: %v", err)
	}

	if err := store.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	if err := assertCoreTables(t, ctx, db); err != nil {
		t.Fatal(err)
	}

	if err := store.Down(ctx, db); err != nil {
		t.Fatalf("Down() error = %v", err)
	}

	if err := assertTableMissing(t, ctx, db, "users"); err != nil {
		t.Fatal(err)
	}
}

func TestOpenFileCreatesDirectory(t *testing.T) {
	ctx := context.Background()
	path := t.TempDir() + "/nested/revues.db"

	db, err := store.Open(ctx, path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	if err := store.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
}

func assertCoreTables(t *testing.T, ctx context.Context, db *sql.DB) error {
	t.Helper()

	tables := []string{
		"users",
		"sessions",
		"allowed_emails",
		"projects",
		"project_members",
		"checklist_templates",
		"template_versions",
		"template_items",
		"checklist_runs",
		"run_items",
		"run_item_events",
		"settings",
	}

	for _, table := range tables {
		var name string
		err := db.QueryRowContext(ctx, `
			SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?
		`, table).Scan(&name)
		if err != nil {
			return err
		}
	}

	return nil
}

func assertTableMissing(t *testing.T, ctx context.Context, db *sql.DB, table string) error {
	t.Helper()

	var name string
	err := db.QueryRowContext(ctx, `
		SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?
	`, table).Scan(&name)
	if err == nil {
		t.Fatalf("table %q still exists after down", table)
	}
	if err != sql.ErrNoRows {
		return err
	}

	return nil
}
