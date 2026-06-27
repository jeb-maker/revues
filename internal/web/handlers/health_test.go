package handlers_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/store"
	appweb "github.com/jeb-maker/revues/internal/web"
)

func testRouter(t *testing.T) (http.Handler, *sql.DB) {
	t.Helper()

	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/test.db")
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

	cfg := config.Config{
		Addr:          ":8080",
		BaseURL:       "http://example.com",
		SessionSecret: "test-secret-at-least-thirty-two-bytes",
		Env:           "development",
	}

	handler, err := appweb.NewRouter(appweb.Deps{Config: cfg, DB: db})
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	return handler, db
}

func TestHealthz(t *testing.T) {
	t.Parallel()

	handler, err := appweb.NewRouter(appweb.Deps{
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

func TestLoginPage(t *testing.T) {
	handler, _ := testRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "Se connecter avec GitHub") {
		t.Fatalf("expected login page content")
	}
}

func TestCSRF_MissingToken(t *testing.T) {
	handler, db := testRouter(t)
	st := store.New(db)
	ctx := context.Background()

	user, err := st.UpsertGitHubUser(ctx, 1, "alice", "alice@example.com", "Alice", "", "reader")
	if err != nil {
		t.Fatalf("UpsertGitHubUser() error = %v", err)
	}

	sessions := &auth.SessionManager{
		Store:         st,
		SessionSecret: "test-secret-at-least-thirty-two-bytes",
	}
	token, _, err := sessions.CreateLoginSession(ctx, user.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestUnauthenticatedLogoutForbiddenWithoutSession(t *testing.T) {
	handler, _ := testRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/logout", strings.NewReader("csrf_token=invalid"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestSession_Fixation(t *testing.T) {
	ctx := context.Background()
	db := mustMemoryDB(t)
	st := store.New(db)

	user, err := st.UpsertGitHubUser(ctx, 42, "bob", "bob@example.com", "Bob", "", "reader")
	if err != nil {
		t.Fatalf("UpsertGitHubUser() error = %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}

	oldToken, _, err := sessions.CreateLoginSession(ctx, user.ID)
	if err != nil {
		t.Fatalf("old session: %v", err)
	}

	newToken, _, err := sessions.CreateLoginSession(ctx, user.ID)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}

	if _, err := st.UserIDByTokenHash(ctx, auth.HashToken(oldToken)); err == nil {
		t.Fatal("old session token should be invalid after rotation")
	}

	if _, err := st.UserIDByTokenHash(ctx, auth.HashToken(newToken)); err != nil {
		t.Fatalf("new session should be valid: %v", err)
	}
}

func TestLogoutWithCSRF(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)

	user, err := st.UpsertGitHubUser(ctx, 7, "carol", "carol@example.com", "Carol", "", "reader")
	if err != nil {
		t.Fatalf("UpsertGitHubUser() error = %v", err)
	}

	secret := "test-secret-at-least-thirty-two-bytes"
	sessions := &auth.SessionManager{Store: st, SessionSecret: secret}
	token, _, err := sessions.CreateLoginSession(ctx, user.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession() error = %v", err)
	}

	form := url.Values{}
	form.Set("csrf_token", auth.CSRFToken(token, secret))
	req := httptest.NewRequest(http.MethodPost, "/logout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}

	if _, err := st.UserIDByTokenHash(ctx, auth.HashToken(token)); err == nil {
		t.Fatal("session should be deleted after logout")
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
