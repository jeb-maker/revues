package web_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/store"
	appweb "github.com/jeb-maker/revues/internal/web"
)

func TestHealthz(t *testing.T) {
	t.Parallel()

	handler, _, err := appweb.NewRouter(appweb.Deps{
		Config: config.Config{SessionSecret: "test-secret-at-least-thirty-two-bytes"},
		DB:     mustMemoryDB(t),
	})
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "ok" {
		t.Errorf("body = %q, want %q", rec.Body.String(), "ok")
	}
}

func mustMemoryDB(t *testing.T) *sql.DB {
	t.Helper()

	ctx := context.Background()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Open memory db: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("foreign_keys: %v", err)
	}
	if err := store.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	return db
}
