package apiv1_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/integrations/notion"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
	appweb "github.com/jeb-maker/revues/internal/web"
)

func TestAdminNotion_MaskedTokenAndOrgAdmin(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		SessionSecret:  "test-secret-at-least-thirty-two-bytes",
		EncryptionKey:  testutil.EncryptionKey(),
		Env:            "development",
		AttachmentsDir: t.TempDir(),
	}
	handler, st := newTestRouterWithStore(t, cfg)

	_, session, csrf := seedSessionUser(t, st, cfg, "notion-admin@example.com", "Notion Admin", auth.RoleAdmin, true)
	_, memberSession, memberCSRF := seedSessionUser(t, st, cfg, "notion-member@example.com", "Member", auth.RoleEditor, true)

	denied := doJSON(t, handler, http.MethodGet, "/api/v1/admin/integrations/notion", nil, memberSession, "")
	if denied.Code != http.StatusForbidden {
		t.Fatalf("member get status = %d", denied.Code)
	}
	deniedPut := doJSON(t, handler, http.MethodPut, "/api/v1/admin/integrations/notion", map[string]any{
		"api_token": "secret_token", "workspace_name": "WS",
	}, memberSession, memberCSRF)
	if deniedPut.Code != http.StatusForbidden {
		t.Fatalf("member put status = %d", deniedPut.Code)
	}

	put := doJSON(t, handler, http.MethodPut, "/api/v1/admin/integrations/notion", map[string]any{
		"api_token": "super-secret-notion-token", "workspace_name": "Acme",
		"default_database_id": "abc123def4567890abc123def4567890",
	}, session, csrf)
	if put.Code != http.StatusOK {
		t.Fatalf("put status = %d body=%s", put.Code, put.Body.String())
	}
	body := put.Body.String()
	if strings.Contains(body, "super-secret-notion-token") {
		t.Fatalf("token leaked: %s", body)
	}
	var settings map[string]any
	if err := json.Unmarshal(put.Body.Bytes(), &settings); err != nil {
		t.Fatalf("json: %v", err)
	}
	if settings["has_api_token"] != true || settings["configured"] != true || settings["export_ready"] != true {
		t.Fatalf("settings = %#v", settings)
	}
	if settings["workspace_name"] != "Acme" {
		t.Fatalf("workspace = %v", settings["workspace_name"])
	}
	if _, ok := settings["api_token"]; ok {
		t.Fatal("api_token must be absent")
	}

	// Blank token keeps existing
	keep := doJSON(t, handler, http.MethodPut, "/api/v1/admin/integrations/notion", map[string]any{
		"api_token": "", "workspace_name": "Acme 2", "default_database_id": "abc123def4567890abc123def4567890",
	}, session, csrf)
	if keep.Code != http.StatusOK {
		t.Fatalf("keep status = %d body=%s", keep.Code, keep.Body.String())
	}
	var kept map[string]any
	_ = json.Unmarshal(keep.Body.Bytes(), &kept)
	if kept["has_api_token"] != true || kept["workspace_name"] != "Acme 2" {
		t.Fatalf("kept = %#v", kept)
	}

	del := doJSON(t, handler, http.MethodDelete, "/api/v1/admin/integrations/notion", nil, session, csrf)
	if del.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d", del.Code)
	}
	after := doJSON(t, handler, http.MethodGet, "/api/v1/admin/integrations/notion", nil, session, "")
	var cleared map[string]any
	_ = json.Unmarshal(after.Body.Bytes(), &cleared)
	if cleared["configured"] != false {
		t.Fatalf("configured after delete = %v", cleared["configured"])
	}
}

func TestAdminNotion_TestConnectionMock(t *testing.T) {
	t.Parallel()

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/users/me" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"Bot","bot":{"workspace_name":"Revues WS"}}`))
	}))
	t.Cleanup(mock.Close)

	cfg := config.Config{
		SessionSecret:  "test-secret-at-least-thirty-two-bytes",
		EncryptionKey:  testutil.EncryptionKey(),
		Env:            "development",
		AttachmentsDir: t.TempDir(),
	}
	handler, st := newTestRouterWithNotion(t, cfg, &notion.Client{
		HTTPClient: mock.Client(),
		APIBaseURL: mock.URL + "/v1",
	})
	_, session, csrf := seedSessionUser(t, st, cfg, "notion-test@example.com", "Tester", auth.RoleAdmin, true)

	put := doJSON(t, handler, http.MethodPut, "/api/v1/admin/integrations/notion", map[string]any{
		"api_token": "tok", "workspace_name": "WS",
	}, session, csrf)
	if put.Code != http.StatusOK {
		t.Fatalf("put status = %d body=%s", put.Code, put.Body.String())
	}

	testRec := doJSON(t, handler, http.MethodPost, "/api/v1/admin/integrations/notion/test", nil, session, csrf)
	if testRec.Code != http.StatusOK {
		t.Fatalf("test status = %d body=%s", testRec.Code, testRec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(testRec.Body.Bytes(), &resp)
	if resp["ok"] != true || !strings.Contains(fmt.Sprint(resp["message"]), "Revues WS") {
		t.Fatalf("resp = %#v", resp)
	}
}

func TestNotionImport_MockWizard(t *testing.T) {
	t.Parallel()

	dbID := "a1b2c3d4e5f6478990abcdef12345678"
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/databases/"):
			_, _ = w.Write([]byte(`{
				"id":"` + dbID + `",
				"title":[{"plain_text":"Checklist"}],
				"properties":{
					"Name":{"type":"title"},
					"Section":{"type":"select"},
					"Aide":{"type":"rich_text"},
					"Requis":{"type":"checkbox"}
				}
			}`))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/query"):
			_, _ = w.Write([]byte(`{
				"results":[{
					"properties":{
						"Name":{"title":[{"plain_text":"Backup"}]},
						"Section":{"select":{"name":"Infra"}},
						"Aide":{"rich_text":[{"plain_text":"Vérifier"}]},
						"Requis":{"checkbox":true}
					}
				}],
				"has_more":false
			}`))
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
	}
	handler, st := newTestRouterWithNotion(t, cfg, &notion.Client{
		HTTPClient: mock.Client(),
		APIBaseURL: mock.URL + "/v1",
	})
	_, session, csrf := seedSessionUser(t, st, cfg, "notion-import@example.com", "Importer", auth.RoleEditor, true)

	// Save notion config as org admin path — editor cannot; use admin
	admin, adminSession, adminCSRF := seedSessionUser(t, st, cfg, "notion-import-admin@example.com", "Admin", auth.RoleAdmin, true)
	_ = admin
	put := doJSON(t, handler, http.MethodPut, "/api/v1/admin/integrations/notion", map[string]any{
		"api_token": "tok",
	}, adminSession, adminCSRF)
	if put.Code != http.StatusOK {
		t.Fatalf("put notion = %d %s", put.Code, put.Body.String())
	}

	fetch := doJSON(t, handler, http.MethodPost, "/api/v1/templates/notion-import", map[string]any{
		"action": "fetch", "database_ref": dbID,
	}, session, csrf)
	if fetch.Code != http.StatusOK {
		t.Fatalf("fetch status = %d body=%s", fetch.Code, fetch.Body.String())
	}
	var fetchBody map[string]any
	_ = json.Unmarshal(fetch.Body.Bytes(), &fetchBody)
	if fetchBody["step"] != "mapping" {
		t.Fatalf("step = %v", fetchBody["step"])
	}

	preview := doJSON(t, handler, http.MethodPost, "/api/v1/templates/notion-import", map[string]any{
		"action": "preview", "database_id": dbID,
		"mapping":       map[string]any{"label": "Name", "section": "Section", "help_text": "Aide", "required": "Requis"},
		"template_name": "Depuis Notion",
	}, session, csrf)
	if preview.Code != http.StatusOK {
		t.Fatalf("preview status = %d body=%s", preview.Code, preview.Body.String())
	}

	imp := doJSON(t, handler, http.MethodPost, "/api/v1/templates/notion-import", map[string]any{
		"action": "import", "database_id": dbID,
		"mapping":       map[string]any{"label": "Name", "section": "Section", "help_text": "Aide", "required": "Requis"},
		"template_name": "Depuis Notion",
	}, session, csrf)
	if imp.Code != http.StatusCreated {
		t.Fatalf("import status = %d body=%s", imp.Code, imp.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(imp.Body.Bytes(), &created)
	if created["step"] != "done" {
		t.Fatalf("created = %#v", created)
	}
	tpl, _ := created["template"].(map[string]any)
	if tpl == nil || tpl["name"] != "Depuis Notion" {
		t.Fatalf("template = %#v", tpl)
	}
}

func TestNotionExport_Mock(t *testing.T) {
	t.Parallel()

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"page-id","url":"https://notion.so/exported-page"}`))
	}))
	t.Cleanup(mock.Close)

	cfg := config.Config{
		SessionSecret:  "test-secret-at-least-thirty-two-bytes",
		EncryptionKey:  testutil.EncryptionKey(),
		Env:            "development",
		BaseURL:        "http://example.com",
		AttachmentsDir: t.TempDir(),
	}
	handler, st := newTestRouterWithNotion(t, cfg, &notion.Client{
		HTTPClient: mock.Client(),
		APIBaseURL: mock.URL + "/v1",
	})
	editor, session, csrf := seedSessionUser(t, st, cfg, "notion-export@example.com", "Exporter", auth.RoleEditor, true)

	admin, adminSession, adminCSRF := seedSessionUser(t, st, cfg, "notion-export-admin@example.com", "Admin", auth.RoleAdmin, true)
	_ = admin
	put := doJSON(t, handler, http.MethodPut, "/api/v1/admin/integrations/notion", map[string]any{
		"api_token": "tok", "default_database_id": "abc123def4567890abc123def4567890",
	}, adminSession, adminCSRF)
	if put.Code != http.StatusOK {
		t.Fatalf("put = %d %s", put.Code, put.Body.String())
	}

	tplRec := doJSON(t, handler, http.MethodPost, "/api/v1/templates", map[string]any{
		"name": "Export tpl", "items": []map[string]any{{"label": "Point", "required": true}},
	}, session, csrf)
	if tplRec.Code != http.StatusCreated {
		t.Fatalf("tpl = %d %s", tplRec.Code, tplRec.Body.String())
	}
	var tpl map[string]any
	_ = json.Unmarshal(tplRec.Body.Bytes(), &tpl)
	tplID := int64(tpl["id"].(float64))

	subRec := doJSON(t, handler, http.MethodPost, "/api/v1/subjects", map[string]any{
		"name": "Sujet Notion",
	}, session, csrf)
	if subRec.Code != http.StatusCreated {
		t.Fatalf("subject = %d %s", subRec.Code, subRec.Body.String())
	}
	var sub map[string]any
	_ = json.Unmarshal(subRec.Body.Bytes(), &sub)
	subjectID := int64(sub["id"].(float64))

	launch := doJSON(t, handler, http.MethodPost, fmt.Sprintf("/api/v1/subjects/%d/runs", subjectID), map[string]any{
		"template_id": tplID,
	}, session, csrf)
	if launch.Code != http.StatusCreated {
		t.Fatalf("launch = %d %s", launch.Code, launch.Body.String())
	}
	var run map[string]any
	_ = json.Unmarshal(launch.Body.Bytes(), &run)
	runID := int64(run["id"].(float64))
	items, _ := run["items"].([]any)
	item := items[0].(map[string]any)
	itemID := int64(item["id"].(float64))

	_ = doJSON(t, handler, http.MethodPatch, fmt.Sprintf("/api/v1/runs/%d/items/%d", runID, itemID), map[string]any{
		"status": "ok",
	}, session, csrf)
	complete := doJSON(t, handler, http.MethodPost, fmt.Sprintf("/api/v1/runs/%d/complete", runID), map[string]any{
		"closing_note": "done",
	}, session, csrf)
	if complete.Code != http.StatusOK {
		t.Fatalf("complete = %d %s", complete.Code, complete.Body.String())
	}

	detail := doJSON(t, handler, http.MethodGet, fmt.Sprintf("/api/v1/runs/%d", runID), nil, session, "")
	var detailBody map[string]any
	_ = json.Unmarshal(detail.Body.Bytes(), &detailBody)
	caps := detailBody["capabilities"].(map[string]any)
	if caps["can_export_notion"] != true {
		t.Fatalf("can_export_notion = %v (editor=%d)", caps["can_export_notion"], editor.ID)
	}

	exp := doJSON(t, handler, http.MethodPost, fmt.Sprintf("/api/v1/runs/%d/notion-export", runID), nil, session, csrf)
	if exp.Code != http.StatusOK {
		t.Fatalf("export = %d %s", exp.Code, exp.Body.String())
	}
	var expBody map[string]any
	_ = json.Unmarshal(exp.Body.Bytes(), &expBody)
	if expBody["notion_url"] != "https://notion.so/exported-page" {
		t.Fatalf("export body = %#v", expBody)
	}

	again := doJSON(t, handler, http.MethodPost, fmt.Sprintf("/api/v1/runs/%d/notion-export", runID), nil, session, csrf)
	if again.Code != http.StatusConflict {
		t.Fatalf("re-export status = %d body=%s", again.Code, again.Body.String())
	}
}

func newTestRouterWithNotion(t *testing.T, cfg config.Config, client *notion.Client) (http.Handler, *store.Store) {
	t.Helper()
	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/test.db", 0)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err = store.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	st := store.New(db)
	_ = testutil.DefaultOrgContext(ctx, st)

	handler, _, _, err := appweb.NewRouter(appweb.Deps{Config: cfg, DB: db, NotionClient: client})
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}
	return handler, st
}
