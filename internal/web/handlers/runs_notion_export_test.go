package handlers_test

import (
	"context"
	"database/sql"
	"encoding/base64"
	"github.com/jeb-maker/revues/internal/testutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	runs "github.com/jeb-maker/revues/internal/features/runs"
	"github.com/jeb-maker/revues/internal/integrations/notion"
	"github.com/jeb-maker/revues/internal/store"
	appmiddleware "github.com/jeb-maker/revues/internal/web/middleware"
	viewtemplates "github.com/jeb-maker/revues/internal/web/templates"
)

func setupNotionExport(t *testing.T, st *store.Store, ctx context.Context, encKey string) {
	t.Helper()
	cfg := config.Config{EncryptionKey: encKey}
	key, _ := cfg.EncryptionKeyBytes()
	if err := (&notion.Service{Store: st, EncryptionKey: key}).Save(ctx, notion.Config{
		APIToken: "notion-token", DefaultDatabaseID: "abc123def4567890abc123def4567890",
	}); err != nil {
		t.Fatalf("Save(): %v", err)
	}
}

func testNotionExportRouter(t *testing.T, encKey string, notionClient *notion.Client) (http.Handler, *sql.DB) {
	t.Helper()
	ctx := context.Background()
	db, _ := store.Open(ctx, t.TempDir()+"/test.db", 0)
	t.Cleanup(func() { _ = db.Close() })
	_ = store.Migrate(ctx, db)
	key, _ := base64.StdEncoding.DecodeString(encKey)
	tpl, _ := viewtemplates.Parse("")
	st := store.New(db)
	deps := runs.Deps{Templates: tpl, Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	runsHandler := &runs.Runs{Deps: deps, EncryptionKey: key, BaseURL: "http://example.com", NotionClient: notionClient}
	r := chi.NewRouter()
	r.Use(appmiddleware.LoadUser(st))
	r.Use(appmiddleware.LoadActiveOrganization(st))
	r.Use(appmiddleware.CSRF("test-secret-at-least-thirty-two-bytes"))
	r.Group(func(r chi.Router) {
		r.Use(appmiddleware.RequireAuth)
		r.Get("/runs/{id}", runsHandler.Show)
		r.Post("/runs/{id}/export/notion", runsHandler.ExportNotion)
	})
	return r, db
}

func TestRuns_ExportNotion(t *testing.T) {
	encKey := config.TestEncryptionKey()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"page-id","url":"https://notion.so/revue-42"}`))
	}))
	t.Cleanup(srv.Close)
	handler, db := testNotionExportRouter(t, encKey, &notion.Client{HTTPClient: srv.Client(), APIBaseURL: srv.URL + "/v1"})
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	setupNotionExport(t, st, ctx, encKey)
	lead, _ := st.UpsertGitHubUser(ctx, 60, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	project, _ := st.CreateProject(ctx, "Alpha", "", lead.ID, nil)
	run := setupDoneRun(t, st, ctx, lead, project)
	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, _ := sessions.CreateLoginSession(ctx, lead.ID, 0)
	form := url.Values{"csrf_token": {auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes")}}
	req := httptest.NewRequest(http.MethodPost, "/runs/"+strconv.FormatInt(run.ID, 10)+"/export/notion", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d", rec.Code)
	}
	updated, _ := st.RunByID(ctx, run.ID)
	if updated.NotionURL != "https://notion.so/revue-42" {
		t.Fatalf("notion_url = %q", updated.NotionURL)
	}
}

func TestRuns_ExportNotion_IDOR(t *testing.T) {
	handler, db := testRouterWithEncryptionKey(t, config.TestEncryptionKey())
	ctx := context.Background()
	st := store.New(db)
	alice, _ := st.UpsertGitHubUser(ctx, 61, "alice", "alice@example.com", "Alice", "", auth.RoleEditor)
	bob, _ := st.UpsertGitHubUser(ctx, 62, "bob", "bob@example.com", "Bob", "", auth.RoleEditor)
	ctx = testutil.SetupIsolatedOrg(ctx, st, "Alice Org", "alice-notion-export", alice.ID)
	setupNotionExport(t, st, ctx, config.TestEncryptionKey())
	projectA, _ := st.CreateProject(ctx, "Secret", "", alice.ID, nil)
	run := setupDoneRun(t, st, ctx, alice, projectA)
	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	bobToken, _, _ := sessions.CreateLoginSession(ctx, bob.ID, 0)
	form := url.Values{"csrf_token": {auth.CSRFToken(bobToken, "test-secret-at-least-thirty-two-bytes")}}
	req := httptest.NewRequest(http.MethodPost, "/runs/"+strconv.FormatInt(run.ID, 10)+"/export/notion", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: bobToken})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestRuns_ExportNotion_ReaderForbidden(t *testing.T) {
	handler, db := testRouterWithEncryptionKey(t, config.TestEncryptionKey())
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	setupNotionExport(t, st, ctx, config.TestEncryptionKey())
	lead, _ := st.UpsertGitHubUser(ctx, 63, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	reader, _ := st.UpsertGitHubUser(ctx, 64, "reader", "reader@example.com", "Reader", "", auth.RoleReader)
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, reader.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(reader): %v", err)
	}
	project, _ := st.CreateProject(ctx, "Team", "", lead.ID, nil)
	run := setupDoneRun(t, st, ctx, lead, project)
	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, _ := sessions.CreateLoginSession(ctx, reader.ID, 0)
	form := url.Values{"csrf_token": {auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes")}}
	req := httptest.NewRequest(http.MethodPost, "/runs/"+strconv.FormatInt(run.ID, 10)+"/export/notion", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestRuns_ShowDoneIncludesNotionExportButton(t *testing.T) {
	handler, db := testRouterWithEncryptionKey(t, config.TestEncryptionKey())
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	setupNotionExport(t, st, ctx, config.TestEncryptionKey())
	lead, _ := st.UpsertGitHubUser(ctx, 65, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	project, _ := st.CreateProject(ctx, "Alpha", "", lead.ID, nil)
	run := setupDoneRun(t, st, ctx, lead, project)
	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, _ := sessions.CreateLoginSession(ctx, lead.ID, 0)
	req := httptest.NewRequest(http.MethodGet, "/runs/"+strconv.FormatInt(run.ID, 10), nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if !strings.Contains(rec.Body.String(), "Exporter vers Notion") {
		t.Fatal("expected export button")
	}
}

func TestRuns_ShowDoneIncludesNotionLinkAfterExport(t *testing.T) {
	handler, db := testRouterWithEncryptionKey(t, config.TestEncryptionKey())
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	lead, _ := st.UpsertGitHubUser(ctx, 66, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	project, _ := st.CreateProject(ctx, "Alpha", "", lead.ID, nil)
	run := setupDoneRun(t, st, ctx, lead, project)
	_ = st.SetRunNotionURL(ctx, run.ID, "https://notion.so/revue-done")
	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, _ := sessions.CreateLoginSession(ctx, lead.ID, 0)
	req := httptest.NewRequest(http.MethodGet, "/runs/"+strconv.FormatInt(run.ID, 10), nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if !strings.Contains(rec.Body.String(), `href="https://notion.so/revue-done"`) {
		t.Fatal("expected notion link")
	}
}

func TestRuns_ExportNotion_NotDone(t *testing.T) {
	handler, db := testRouterWithEncryptionKey(t, config.TestEncryptionKey())
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	setupNotionExport(t, st, ctx, config.TestEncryptionKey())
	lead, _ := st.UpsertGitHubUser(ctx, 67, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	project, _ := st.CreateProject(ctx, "Alpha", "", lead.ID, nil)
	template, _, _ := st.CreateChecklistTemplate(ctx, "Modèle", lead.ID, nil, []store.TemplateItemInput{{Label: "Point", Required: true}})
	run, _ := st.CreateChecklistRun(ctx, project.ID, template.ID, lead.ID)
	_ = st.StartRun(ctx, run.ID)
	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, _ := sessions.CreateLoginSession(ctx, lead.ID, 0)
	form := url.Values{"csrf_token": {auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes")}}
	req := httptest.NewRequest(http.MethodPost, "/runs/"+strconv.FormatInt(run.ID, 10)+"/export/notion", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}
