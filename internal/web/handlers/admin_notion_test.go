package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/crypto"
	"github.com/jeb-maker/revues/internal/integrations/notion"
	"github.com/jeb-maker/revues/internal/store"
)

func TestAdminNotion_ReaderForbidden(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	reader, err := st.UpsertGitHubUser(ctx, 1, "reader", "reader@example.com", "Reader", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, reader.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/admin/integrations/notion", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestAdminNotion_Save(t *testing.T) {
	handler, db := testRouterWithEncryptionKey(t, config.TestEncryptionKey())
	ctx := context.Background()
	st := store.New(db)
	adminUser, err := st.UpsertGitHubUser(ctx, 99, "admin", "admin@example.com", "Admin", "", auth.RoleAdmin)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	secret := "test-secret-at-least-thirty-two-bytes"
	sessions := &auth.SessionManager{Store: st, SessionSecret: secret}
	token, _, err := sessions.CreateLoginSession(ctx, adminUser.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}
	form := url.Values{}
	form.Set("csrf_token", auth.CSRFToken(token, secret))
	form.Set("action", "save")
	form.Set("api_token", "notion-secret-token")
	req := httptest.NewRequest(http.MethodPost, "/admin/integrations/notion", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("save status = %d, body=%q", rec.Code, rec.Body.String())
	}
	key, err := crypto.DecodeKey(config.TestEncryptionKey())
	if err != nil {
		t.Fatalf("DecodeKey(): %v", err)
	}
	cfg, ok, err := (&notion.Service{Store: st, EncryptionKey: key}).Load(ctx)
	if err != nil || !ok || cfg.APIToken != "notion-secret-token" {
		t.Fatalf("Load() = %+v ok=%v err=%v", cfg, ok, err)
	}
}
