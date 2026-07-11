package web_test

import (
	"context"
	"database/sql"
	"github.com/jeb-maker/revues/internal/testutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/features/projects"
	runs "github.com/jeb-maker/revues/internal/features/runs"
	"github.com/jeb-maker/revues/internal/store"
	appweb "github.com/jeb-maker/revues/internal/web"
)

const testSessionSecret = "test-secret-at-least-thirty-two-bytes"

type rbacFixture struct {
	t        *testing.T
	handler  http.Handler
	st       *store.Store
	ctx      context.Context
	sessions *auth.SessionManager

	admin       *store.User
	lead        *store.User
	contributor *store.User
	viewer      *store.User
	reader      *store.User
	outsider    *store.User

	tokens map[string]string

	project  *store.Project
	template *store.ChecklistTemplate
	run      *store.ChecklistRun
	runItem  store.RunItem
}

func newRBACFixture(t *testing.T) *rbacFixture {
	t.Helper()

	ctx := context.Background()
	db := openTestDB(t)
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	cfg := config.Config{
		Addr:          ":8080",
		BaseURL:       "http://example.com",
		SessionSecret: testSessionSecret,
		Env:           "development",
	}

	handler, _, err := appweb.NewRouter(appweb.Deps{Config: cfg, DB: db})
	if err != nil {
		t.Fatalf("NewRouter(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: testSessionSecret}

	admin, err := st.UpsertGitHubUser(ctx, 1, "admin", "admin@example.com", "Admin", "", auth.RoleAdmin)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(admin): %v", err)
	}
	lead, err := st.UpsertGitHubUser(ctx, 2, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(lead): %v", err)
	}
	contributor, err := st.UpsertGitHubUser(ctx, 3, "contrib", "contrib@example.com", "Contrib", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(contributor): %v", err)
	}
	viewer, err := st.UpsertGitHubUser(ctx, 4, "viewer", "viewer@example.com", "Viewer", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(viewer): %v", err)
	}
	reader, err := st.UpsertGitHubUser(ctx, 5, "reader", "reader@example.com", "Reader", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(reader): %v", err)
	}
	outsider, err := st.UpsertGitHubUser(ctx, 6, "outsider", "outsider@example.com", "Outsider", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(outsider): %v", err)
	}

	project, err := st.CreateProject(ctx, "Alpha", "fixture project", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	if err = st.AddProjectMember(ctx, project.ID, contributor.ID, projects.LocalRoleContributor); err != nil {
		t.Fatalf("AddProjectMember(contributor): %v", err)
	}
	if err = st.AddProjectMember(ctx, project.ID, viewer.ID, projects.LocalRoleViewer); err != nil {
		t.Fatalf("AddProjectMember(viewer): %v", err)
	}
	if err = st.AddProjectMember(ctx, project.ID, reader.ID, projects.LocalRoleViewer); err != nil {
		t.Fatalf("AddProjectMember(reader): %v", err)
	}

	template, _, err := st.CreateChecklistTemplate(ctx, "Modèle", lead.ID, nil, []store.TemplateItemInput{
		{Label: "Point", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, "Revue", lead.ID, sql.NullString{})
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}
	if err = st.StartRun(ctx, run.ID); err != nil {
		t.Fatalf("StartRun(): %v", err)
	}
	runItems, err := st.ListRunItems(ctx, run.ID)
	if err != nil || len(runItems) != 1 {
		t.Fatalf("ListRunItems() = %v, %v", runItems, err)
	}

	tokens := map[string]string{}
	for key, user := range map[string]*store.User{
		"admin":       admin,
		"lead":        lead,
		"contributor": contributor,
		"viewer":      viewer,
		"reader":      reader,
		"outsider":    outsider,
	} {
		token, _, err := sessions.CreateLoginSession(ctx, user.ID, 0)
		if err != nil {
			t.Fatalf("CreateLoginSession(%s): %v", key, err)
		}
		tokens[key] = token
	}

	return &rbacFixture{
		t:           t,
		handler:     handler,
		st:          st,
		ctx:         ctx,
		sessions:    sessions,
		admin:       admin,
		lead:        lead,
		contributor: contributor,
		viewer:      viewer,
		reader:      reader,
		outsider:    outsider,
		tokens:      tokens,
		project:     project,
		template:    template,
		run:         run,
		runItem:     runItems[0],
	}
}

func (f *rbacFixture) projectPath(suffix string) string {
	return "/projects/" + strconv.FormatInt(f.project.ID, 10) + suffix
}

func (f *rbacFixture) runPath(suffix string) string {
	return "/runs/" + strconv.FormatInt(f.run.ID, 10) + suffix
}

func (f *rbacFixture) runItemPath() string {
	return f.runPath("/items/" + strconv.FormatInt(f.runItem.ID, 10))
}

func (f *rbacFixture) csrf(token string) string {
	return auth.CSRFToken(token, testSessionSecret)
}

func (f *rbacFixture) do(method, path, tokenKey, body string) *httptest.ResponseRecorder {
	f.t.Helper()

	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	if tokenKey != "" {
		req.AddCookie(&http.Cookie{Name: "revues_session", Value: f.tokens[tokenKey]})
	}
	rec := httptest.NewRecorder()
	f.handler.ServeHTTP(rec, req)
	return rec
}

// TestRBAC_Matrix verifies route access aligned with docs/RBAC.md.
func TestRBAC_Matrix(t *testing.T) {
	f := newRBACFixture(t)

	contributorLaunch := url.Values{}
	contributorLaunch.Set("csrf_token", f.csrf(f.tokens["contributor"]))
	contributorLaunch.Set("template_id", strconv.FormatInt(f.template.ID, 10))
	contributorLaunch.Set("title", "Nouvelle revue")

	viewerLaunch := url.Values{}
	viewerLaunch.Set("csrf_token", f.csrf(f.tokens["viewer"]))
	viewerLaunch.Set("template_id", strconv.FormatInt(f.template.ID, 10))
	viewerLaunch.Set("title", "Nouvelle revue")

	outsiderLaunch := url.Values{}
	outsiderLaunch.Set("csrf_token", f.csrf(f.tokens["outsider"]))
	outsiderLaunch.Set("template_id", strconv.FormatInt(f.template.ID, 10))
	outsiderLaunch.Set("title", "Nouvelle revue")

	contributorItem := url.Values{}
	contributorItem.Set("csrf_token", f.csrf(f.tokens["contributor"]))
	contributorItem.Set("status", runs.StatusOK)
	contributorItem.Set("comment", "")

	viewerItem := url.Values{}
	viewerItem.Set("csrf_token", f.csrf(f.tokens["viewer"]))
	viewerItem.Set("status", runs.StatusOK)
	viewerItem.Set("comment", "")

	outsiderItem := url.Values{}
	outsiderItem.Set("csrf_token", f.csrf(f.tokens["outsider"]))
	outsiderItem.Set("status", runs.StatusOK)
	outsiderItem.Set("comment", "")

	tests := []struct {
		name       string
		method     string
		path       string
		tokenKey   string
		body       string
		wantStatus int
	}{
		// GET /projects — auth required; list filtered by membership (except admin).
		{"GET /projects anonymous", http.MethodGet, "/projects", "", "", http.StatusFound},
		{"GET /projects reader member", http.MethodGet, "/projects", "reader", "", http.StatusOK},
		{"GET /projects outsider", http.MethodGet, "/projects", "outsider", "", http.StatusOK},

		// GET /projects/{id} — auth + project member or admin.
		{"GET /projects/{id} outsider", http.MethodGet, f.projectPath(""), "outsider", "", http.StatusNotFound},
		{"GET /projects/{id} viewer", http.MethodGet, f.projectPath(""), "viewer", "", http.StatusOK},
		{"GET /projects/{id} admin bypass", http.MethodGet, f.projectPath(""), "admin", "", http.StatusOK},
		{"GET /projects/{id} reader member", http.MethodGet, f.projectPath(""), "reader", "", http.StatusOK},

		// POST /projects/{id}/runs — auth + lead/contributor or admin.
		{"POST /projects/{id}/runs viewer denied", http.MethodPost, f.projectPath("/runs"), "viewer", viewerLaunch.Encode(), http.StatusNotFound},
		{"POST /projects/{id}/runs contributor ok", http.MethodPost, f.projectPath("/runs"), "contributor", contributorLaunch.Encode(), http.StatusSeeOther},
		{"POST /projects/{id}/runs outsider denied", http.MethodPost, f.projectPath("/runs"), "outsider", outsiderLaunch.Encode(), http.StatusNotFound},

		// GET /runs/{id} — auth + project member or admin.
		{"GET /runs/{id} outsider denied", http.MethodGet, f.runPath(""), "outsider", "", http.StatusNotFound},
		{"GET /runs/{id} viewer ok", http.MethodGet, f.runPath(""), "viewer", "", http.StatusOK},
		{"GET /runs/{id} admin bypass", http.MethodGet, f.runPath(""), "admin", "", http.StatusOK},

		// POST /runs/{id}/items/{itemId} — auth + contributor+ or admin (RBAC.md PATCH).
		{"POST run item viewer denied", http.MethodPost, f.runItemPath(), "viewer", viewerItem.Encode(), http.StatusNotFound},
		{"POST run item contributor ok", http.MethodPost, f.runItemPath(), "contributor", contributorItem.Encode(), http.StatusSeeOther},
		{"POST run item outsider denied", http.MethodPost, f.runItemPath(), "outsider", outsiderItem.Encode(), http.StatusNotFound},

		// POST /admin/* — auth + admin (except /admin/users → org admin).
		{"GET /admin editor denied", http.MethodGet, "/admin", "lead", "", http.StatusForbidden},
		{"GET /admin admin redirect", http.MethodGet, "/admin", "admin", "", http.StatusFound},
		{"GET /admin/users editor denied", http.MethodGet, "/admin/users", "lead", "", http.StatusForbidden},
		{"GET /admin/users reader denied", http.MethodGet, "/admin/users", "reader", "", http.StatusForbidden},
		{"GET /admin/users admin ok", http.MethodGet, "/admin/users", "admin", "", http.StatusOK},
		{"GET /admin/integrations editor denied", http.MethodGet, "/admin/integrations", "lead", "", http.StatusForbidden},
		{"GET /admin/integrations admin ok", http.MethodGet, "/admin/integrations", "admin", "", http.StatusOK},
		{"POST /admin/users editor denied", http.MethodPost, "/admin/users", "lead", "email=x@example.com&role=editor", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := f.do(tt.method, tt.path, tt.tokenKey, tt.body)
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

// TestIDOR_CrossProject ensures user A cannot access project B resources (404, not 403).
func TestIDOR_CrossProject(t *testing.T) {
	f := newRBACFixture(t)

	aliceProject, err := f.st.CreateProject(f.ctx, "Secret", "hidden", f.lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(alice): %v", err)
	}
	template, _, err := f.st.CreateChecklistTemplate(f.ctx, "Modèle", f.lead.ID, nil, []store.TemplateItemInput{
		{Label: "Point", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := f.st.CreateChecklistRun(f.ctx, aliceProject.ID, template.ID, "Revue secrète", f.lead.ID, sql.NullString{})
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}
	if err = f.st.StartRun(f.ctx, run.ID); err != nil {
		t.Fatalf("StartRun(): %v", err)
	}
	runItems, err := f.st.ListRunItems(f.ctx, run.ID)
	if err != nil || len(runItems) != 1 {
		t.Fatalf("ListRunItems() = %v, %v", runItems, err)
	}

	bobToken := f.tokens["outsider"]
	projectPath := "/projects/" + strconv.FormatInt(aliceProject.ID, 10)
	runPath := "/runs/" + strconv.FormatInt(run.ID, 10)
	itemPath := runPath + "/items/" + strconv.FormatInt(runItems[0].ID, 10)

	itemForm := url.Values{}
	itemForm.Set("csrf_token", f.csrf(bobToken))
	itemForm.Set("status", runs.StatusOK)

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{"GET /projects/{id}", http.MethodGet, projectPath, "", http.StatusNotFound},
		{"GET /runs/{id}", http.MethodGet, runPath, "", http.StatusNotFound},
		{"POST /runs/{id}/items/{itemId}", http.MethodPost, itemPath, itemForm.Encode(), http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := f.do(tt.method, tt.path, "outsider", tt.body)
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (IDOR must not leak resource existence)", rec.Code, tt.wantStatus)
			}
		})
	}
}

// TestCSRF_MissingToken rejects mutating requests without a valid CSRF token.
func TestCSRF_MissingToken(t *testing.T) {
	f := newRBACFixture(t)

	launchForm := url.Values{}
	launchForm.Set("template_id", strconv.FormatInt(f.template.ID, 10))
	launchForm.Set("title", "Sans CSRF")

	itemForm := url.Values{}
	itemForm.Set("status", runs.StatusOK)

	adminForm := url.Values{}
	adminForm.Set("email", "new@example.com")
	adminForm.Set("role", auth.RoleEditor)

	tests := []struct {
		name     string
		method   string
		path     string
		tokenKey string
		body     string
	}{
		{"POST /logout", http.MethodPost, "/logout", "reader", ""},
		{"POST /projects/{id}/runs", http.MethodPost, f.projectPath("/runs"), "contributor", launchForm.Encode()},
		{"POST /runs/{id}/items/{itemId}", http.MethodPost, f.runItemPath(), "contributor", itemForm.Encode()},
		{"POST /admin/users", http.MethodPost, "/admin/users", "admin", adminForm.Encode()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := f.do(tt.method, tt.path, tt.tokenKey, tt.body)
			if rec.Code != http.StatusForbidden {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
			}
		})
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/rbac.db", 0)
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	t.Cleanup(func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Errorf("Close(): %v", closeErr)
		}
	})
	if migrateErr := store.Migrate(ctx, db); migrateErr != nil {
		t.Fatalf("Migrate(): %v", migrateErr)
	}
	return db
}
