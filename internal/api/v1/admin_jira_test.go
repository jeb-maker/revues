package apiv1_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/testutil"
)

func TestAdminJira_MaskedTokenTestAndRBAC(t *testing.T) {
	t.Parallel()

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/rest/api/3/myself" {
			authz := r.Header.Get("Authorization")
			if !strings.HasPrefix(authz, "Basic ") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"accountId":"abc"}`))
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(mock.Close)

	cfg := config.Config{
		SessionSecret:  "test-secret-at-least-thirty-two-bytes",
		EncryptionKey:  testutil.EncryptionKey(),
		Env:            "development",
		AttachmentsDir: t.TempDir(),
		BaseURL:        "http://localhost:8080",
	}
	handler, st := newTestRouterWithStore(t, cfg)

	admin, session, csrf := seedSessionUser(t, st, cfg, "jira-admin@example.com", "Jira Admin", auth.RoleAdmin, true)
	_ = admin
	_, memberSession, memberCSRF := seedSessionUser(t, st, cfg, "jira-member@example.com", "Member", auth.RoleEditor, true)

	denied := doJSON(t, handler, http.MethodGet, "/api/v1/admin/integrations/jira", nil, memberSession, "")
	if denied.Code != http.StatusForbidden {
		t.Fatalf("member get jira status = %d", denied.Code)
	}
	deniedPut := doJSON(t, handler, http.MethodPut, "/api/v1/admin/integrations/jira", map[string]any{
		"base_url": mock.URL, "email": "bot@example.com", "api_token": "secret-token",
	}, memberSession, memberCSRF)
	if deniedPut.Code != http.StatusForbidden {
		t.Fatalf("member put jira status = %d", deniedPut.Code)
	}

	noCSRF := doJSON(t, handler, http.MethodPut, "/api/v1/admin/integrations/jira", map[string]any{
		"base_url": mock.URL, "email": "bot@example.com", "api_token": "secret-token",
	}, session, "")
	if noCSRF.Code != http.StatusForbidden {
		t.Fatalf("no csrf status = %d", noCSRF.Code)
	}

	put := doJSON(t, handler, http.MethodPut, "/api/v1/admin/integrations/jira", map[string]any{
		"base_url":    mock.URL,
		"email":       "bot@example.com",
		"api_token":   "super-secret-jira-token",
		"project_key": "rev",
		"issue_type":  "Bug",
	}, session, csrf)
	if put.Code != http.StatusOK {
		t.Fatalf("put jira status = %d body=%s", put.Code, put.Body.String())
	}
	body := put.Body.String()
	if strings.Contains(body, "super-secret-jira-token") {
		t.Fatalf("api token leaked: %s", body)
	}
	var settings map[string]any
	if err := json.Unmarshal(put.Body.Bytes(), &settings); err != nil {
		t.Fatalf("json: %v", err)
	}
	if settings["has_api_token"] != true || settings["configured"] != true {
		t.Fatalf("settings = %#v", settings)
	}
	if settings["project_key"] != "REV" {
		t.Fatalf("project_key = %v want REV", settings["project_key"])
	}
	if settings["issue_type"] != "Bug" {
		t.Fatalf("issue_type = %v", settings["issue_type"])
	}
	if _, ok := settings["api_token"]; ok {
		t.Fatal("api_token must be absent from JSON")
	}

	// Blank token keeps existing
	putKeep := doJSON(t, handler, http.MethodPut, "/api/v1/admin/integrations/jira", map[string]any{
		"base_url": mock.URL, "email": "bot@example.com", "project_key": "REV",
	}, session, csrf)
	if putKeep.Code != http.StatusOK {
		t.Fatalf("put keep status = %d body=%s", putKeep.Code, putKeep.Body.String())
	}
	var keep map[string]any
	_ = json.Unmarshal(putKeep.Body.Bytes(), &keep)
	if keep["has_api_token"] != true {
		t.Fatal("expected api token retained")
	}

	testOK := doJSON(t, handler, http.MethodPost, "/api/v1/admin/integrations/jira/test", nil, session, csrf)
	if testOK.Code != http.StatusNoContent {
		t.Fatalf("test status = %d body=%s", testOK.Code, testOK.Body.String())
	}

	hub := doJSON(t, handler, http.MethodGet, "/api/v1/admin/integrations", nil, session, "")
	if hub.Code != http.StatusOK {
		t.Fatalf("hub status = %d", hub.Code)
	}
	var overview map[string]any
	_ = json.Unmarshal(hub.Body.Bytes(), &overview)
	found := false
	for _, raw := range overview["items"].([]any) {
		item := raw.(map[string]any)
		if item["key"] == "jira" {
			found = true
			if item["enabled"] != true {
				t.Fatalf("jira enabled = %v", item["enabled"])
			}
			if item["config_path"] != "/admin/integrations/jira" {
				t.Fatalf("config_path = %v", item["config_path"])
			}
		}
	}
	if !found {
		t.Fatal("jira missing from hub")
	}

	del := doJSON(t, handler, http.MethodDelete, "/api/v1/admin/integrations/jira", nil, session, csrf)
	if del.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d body=%s", del.Code, del.Body.String())
	}
	after := doJSON(t, handler, http.MethodGet, "/api/v1/admin/integrations/jira", nil, session, "")
	if after.Code != http.StatusOK {
		t.Fatalf("get after delete = %d", after.Code)
	}
	var cleared map[string]any
	_ = json.Unmarshal(after.Body.Bytes(), &cleared)
	if cleared["configured"] != false || cleared["has_api_token"] != false {
		t.Fatalf("cleared = %#v", cleared)
	}

	testFail := doJSON(t, handler, http.MethodPost, "/api/v1/admin/integrations/jira/test", nil, session, csrf)
	if testFail.Code != http.StatusBadRequest {
		t.Fatalf("test without config status = %d", testFail.Code)
	}
}

func TestAdminJira_RejectsPrivateURL(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		SessionSecret:  "test-secret-at-least-thirty-two-bytes",
		EncryptionKey:  testutil.EncryptionKey(),
		Env:            "development",
		AttachmentsDir: t.TempDir(),
	}
	handler, st := newTestRouterWithStore(t, cfg)
	_, session, csrf := seedSessionUser(t, st, cfg, "jira-ssrf@example.com", "Admin", auth.RoleAdmin, true)

	put := doJSON(t, handler, http.MethodPut, "/api/v1/admin/integrations/jira", map[string]any{
		"base_url": "https://10.0.0.8", "email": "bot@example.com", "api_token": "tok",
	}, session, csrf)
	if put.Code != http.StatusBadRequest {
		t.Fatalf("private url status = %d body=%s", put.Code, put.Body.String())
	}
	if !strings.Contains(put.Body.String(), "non autorisée") {
		t.Fatalf("expected SSRF message, got %s", put.Body.String())
	}
}

func TestRunItemJira_LinkAndCreate(t *testing.T) {
	t.Parallel()

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/rest/api/3/issue/"):
			key := strings.TrimPrefix(r.URL.Path, "/rest/api/3/issue/")
			if key == "REV-1" {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"key":"REV-1"}`))
				return
			}
			http.NotFound(w, r)
		case r.Method == http.MethodPost && r.URL.Path == "/rest/api/3/issue/":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"key":"REV-42"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/rest/api/3/myself":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"accountId":"abc"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(mock.Close)

	cfg := config.Config{
		SessionSecret:  "test-secret-at-least-thirty-two-bytes",
		EncryptionKey:  testutil.EncryptionKey(),
		Env:            "development",
		AttachmentsDir: t.TempDir(),
		BaseURL:        "http://localhost:8080",
	}
	handler, st := newTestRouterWithStore(t, cfg)

	admin, adminSession, adminCSRF := seedSessionUser(t, st, cfg, "jira-link-admin@example.com", "Admin", auth.RoleAdmin, true)
	_ = admin
	editor, session, csrf := seedSessionUser(t, st, cfg, "jira-link-editor@example.com", "Editor", auth.RoleEditor, true)
	_ = editor

	putCfg := doJSON(t, handler, http.MethodPut, "/api/v1/admin/integrations/jira", map[string]any{
		"base_url": mock.URL, "email": "bot@example.com", "api_token": "tok",
		"project_key": "REV", "issue_type": "Task",
	}, adminSession, adminCSRF)
	if putCfg.Code != http.StatusOK {
		t.Fatalf("put jira cfg status = %d body=%s", putCfg.Code, putCfg.Body.String())
	}

	runID, itemID := seedRunForAttachments(t, handler, session, csrf)

	// Gate subject membership so org_member_legacy does not grant IDOR access.
	detailGate := doJSON(t, handler, http.MethodGet, fmt.Sprintf("/api/v1/runs/%d", runID), nil, session, "")
	if detailGate.Code != http.StatusOK {
		t.Fatalf("run detail: %d %s", detailGate.Code, detailGate.Body.String())
	}
	var runGate map[string]any
	_ = json.Unmarshal(detailGate.Body.Bytes(), &runGate)
	subjectID := int64(runGate["subject_id"].(float64))
	_ = doJSON(t, handler, http.MethodPost,
		fmt.Sprintf("/api/v1/subjects/%d/members", subjectID),
		map[string]any{"email": editor.Email, "role": "lead"},
		session, csrf)

	// Mark nok
	nok := doJSON(t, handler, http.MethodPatch,
		fmt.Sprintf("/api/v1/runs/%d/items/%d", runID, itemID),
		map[string]any{"status": "nok", "comment": "Échec contrôle"},
		session, csrf)
	if nok.Code != http.StatusOK {
		t.Fatalf("nok status = %d body=%s", nok.Code, nok.Body.String())
	}

	status := doJSON(t, handler, http.MethodGet,
		fmt.Sprintf("/api/v1/runs/%d/items/%d/jira", runID, itemID), nil, session, "")
	if status.Code != http.StatusOK {
		t.Fatalf("get jira status = %d body=%s", status.Code, status.Body.String())
	}
	var jiraStatus map[string]any
	_ = json.Unmarshal(status.Body.Bytes(), &jiraStatus)
	if jiraStatus["configured"] != true || jiraStatus["can_link"] != true || jiraStatus["can_create"] != true {
		t.Fatalf("jira status = %#v", jiraStatus)
	}

	link := doJSON(t, handler, http.MethodPut,
		fmt.Sprintf("/api/v1/runs/%d/items/%d/jira", runID, itemID),
		map[string]any{"issue": "REV-1"},
		session, csrf)
	if link.Code != http.StatusOK {
		t.Fatalf("link status = %d body=%s", link.Code, link.Body.String())
	}
	var linked map[string]any
	_ = json.Unmarshal(link.Body.Bytes(), &linked)
	if linked["external_key"] != "REV-1" {
		t.Fatalf("linked = %#v", linked)
	}
	if !strings.Contains(fmt.Sprint(linked["external_url"]), "/browse/REV-1") {
		t.Fatalf("external_url = %v", linked["external_url"])
	}

	detail := doJSON(t, handler, http.MethodGet,
		fmt.Sprintf("/api/v1/runs/%d/items/%d", runID, itemID), nil, session, "")
	if detail.Code != http.StatusOK || !strings.Contains(detail.Body.String(), `"REV-1"`) {
		t.Fatalf("detail missing jira link: %s", detail.Body.String())
	}

	// Second item for create (no link yet)
	tplRec := doJSON(t, handler, http.MethodPost, "/api/v1/templates", map[string]any{
		"name": "JiraTpl2", "domains": []string{"ops"},
		"items": []map[string]any{{"section": "S", "label": "Autre point", "required": true}},
	}, session, csrf)
	var tpl map[string]any
	_ = json.Unmarshal(tplRec.Body.Bytes(), &tpl)
	subRec := doJSON(t, handler, http.MethodPost, "/api/v1/subjects", map[string]any{
		"name": "Jira Sub 2", "domains": []string{"ops"},
	}, session, csrf)
	var sub map[string]any
	_ = json.Unmarshal(subRec.Body.Bytes(), &sub)
	launch := doJSON(t, handler, http.MethodPost,
		fmt.Sprintf("/api/v1/subjects/%d/runs", int64(sub["id"].(float64))),
		map[string]any{"template_id": int64(tpl["id"].(float64))}, session, csrf)
	var run2 map[string]any
	_ = json.Unmarshal(launch.Body.Bytes(), &run2)
	run2ID := int64(run2["id"].(float64))
	item2ID := int64(run2["items"].([]any)[0].(map[string]any)["id"].(float64))

	nok2 := doJSON(t, handler, http.MethodPatch,
		fmt.Sprintf("/api/v1/runs/%d/items/%d", run2ID, item2ID),
		map[string]any{"status": "nok", "comment": "Fail"},
		session, csrf)
	if nok2.Code != http.StatusOK {
		t.Fatalf("nok2 status = %d body=%s", nok2.Code, nok2.Body.String())
	}

	create := doJSON(t, handler, http.MethodPost,
		fmt.Sprintf("/api/v1/runs/%d/items/%d/jira", run2ID, item2ID),
		map[string]any{"title": "Ticket auto", "description": "Détails"},
		session, csrf)
	if create.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", create.Code, create.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(create.Body.Bytes(), &created)
	if created["external_key"] != "REV-42" {
		t.Fatalf("created = %#v", created)
	}

	dup := doJSON(t, handler, http.MethodPost,
		fmt.Sprintf("/api/v1/runs/%d/items/%d/jira", run2ID, item2ID),
		map[string]any{}, session, csrf)
	if dup.Code != http.StatusConflict {
		t.Fatalf("dup create status = %d want 409 body=%s", dup.Code, dup.Body.String())
	}

	// IDOR: bob cannot link
	_, bobSession, bobCSRF := seedSessionUser(t, st, cfg, "jira-bob@example.com", "Bob", auth.RoleEditor, true)
	denied := doJSON(t, handler, http.MethodPut,
		fmt.Sprintf("/api/v1/runs/%d/items/%d/jira", runID, itemID),
		map[string]any{"issue": "REV-1"}, bobSession, bobCSRF)
	if denied.Code != http.StatusNotFound {
		t.Fatalf("bob link status = %d want 404 body=%s", denied.Code, denied.Body.String())
	}
}
