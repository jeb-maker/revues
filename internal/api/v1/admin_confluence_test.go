package apiv1_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/testutil"
)

func TestAdminConfluence_MaskedTokenTestAndRBAC(t *testing.T) {
	t.Parallel()

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/wiki/rest/api/user/current" {
			authz := r.Header.Get("Authorization")
			if !strings.HasPrefix(authz, "Basic ") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"accountId":"abc","type":"known"}`))
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

	_, session, csrf := seedSessionUser(t, st, cfg, "conf-admin@example.com", "Conf Admin", auth.RoleAdmin, true)
	_, memberSession, memberCSRF := seedSessionUser(t, st, cfg, "conf-member@example.com", "Member", auth.RoleEditor, true)

	denied := doJSON(t, handler, http.MethodGet, "/api/v1/admin/integrations/confluence", nil, memberSession, "")
	if denied.Code != http.StatusForbidden {
		t.Fatalf("member get confluence status = %d", denied.Code)
	}
	deniedPut := doJSON(t, handler, http.MethodPut, "/api/v1/admin/integrations/confluence", map[string]any{
		"base_url": mock.URL, "email": "bot@example.com", "api_token": "secret-token", "space_key": "REV",
	}, memberSession, memberCSRF)
	if deniedPut.Code != http.StatusForbidden {
		t.Fatalf("member put confluence status = %d", deniedPut.Code)
	}

	noCSRF := doJSON(t, handler, http.MethodPut, "/api/v1/admin/integrations/confluence", map[string]any{
		"base_url": mock.URL, "email": "bot@example.com", "api_token": "secret-token", "space_key": "REV",
	}, session, "")
	if noCSRF.Code != http.StatusForbidden {
		t.Fatalf("no csrf status = %d", noCSRF.Code)
	}

	put := doJSON(t, handler, http.MethodPut, "/api/v1/admin/integrations/confluence", map[string]any{
		"base_url":       mock.URL,
		"email":          "bot@example.com",
		"api_token":      "super-secret-confluence-token",
		"space_key":      "rev",
		"parent_page_id": "42",
	}, session, csrf)
	if put.Code != http.StatusOK {
		t.Fatalf("put confluence status = %d body=%s", put.Code, put.Body.String())
	}
	body := put.Body.String()
	if strings.Contains(body, "super-secret-confluence-token") {
		t.Fatalf("api token leaked: %s", body)
	}
	var settings map[string]any
	if err := json.Unmarshal(put.Body.Bytes(), &settings); err != nil {
		t.Fatalf("json: %v", err)
	}
	if settings["has_api_token"] != true || settings["configured"] != true {
		t.Fatalf("settings = %#v", settings)
	}
	if settings["space_key"] != "REV" {
		t.Fatalf("space_key = %v want REV", settings["space_key"])
	}
	if settings["parent_page_id"] != "42" {
		t.Fatalf("parent_page_id = %v", settings["parent_page_id"])
	}
	if _, ok := settings["api_token"]; ok {
		t.Fatal("api_token must be absent from JSON")
	}

	putKeep := doJSON(t, handler, http.MethodPut, "/api/v1/admin/integrations/confluence", map[string]any{
		"base_url": mock.URL, "email": "bot@example.com", "space_key": "REV", "parent_page_id": "42",
	}, session, csrf)
	if putKeep.Code != http.StatusOK {
		t.Fatalf("put keep status = %d body=%s", putKeep.Code, putKeep.Body.String())
	}
	var keep map[string]any
	_ = json.Unmarshal(putKeep.Body.Bytes(), &keep)
	if keep["has_api_token"] != true {
		t.Fatal("expected api token retained")
	}

	testOK := doJSON(t, handler, http.MethodPost, "/api/v1/admin/integrations/confluence/test", nil, session, csrf)
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
		if item["key"] == "confluence" {
			found = true
			if item["enabled"] != true {
				t.Fatalf("confluence enabled = %v", item["enabled"])
			}
			if item["config_path"] != "/admin/integrations/confluence" {
				t.Fatalf("config_path = %v", item["config_path"])
			}
		}
	}
	if !found {
		t.Fatal("confluence missing from hub")
	}

	del := doJSON(t, handler, http.MethodDelete, "/api/v1/admin/integrations/confluence", nil, session, csrf)
	if del.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d body=%s", del.Code, del.Body.String())
	}
	after := doJSON(t, handler, http.MethodGet, "/api/v1/admin/integrations/confluence", nil, session, "")
	if after.Code != http.StatusOK {
		t.Fatalf("get after delete = %d", after.Code)
	}
	var cleared map[string]any
	_ = json.Unmarshal(after.Body.Bytes(), &cleared)
	if cleared["configured"] != false || cleared["has_api_token"] != false {
		t.Fatalf("cleared = %#v", cleared)
	}

	testFail := doJSON(t, handler, http.MethodPost, "/api/v1/admin/integrations/confluence/test", nil, session, csrf)
	if testFail.Code != http.StatusBadRequest {
		t.Fatalf("test without config status = %d", testFail.Code)
	}
}

func TestAdminConfluence_RejectsPrivateURL(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		SessionSecret:  "test-secret-at-least-thirty-two-bytes",
		EncryptionKey:  testutil.EncryptionKey(),
		Env:            "development",
		AttachmentsDir: t.TempDir(),
	}
	handler, st := newTestRouterWithStore(t, cfg)
	_, session, csrf := seedSessionUser(t, st, cfg, "conf-ssrf@example.com", "Admin", auth.RoleAdmin, true)

	put := doJSON(t, handler, http.MethodPut, "/api/v1/admin/integrations/confluence", map[string]any{
		"base_url": "https://10.0.0.8", "email": "bot@example.com", "api_token": "tok", "space_key": "REV",
	}, session, csrf)
	if put.Code != http.StatusBadRequest {
		t.Fatalf("private url status = %d body=%s", put.Code, put.Body.String())
	}
	if !strings.Contains(put.Body.String(), "non autorisée") {
		t.Fatalf("expected SSRF message, got %s", put.Body.String())
	}
}

func TestRunConfluence_Publish(t *testing.T) {
	t.Parallel()

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/wiki/rest/api/user/current":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"accountId":"abc","type":"known"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/wiki/rest/api/content":
			body, _ := io.ReadAll(r.Body)
			var payload map[string]any
			_ = json.Unmarshal(body, &payload)
			if payload["type"] != "page" {
				http.Error(w, "bad type", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w,
				`{"id":"777","_links":{"base":%q,"webui":"/wiki/spaces/REV/pages/777/Revue"}}`,
				"https://example.atlassian.net",
			)
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

	admin, adminSession, adminCSRF := seedSessionUser(t, st, cfg, "conf-pub-admin@example.com", "Admin", auth.RoleAdmin, true)
	_ = admin
	editor, session, csrf := seedSessionUser(t, st, cfg, "conf-pub-editor@example.com", "Editor", auth.RoleEditor, true)

	putCfg := doJSON(t, handler, http.MethodPut, "/api/v1/admin/integrations/confluence", map[string]any{
		"base_url": mock.URL, "email": "bot@example.com", "api_token": "tok", "space_key": "REV",
	}, adminSession, adminCSRF)
	if putCfg.Code != http.StatusOK {
		t.Fatalf("put confluence cfg status = %d body=%s", putCfg.Code, putCfg.Body.String())
	}

	runID, itemID := seedRunForAttachments(t, handler, session, csrf)

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

	okItem := doJSON(t, handler, http.MethodPatch,
		fmt.Sprintf("/api/v1/runs/%d/items/%d", runID, itemID),
		map[string]any{"status": "ok", "comment": "OK"},
		session, csrf)
	if okItem.Code != http.StatusOK {
		t.Fatalf("ok status = %d body=%s", okItem.Code, okItem.Body.String())
	}

	notDone := doJSON(t, handler, http.MethodPost,
		fmt.Sprintf("/api/v1/runs/%d/confluence", runID), nil, session, csrf)
	if notDone.Code != http.StatusBadRequest {
		t.Fatalf("publish in-progress status = %d body=%s", notDone.Code, notDone.Body.String())
	}

	complete := doJSON(t, handler, http.MethodPost,
		fmt.Sprintf("/api/v1/runs/%d/complete", runID),
		map[string]any{"closing_note": "Terminé"},
		session, csrf)
	if complete.Code != http.StatusOK {
		t.Fatalf("complete status = %d body=%s", complete.Code, complete.Body.String())
	}

	state := doJSON(t, handler, http.MethodGet,
		fmt.Sprintf("/api/v1/runs/%d/confluence", runID), nil, session, "")
	if state.Code != http.StatusOK {
		t.Fatalf("get confluence status = %d body=%s", state.Code, state.Body.String())
	}
	var confState map[string]any
	_ = json.Unmarshal(state.Body.Bytes(), &confState)
	if confState["configured"] != true || confState["can_publish"] != true {
		t.Fatalf("confluence state = %#v", confState)
	}

	publish := doJSON(t, handler, http.MethodPost,
		fmt.Sprintf("/api/v1/runs/%d/confluence", runID), nil, session, csrf)
	if publish.Code != http.StatusCreated {
		t.Fatalf("publish status = %d body=%s", publish.Code, publish.Body.String())
	}
	var link map[string]any
	_ = json.Unmarshal(publish.Body.Bytes(), &link)
	if !strings.Contains(fmt.Sprint(link["url"]), "/wiki/spaces/REV/pages/777") {
		t.Fatalf("link = %#v", link)
	}

	detail := doJSON(t, handler, http.MethodGet, fmt.Sprintf("/api/v1/runs/%d", runID), nil, session, "")
	if detail.Code != http.StatusOK {
		t.Fatalf("detail after publish = %d", detail.Code)
	}
	var runDetail map[string]any
	_ = json.Unmarshal(detail.Body.Bytes(), &runDetail)
	if !strings.Contains(fmt.Sprint(runDetail["confluence_url"]), "/wiki/spaces/REV/pages/777") {
		t.Fatalf("run detail missing confluence_url: %s", detail.Body.String())
	}
	caps := runDetail["capabilities"].(map[string]any)
	if caps["can_publish_confluence"] != true {
		t.Fatalf("can_publish_confluence = %v", caps["can_publish_confluence"])
	}

	_, bobSession, bobCSRF := seedSessionUser(t, st, cfg, "conf-bob@example.com", "Bob", auth.RoleEditor, true)
	denied := doJSON(t, handler, http.MethodPost,
		fmt.Sprintf("/api/v1/runs/%d/confluence", runID), nil, bobSession, bobCSRF)
	if denied.Code != http.StatusNotFound {
		t.Fatalf("bob publish status = %d want 404 body=%s", denied.Code, denied.Body.String())
	}
}
