package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/crypto"
	"github.com/jeb-maker/revues/internal/integrations/jira"
	"github.com/jeb-maker/revues/internal/store"
)

func TestAdminJira_ReaderForbidden(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodGet, "/admin/integrations/jira", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestAdminJira_SaveAndTestCloud(t *testing.T) {
	jiraSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/3/myself" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"accountId": "abc"})
	}))
	t.Cleanup(jiraSrv.Close)

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
	csrf := auth.CSRFToken(token, secret)

	saveForm := url.Values{}
	saveForm.Set("csrf_token", csrf)
	saveForm.Set("action", "save")
	saveForm.Set("instance_type", jira.InstanceCloud)
	saveForm.Set("base_url", jiraSrv.URL)
	saveForm.Set("email", "admin@example.com")
	saveForm.Set("api_token", "cloud-token")
	saveReq := httptest.NewRequest(http.MethodPost, "/admin/integrations/jira", strings.NewReader(saveForm.Encode()))
	saveReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	saveReq.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	saveRec := httptest.NewRecorder()
	handler.ServeHTTP(saveRec, saveReq)
	if saveRec.Code != http.StatusSeeOther {
		t.Fatalf("save status = %d, want %d, body=%q", saveRec.Code, http.StatusSeeOther, saveRec.Body.String())
	}

	key, err := crypto.DecodeKey(config.TestEncryptionKey())
	if err != nil {
		t.Fatalf("DecodeKey(): %v", err)
	}
	svc := &jira.Service{Store: st, EncryptionKey: key}
	cfg, ok, err := svc.Load(ctx)
	if err != nil || !ok {
		t.Fatalf("Load() = ok=%v err=%v", ok, err)
	}
	if cfg.InstanceType != jira.InstanceCloud || cfg.Email != "admin@example.com" || cfg.APIToken != "cloud-token" {
		t.Fatalf("Load() = %+v", cfg)
	}

	testForm := url.Values{}
	testForm.Set("csrf_token", csrf)
	testForm.Set("action", "test")
	testReq := httptest.NewRequest(http.MethodPost, "/admin/integrations/jira", strings.NewReader(testForm.Encode()))
	testReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	testReq.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	testRec := httptest.NewRecorder()
	handler.ServeHTTP(testRec, testReq)
	if testRec.Code != http.StatusSeeOther {
		t.Fatalf("test status = %d, want %d, body=%q", testRec.Code, http.StatusSeeOther, testRec.Body.String())
	}
}

func TestAdminJira_SaveServer(t *testing.T) {
	handler, db := testRouterWithEncryptionKey(t, config.TestEncryptionKey())
	ctx := context.Background()
	st := store.New(db)

	adminUser, err := st.UpsertGitHubUser(ctx, 100, "admin2", "admin2@example.com", "Admin", "", auth.RoleAdmin)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	secret := "test-secret-at-least-thirty-two-bytes"
	sessions := &auth.SessionManager{Store: st, SessionSecret: secret}
	token, _, err := sessions.CreateLoginSession(ctx, adminUser.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}
	csrf := auth.CSRFToken(token, secret)

	saveForm := url.Values{}
	saveForm.Set("csrf_token", csrf)
	saveForm.Set("action", "save")
	saveForm.Set("instance_type", jira.InstanceServer)
	saveForm.Set("base_url", "https://jira.example.com")
	saveForm.Set("pat", "server-pat-value")
	saveReq := httptest.NewRequest(http.MethodPost, "/admin/integrations/jira", strings.NewReader(saveForm.Encode()))
	saveReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	saveReq.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	saveRec := httptest.NewRecorder()
	handler.ServeHTTP(saveRec, saveReq)
	if saveRec.Code != http.StatusSeeOther {
		t.Fatalf("save status = %d, want %d, body=%q", saveRec.Code, http.StatusSeeOther, saveRec.Body.String())
	}

	key, err := crypto.DecodeKey(config.TestEncryptionKey())
	if err != nil {
		t.Fatalf("DecodeKey(): %v", err)
	}
	svc := &jira.Service{Store: st, EncryptionKey: key}
	cfg, ok, err := svc.Load(ctx)
	if err != nil || !ok {
		t.Fatalf("Load() = ok=%v err=%v", ok, err)
	}
	if cfg.InstanceType != jira.InstanceServer || cfg.PAT != "server-pat-value" {
		t.Fatalf("Load() = %+v", cfg)
	}
}
