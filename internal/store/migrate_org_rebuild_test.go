package store_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/migrations"
	"github.com/pressly/goose/v3"
)

// 00006 rebuilds organizations. With real runs, DROP TABLE organizations fails
// inside a goose transaction (PRAGMA foreign_keys=OFF is ignored; error 1811
// from subjects←checklist_runs ON DELETE RESTRICT). NO TRANSACTION is required.
func TestMigrate6OrganizationsRebuildWithRuns(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "revues.db"), 10)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	goose.SetBaseFS(migrations.Files)
	err = goose.SetDialect("sqlite3")
	if err != nil {
		t.Fatal(err)
	}
	err = goose.UpToContext(ctx, db, ".", 5)
	if err != nil {
		t.Fatalf("up to 5: %v", err)
	}

	var oid int64
	err = db.QueryRow(`SELECT id FROM organizations LIMIT 1`).Scan(&oid)
	if err != nil {
		t.Fatalf("org: %v", err)
	}
	res, err := db.Exec(`INSERT INTO subjects (organization_id, name, description, visibility, created_at, updated_at) VALUES (?, 'S', '', 'normal', datetime('now'), datetime('now'))`, oid)
	if err != nil {
		t.Fatalf("subject: %v", err)
	}
	sid, _ := res.LastInsertId()
	res, err = db.Exec(`INSERT INTO checklist_templates (organization_id, name, created_at) VALUES (?, 'T', datetime('now'))`, oid)
	if err != nil {
		t.Fatalf("template: %v", err)
	}
	tid, _ := res.LastInsertId()
	res, err = db.Exec(`INSERT INTO template_versions (template_id, version, published_at) VALUES (?, 1, datetime('now'))`, tid)
	if err != nil {
		t.Fatalf("version: %v", err)
	}
	vid, _ := res.LastInsertId()
	_, err = db.Exec(`INSERT INTO checklist_runs (subject_id, template_version_id, status, created_at) VALUES (?, ?, 'draft', datetime('now'))`, sid, vid)
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = tx.ExecContext(ctx, `PRAGMA foreign_keys=OFF`)
	_, dropErr := tx.ExecContext(ctx, `DROP TABLE organizations`)
	_ = tx.Rollback()
	if dropErr == nil {
		t.Fatal("expected DROP inside TX to fail when runs exist")
	}
	if !strings.Contains(dropErr.Error(), "constraint") {
		t.Fatalf("unexpected DROP err: %v", dropErr)
	}

	err = goose.UpToContext(ctx, db, ".", 6)
	if err != nil {
		t.Fatalf("up to 6: %v", err)
	}

	var runs int
	err = db.QueryRow(`SELECT COUNT(*) FROM checklist_runs`).Scan(&runs)
	if err != nil {
		t.Fatal(err)
	}
	if runs != 1 {
		t.Fatalf("runs = %d, want 1", runs)
	}
}
