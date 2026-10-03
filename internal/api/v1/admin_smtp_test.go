package apiv1_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
)

func TestAdminSMTP_MaskedPasswordAndOrgAdmin(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		SessionSecret:  "test-secret-at-least-thirty-two-bytes",
		EncryptionKey:  testutil.EncryptionKey(),
		Env:            "development",
		AttachmentsDir: t.TempDir(),
	}
	handler, st := newTestRouterWithStore(t, cfg)

	admin, session, csrf := seedSessionUser(t, st, cfg, "smtp-admin@example.com", "SMTP Admin", auth.RoleAdmin, true)
	_ = admin

	member, memberSession, memberCSRF := seedSessionUser(t, st, cfg, "smtp-member@example.com", "Member", auth.RoleEditor, true)
	_ = member

	// Non-org-admin denied
	denied := doJSON(t, handler, http.MethodGet, "/api/v1/admin/settings/smtp", nil, memberSession, "")
	if denied.Code != http.StatusForbidden {
		t.Fatalf("member get smtp status = %d body=%s", denied.Code, denied.Body.String())
	}
	deniedPut := doJSON(t, handler, http.MethodPut, "/api/v1/admin/settings/smtp", map[string]any{
		"host": "smtp.example.com", "port": 587, "from": "revues@example.com", "password": "s3cret",
	}, memberSession, memberCSRF)
	if deniedPut.Code != http.StatusForbidden {
		t.Fatalf("member put smtp status = %d", deniedPut.Code)
	}

	// CSRF required
	noCSRF := doJSON(t, handler, http.MethodPut, "/api/v1/admin/settings/smtp", map[string]any{
		"host": "smtp.example.com", "port": 587, "from": "revues@example.com", "password": "s3cret",
	}, session, "")
	if noCSRF.Code != http.StatusForbidden {
		t.Fatalf("no csrf status = %d", noCSRF.Code)
	}

	put := doJSON(t, handler, http.MethodPut, "/api/v1/admin/settings/smtp", map[string]any{
		"host": "smtp.example.com", "port": 587, "tls": true,
		"username": "relay", "from": "revues@example.com", "password": "super-secret-password",
	}, session, csrf)
	if put.Code != http.StatusOK {
		t.Fatalf("put smtp status = %d body=%s", put.Code, put.Body.String())
	}
	body := put.Body.String()
	if strings.Contains(body, "super-secret-password") {
		t.Fatalf("password leaked in response: %s", body)
	}
	var smtp map[string]any
	if err := json.Unmarshal(put.Body.Bytes(), &smtp); err != nil {
		t.Fatalf("json: %v", err)
	}
	if smtp["has_password"] != true {
		t.Fatalf("has_password = %v", smtp["has_password"])
	}
	if smtp["host"] != "smtp.example.com" || smtp["configured"] != true || smtp["enabled"] != true {
		t.Fatalf("smtp = %#v", smtp)
	}
	if _, ok := smtp["password"]; ok {
		t.Fatal("password field must be absent from JSON")
	}

	get := doJSON(t, handler, http.MethodGet, "/api/v1/admin/settings/smtp", nil, session, "")
	if get.Code != http.StatusOK {
		t.Fatalf("get smtp status = %d", get.Code)
	}
	if strings.Contains(get.Body.String(), "super-secret-password") {
		t.Fatalf("password leaked on get: %s", get.Body.String())
	}

	// Blank password keeps existing
	putKeep := doJSON(t, handler, http.MethodPut, "/api/v1/admin/settings/smtp", map[string]any{
		"host": "smtp.example.com", "port": 587, "from": "revues@example.com", "password": "",
	}, session, csrf)
	if putKeep.Code != http.StatusOK {
		t.Fatalf("put keep status = %d body=%s", putKeep.Code, putKeep.Body.String())
	}
	var keep map[string]any
	_ = json.Unmarshal(putKeep.Body.Bytes(), &keep)
	if keep["has_password"] != true {
		t.Fatal("expected password retained")
	}

	hub := doJSON(t, handler, http.MethodGet, "/api/v1/admin/integrations", nil, session, "")
	if hub.Code != http.StatusOK {
		t.Fatalf("hub status = %d body=%s", hub.Code, hub.Body.String())
	}
	var overview map[string]any
	_ = json.Unmarshal(hub.Body.Bytes(), &overview)
	items, _ := overview["items"].([]any)
	if len(items) != 3 {
		t.Fatalf("hub items = %d, want 3 (smtp, jira, webhooks)", len(items))
	}
	foundSMTP := false
	for _, raw := range items {
		item := raw.(map[string]any)
		if item["key"] == "smtp" {
			foundSMTP = true
			if item["enabled"] != true {
				t.Fatalf("smtp enabled = %v", item["enabled"])
			}
			if item["config_path"] != "/admin/settings/smtp" {
				t.Fatalf("config_path = %v", item["config_path"])
			}
		}
	}
	if !foundSMTP {
		t.Fatal("smtp row missing from hub")
	}

	del := doJSON(t, handler, http.MethodDelete, "/api/v1/admin/settings/smtp", nil, session, csrf)
	if del.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d", del.Code)
	}
	after := doJSON(t, handler, http.MethodGet, "/api/v1/admin/settings/smtp", nil, session, "")
	var cleared map[string]any
	_ = json.Unmarshal(after.Body.Bytes(), &cleared)
	if cleared["configured"] != false {
		t.Fatalf("configured after delete = %v", cleared["configured"])
	}
}

func TestAdminIntegrations_RequiresOrgAdmin(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		SessionSecret:  "test-secret-at-least-thirty-two-bytes",
		EncryptionKey:  testutil.EncryptionKey(),
		Env:            "development",
		AttachmentsDir: t.TempDir(),
	}
	handler, st := newTestRouterWithStore(t, cfg)
	_, memberSession, _ := seedSessionUser(t, st, cfg, "hub-member@example.com", "Member", auth.RoleEditor, true)

	rec := doJSON(t, handler, http.MethodGet, "/api/v1/admin/integrations", nil, memberSession, "")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestAdminSMTP_OrgOwnerAllowed(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		SessionSecret:  "test-secret-at-least-thirty-two-bytes",
		EncryptionKey:  testutil.EncryptionKey(),
		Env:            "development",
		AttachmentsDir: t.TempDir(),
	}
	handler, st := newTestRouterWithStore(t, cfg)
	owner, _, _ := seedSessionUser(t, st, cfg, "org-owner@example.com", "Owner", auth.RoleEditor, false)
	orgID := mustDefaultOrgID(t, st)
	if err := st.AddOrganizationMember(context.Background(), orgID, owner.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember: %v", err)
	}
	// Refresh session with active org
	sessions := &auth.SessionManager{Store: st, SessionSecret: cfg.SessionSecret}
	token, csrf, err := sessions.CreateLoginSession(context.Background(), owner.ID, orgID)
	if err != nil {
		t.Fatalf("CreateLoginSession: %v", err)
	}
	session := &http.Cookie{Name: "revues_session", Value: token}

	put := doJSON(t, handler, http.MethodPut, "/api/v1/admin/settings/smtp", map[string]any{
		"host": "mail.example.com", "port": 465, "tls": true, "from": "noreply@example.com", "password": "x",
	}, session, csrf)
	if put.Code != http.StatusOK {
		t.Fatalf("owner put status = %d body=%s", put.Code, put.Body.String())
	}
	if strings.Contains(put.Body.String(), `"password"`) || strings.Contains(put.Body.String(), `"x"`) {
		// "x" alone might appear elsewhere; check password key specifically
		var body map[string]any
		_ = json.Unmarshal(put.Body.Bytes(), &body)
		if _, ok := body["password"]; ok {
			t.Fatal("password must not be present")
		}
	}
}
