package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	authhandler "github.com/jeb-maker/revues/internal/features/auth"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
)

func TestDevLogin_SwitchesUser(t *testing.T) {
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
	ctx = testutil.DefaultOrgContext(ctx, st)

	admin, err := st.UpsertGitHubUser(ctx, 1, "admin", "admin@example.com", "Admin", "", auth.RoleAdmin)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(admin): %v", err)
	}
	alice, err := st.UpsertGitHubUser(ctx, 2, "alice", "alice@example.com", "Alice", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(alice): %v", err)
	}
	org, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, org.ID, admin.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(admin): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, org.ID, alice.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(alice): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	h := &authhandler.Auth{
		Store:    st,
		Sessions: sessions,
		Config:   config.Config{Env: "development", DevAuth: true},
	}

	form := url.Values{"user_id": {strconv.FormatInt(alice.ID, 10)}}
	req := httptest.NewRequest(http.MethodPost, "/auth/dev/login", strings.NewReader(form.Encode()))
	req.Host = "127.0.0.1:8080"
	req.RemoteAddr = "127.0.0.1:54321"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.DevLogin(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var cookie string
	for _, c := range rec.Result().Cookies() {
		if c.Name == "revues_session" {
			cookie = c.Value
		}
	}
	if cookie == "" {
		t.Fatal("expected session cookie")
	}
	userID, err := st.UserIDByTokenHash(ctx, auth.HashToken(cookie))
	if err != nil {
		t.Fatalf("UserIDByTokenHash(): %v", err)
	}
	if userID != alice.ID {
		t.Fatalf("session user = %d, want %d", userID, alice.ID)
	}
}

func TestDevLogin_Disabled(t *testing.T) {
	t.Parallel()

	h := &authhandler.Auth{
		Config: config.Config{Env: "development", DevAuth: false},
	}
	req := httptest.NewRequest(http.MethodPost, "/auth/dev/login", strings.NewReader("user_id=1"))
	req.Host = "127.0.0.1:8080"
	req.RemoteAddr = "127.0.0.1:54321"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.DevLogin(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestDevLogin_RejectsNonLoopback(t *testing.T) {
	t.Parallel()

	h := &authhandler.Auth{
		Config: config.Config{Env: "development", DevAuth: true},
	}
	req := httptest.NewRequest(http.MethodPost, "/auth/dev/login", strings.NewReader("user_id=1"))
	req.Host = "example.com"
	req.RemoteAddr = "203.0.113.10:54321"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.DevLogin(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestDevLogin_ProductionNever(t *testing.T) {
	t.Parallel()

	h := &authhandler.Auth{
		Config: config.Config{Env: "production", DevAuth: true},
	}
	req := httptest.NewRequest(http.MethodPost, "/auth/dev/login", strings.NewReader("user_id=1"))
	req.Host = "127.0.0.1:8080"
	req.RemoteAddr = "127.0.0.1:54321"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.DevLogin(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
