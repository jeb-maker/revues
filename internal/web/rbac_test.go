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
	runs "github.com/jeb-maker/revues/internal/features/runs"
	"github.com/jeb-maker/revues/internal/orgctx"
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
	orgAdmin    *store.User
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
	orgAdmin, err := st.UpsertGitHubUser(ctx, 7, "orgadmin", "orgadmin@example.com", "OrgAdmin", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(orgAdmin): %v", err)
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

	teamOrg, err := st.CreateOrganization(ctx, "Team", "team-fixture", lead.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(): %v", err)
	}
	for _, member := range []struct {
		userID int64
		role   string
	}{
		{admin.ID, store.OrgRoleMember},
		{orgAdmin.ID, store.OrgRoleAdmin},
		{lead.ID, store.OrgRoleMember},
		{contributor.ID, store.OrgRoleMember},
		{viewer.ID, store.OrgRoleMember},
		{reader.ID, store.OrgRoleMember},
	} {
		if err = st.AddOrganizationMember(ctx, teamOrg.ID, member.userID, member.role); err != nil {
			t.Fatalf("AddOrganizationMember(%d): %v", member.userID, err)
		}
	}

	teamCtx := orgctx.WithOrganizationID(ctx, teamOrg.ID)
	project, err := st.CreateProject(teamCtx, "Alpha", "fixture project", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}

	template, _, err := st.CreateChecklistTemplate(teamCtx, "Modèle", lead.ID, nil, []store.TemplateItemInput{
		{Label: "Point", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(teamCtx, project.ID, template.ID, lead.ID)
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}
	if err = st.StartRun(teamCtx, run.ID); err != nil {
		t.Fatalf("StartRun(): %v", err)
	}
	runItems, err := st.ListRunItems(teamCtx, run.ID)
	if err != nil || len(runItems) != 1 {
		t.Fatalf("ListRunItems() = %v, %v", runItems, err)
	}

	tokens := map[string]string{}
	for key, user := range map[string]*store.User{
		"admin":       admin,
		"orgAdmin":    orgAdmin,
		"lead":        lead,
		"contributor": contributor,
		"viewer":      viewer,
		"reader":      reader,
	} {
		token, _, loginErr := sessions.CreateLoginSession(ctx, user.ID, teamOrg.ID)
		if loginErr != nil {
			t.Fatalf("CreateLoginSession(%s): %v", key, loginErr)
		}
		tokens[key] = token
	}
	token, _, err := sessions.CreateLoginSession(ctx, outsider.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(outsider): %v", err)
	}
	tokens["outsider"] = token

	return &rbacFixture{
		t:           t,
		handler:     handler,
		st:          st,
		ctx:         teamCtx,
		sessions:    sessions,
		admin:       admin,
		orgAdmin:    orgAdmin,
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

func (f *rbacFixture) subjectPath(suffix string) string {
	return "/subjects/" + strconv.FormatInt(f.project.ID, 10) + suffix
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

	readerLaunch := url.Values{}
	readerLaunch.Set("csrf_token", f.csrf(f.tokens["reader"]))
	readerLaunch.Set("template_id", strconv.FormatInt(f.template.ID, 10))
	readerLaunch.Set("title", "Nouvelle revue")

	readerItem := url.Values{}
	readerItem.Set("csrf_token", f.csrf(f.tokens["reader"]))
	readerItem.Set("status", runs.StatusOK)
	readerItem.Set("comment", "")

	tests := []struct {
		name       string
		method     string
		path       string
		tokenKey   string
		body       string
		wantStatus int
	}{
		// GET /subjects — auth required; list filtered by membership (except admin).
		{"GET /subjects anonymous", http.MethodGet, "/subjects", "", "", http.StatusFound},
		{"GET /subjects reader member", http.MethodGet, "/subjects", "reader", "", http.StatusOK},
		{"GET /subjects outsider", http.MethodGet, "/subjects", "outsider", "", http.StatusOK},

		// GET /subjects/{id} — auth + org member or admin.
		{"GET /subjects/{id} outsider", http.MethodGet, f.subjectPath(""), "outsider", "", http.StatusNotFound},
		{"GET /subjects/{id} viewer", http.MethodGet, f.subjectPath(""), "viewer", "", http.StatusOK},
		{"GET /subjects/{id} admin bypass", http.MethodGet, f.subjectPath(""), "admin", "", http.StatusOK},
		{"GET /subjects/{id} reader member", http.MethodGet, f.subjectPath(""), "reader", "", http.StatusOK},

		// POST /subjects/{id}/revues — auth + org member editor+ (v1).
		{"POST /subjects/{id}/revues viewer ok", http.MethodPost, f.subjectPath("/revues"), "viewer", viewerLaunch.Encode(), http.StatusSeeOther},
		{"POST /subjects/{id}/revues contributor ok", http.MethodPost, f.subjectPath("/revues"), "contributor", contributorLaunch.Encode(), http.StatusSeeOther},
		{"POST /subjects/{id}/revues outsider denied", http.MethodPost, f.subjectPath("/revues"), "outsider", outsiderLaunch.Encode(), http.StatusNotFound},
		{"POST /subjects/{id}/revues reader denied", http.MethodPost, f.subjectPath("/revues"), "reader", readerLaunch.Encode(), http.StatusNotFound},

		// GET /runs/{id} — auth + org member or admin.
		{"GET /runs/{id} outsider denied", http.MethodGet, f.runPath(""), "outsider", "", http.StatusNotFound},
		{"GET /runs/{id} viewer ok", http.MethodGet, f.runPath(""), "viewer", "", http.StatusOK},
		{"GET /runs/{id} admin bypass", http.MethodGet, f.runPath(""), "admin", "", http.StatusOK},

		// POST /runs/{id}/items/{itemId} — auth + org member editor+ (v1).
		{"POST run item viewer ok", http.MethodPost, f.runItemPath(), "viewer", viewerItem.Encode(), http.StatusSeeOther},
		{"POST run item contributor ok", http.MethodPost, f.runItemPath(), "contributor", contributorItem.Encode(), http.StatusSeeOther},
		{"POST run item outsider denied", http.MethodPost, f.runItemPath(), "outsider", outsiderItem.Encode(), http.StatusNotFound},
		{"POST run item reader denied", http.MethodPost, f.runItemPath(), "reader", readerItem.Encode(), http.StatusNotFound},

		// POST /admin/* — auth + org owner/admin (ou admin global).
		{"GET /admin editor denied", http.MethodGet, "/admin", "lead", "", http.StatusForbidden},
		{"GET /admin admin ok", http.MethodGet, "/admin", "admin", "", http.StatusOK},
		{"GET /admin/users editor denied", http.MethodGet, "/admin/users", "lead", "", http.StatusForbidden},
		{"GET /admin/users reader denied", http.MethodGet, "/admin/users", "reader", "", http.StatusForbidden},
		{"GET /admin/users admin ok", http.MethodGet, "/admin/users", "admin", "", http.StatusOK},
		{"GET /admin/teams member denied", http.MethodGet, "/admin/teams", "lead", "", http.StatusForbidden},
		{"GET /admin/teams admin ok", http.MethodGet, "/admin/teams", "admin", "", http.StatusOK},
		{"GET /admin/teams org admin ok", http.MethodGet, "/admin/teams", "orgAdmin", "", http.StatusOK},
		{"GET /admin/integrations member denied", http.MethodGet, "/admin/integrations", "lead", "", http.StatusForbidden},
		{"GET /admin/integrations admin ok", http.MethodGet, "/admin/integrations", "admin", "", http.StatusOK},
		{"GET /admin/integrations org admin ok", http.MethodGet, "/admin/integrations", "orgAdmin", "", http.StatusOK},
		{"GET /admin/settings/smtp org admin ok", http.MethodGet, "/admin/settings/smtp", "orgAdmin", "", http.StatusOK},
		{"POST /admin/users editor denied", http.MethodPost, "/admin/users", "lead", "email=x@example.com&role=editor", http.StatusForbidden},
		{"POST /admin/teams member denied", http.MethodPost, "/admin/teams", "lead", "name=QA&slug=qa", http.StatusForbidden},
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

// TestIDOR_CrossSubject ensures user A cannot access subject B resources (404, not 403).
func TestIDOR_CrossSubject(t *testing.T) {
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
	run, err := f.st.CreateChecklistRun(f.ctx, aliceProject.ID, template.ID, f.lead.ID)
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
	subjectPath := "/subjects/" + strconv.FormatInt(aliceProject.ID, 10)
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
		{"GET /subjects/{id}", http.MethodGet, subjectPath, "", http.StatusNotFound},
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

// TestIDOR_GatedSubjectWithoutGrant ensures org members without subject/team grant get 404.
func TestIDOR_GatedSubjectWithoutGrant(t *testing.T) {
	f := newRBACFixture(t)

	gated, err := f.st.CreateSubject(f.ctx, "Gated", "", f.lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject(gated): %v", err)
	}
	if err = f.st.UpsertDirectSubjectMember(f.ctx, gated.ID, f.lead.ID, store.SubjectRoleLead); err != nil {
		t.Fatalf("UpsertDirectSubjectMember(): %v", err)
	}
	team, err := f.st.CreateTeam(f.ctx, "Alpha Team", "alpha-team", "")
	if err != nil {
		t.Fatalf("CreateTeam(): %v", err)
	}
	if err = f.st.AddTeamMember(f.ctx, team.ID, f.viewer.ID); err != nil {
		t.Fatalf("AddTeamMember(): %v", err)
	}
	if err = f.st.GrantTeamSubjectRole(f.ctx, team.ID, gated.ID, store.SubjectRoleViewer, f.lead.ID); err != nil {
		t.Fatalf("GrantTeamSubjectRole(): %v", err)
	}

	run, err := f.st.CreateChecklistRun(f.ctx, gated.ID, f.template.ID, f.lead.ID)
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}

	gatedSubjectPath := "/subjects/" + strconv.FormatInt(gated.ID, 10)
	gatedRunPath := "/runs/" + strconv.FormatInt(run.ID, 10)

	tests := []struct {
		name       string
		tokenKey   string
		path       string
		wantStatus int
	}{
		{"lead sees gated subject", "lead", gatedSubjectPath, http.StatusOK},
		{"team viewer sees gated subject", "viewer", gatedSubjectPath, http.StatusOK},
		{"org member without grant 404 subject", "contributor", gatedSubjectPath, http.StatusNotFound},
		{"org member without grant 404 run", "contributor", gatedRunPath, http.StatusNotFound},
		{"team viewer sees gated run", "viewer", gatedRunPath, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := f.do(http.MethodGet, tt.path, tt.tokenKey, "")
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
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
		{"POST /subjects/{id}/revues", http.MethodPost, f.subjectPath("/revues"), "contributor", launchForm.Encode()},
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
