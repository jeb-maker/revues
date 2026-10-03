package web_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/store"
	appweb "github.com/jeb-maker/revues/internal/web"
)

func TestHealthz(t *testing.T) {
	t.Parallel()

	handler, _, _, err := appweb.NewRouter(appweb.Deps{
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

func TestAPIV1Health(t *testing.T) {
	t.Parallel()

	handler, _, _, err := appweb.NewRouter(appweb.Deps{
		Config: config.Config{SessionSecret: "test-secret-at-least-thirty-two-bytes"},
		DB:     mustMemoryDB(t),
	})
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := strings.TrimSpace(rec.Body.String())
	if !strings.Contains(body, `"status":"ok"`) {
		t.Fatalf("body = %q, want JSON status ok", body)
	}
}

func TestSPAStubWhenBuildMissing(t *testing.T) {
	t.Setenv("REVUES_SPA_DIR", t.TempDir()+"/missing-spa")

	handler, _, _, err := appweb.NewRouter(appweb.Deps{
		Config: config.Config{SessionSecret: "test-secret-at-least-thirty-two-bytes"},
		DB:     mustMemoryDB(t),
	})
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d (stub must not look healthy)", rec.Code, http.StatusServiceUnavailable)
	}
	body := rec.Body.String()
	for _, part := range []string{"SPA SvelteKit", "npm run build", "/healthz"} {
		if !strings.Contains(body, part) {
			t.Fatalf("stub body missing %q: %q", part, body)
		}
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
