package subjects_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
	appweb "github.com/jeb-maker/revues/internal/web"
)

func testRouter(t *testing.T) (http.Handler, *sql.DB) {
	t.Helper()

	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/test.db", 0)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Errorf("Close() error = %v", closeErr)
		}
	})
	if migrateErr := store.Migrate(ctx, db); migrateErr != nil {
		t.Fatalf("Migrate() error = %v", migrateErr)
	}

	cfg := config.Config{
		Addr:           ":8080",
		BaseURL:        "http://example.com",
		SessionSecret:  "test-secret-at-least-thirty-two-bytes",
		Env:            "development",
		AttachmentsDir: t.TempDir() + "/attachments",
	}

	handler, _, err := appweb.NewRouter(appweb.Deps{Config: cfg, DB: db})
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	return handler, db
}

func TestHandlers_IDOR_CrossSubject(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)

	userA, err := st.UpsertGitHubUser(ctx, 10, "alice", "alice@example.com", "Alice", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(alice): %v", err)
	}
	bob, err := st.UpsertGitHubUser(ctx, 11, "bob", "bob@example.com", "Bob", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(bob): %v", err)
	}

	aliceOrg, err := st.CreateOrganization(ctx, "Alice Org", "alice-org-idor", userA.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, aliceOrg.ID, userA.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(alice): %v", err)
	}
	aliceCtx := testutil.OrgContext(ctx, aliceOrg.ID)

	subject, err := st.CreateSubject(aliceCtx, "Secret", "hidden", userA.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	bobToken, _, err := sessions.CreateLoginSession(ctx, bob.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(bob): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/subjects/"+strconv.FormatInt(subject.ID, 10), nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: bobToken})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d (IDOR must return 404)", rec.Code, http.StatusNotFound)
	}
}

func TestIDOR_CrossOrganization(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	alice, err := st.UpsertGitHubUser(ctx, 10, "alice", "alice@example.com", "Alice", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(alice): %v", err)
	}
	bob, err := st.UpsertGitHubUser(ctx, 11, "bob", "bob@example.com", "Bob", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(bob): %v", err)
	}

	orgA, err := st.CreateOrganization(ctx, "Org A", "org-a", alice.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(org-a): %v", err)
	}
	orgB, err := st.CreateOrganization(ctx, "Org B", "org-b", bob.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(org-b): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, orgA.ID, alice.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(alice): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, orgB.ID, bob.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(bob): %v", err)
	}

	ctxA := orgctx.WithOrganizationID(ctx, orgA.ID)
	subject, err := st.CreateSubject(ctxA, "Secret", "hidden", alice.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	bobToken, _, err := sessions.CreateLoginSession(ctx, bob.ID, orgB.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(bob): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/subjects/"+strconv.FormatInt(subject.ID, 10), nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: bobToken})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d (cross-org IDOR must return 404)", rec.Code, http.StatusNotFound)
	}
}

func TestSubjects_CreateAndList(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	editor, err := st.UpsertGitHubUser(ctx, 20, "carol", "carol@example.com", "Carol", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, editor.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}
	csrf := auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes")

	form := url.Values{}
	form.Set("csrf_token", csrf)
	form.Set("name", "Sujet test")
	form.Set("description", "desc")
	req := httptest.NewRequest(http.MethodPost, "/subjects", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("create status = %d, want %d", rec.Code, http.StatusSeeOther)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/subjects", nil)
	listReq.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	listRec := httptest.NewRecorder()
	handler.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", listRec.Code, http.StatusOK)
	}
	if !strings.Contains(listRec.Body.String(), "Sujet test") {
		t.Fatal("expected subject name in list")
	}
	if !strings.Contains(listRec.Body.String(), "list-toolbar") {
		t.Fatal("expected list toolbar on subjects page")
	}
	if !strings.Contains(listRec.Body.String(), `href="/subjects/new"`) {
		t.Fatal("expected create button in toolbar")
	}
}

func TestSubjects_ReaderCannotCreate(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	reader, err := st.UpsertGitHubUser(ctx, 30, "dave", "dave@example.com", "Dave", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, reader.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	form := url.Values{}
	form.Set("csrf_token", auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes"))
	form.Set("name", "Nope")
	req := httptest.NewRequest(http.MethodPost, "/subjects", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestRunsListEmptyState_ByRole(t *testing.T) {
	roles := []struct {
		role    string
		orgRole string
		want    string
		notWant string
	}{
		{auth.RoleAdmin, store.OrgRoleOwner, "Lancer une revue", "ne vous est encore assigné"},
		{auth.RoleEditor, store.OrgRoleMember, "Lancer une revue", "Gérer les emails autorisés"},
		{auth.RoleReader, store.OrgRoleMember, "Aucun sujet disponible", "Lancer une revue"},
	}

	for _, tt := range roles {
		t.Run(tt.role, func(t *testing.T) {
			handler, db := testRouter(t)
			ctx := context.Background()
			st := store.New(db)
			ctx = testutil.DefaultOrgContext(ctx, st)
			defaultOrg, err := st.OrganizationBySlug(ctx, "default")
			if err != nil {
				t.Fatalf("OrganizationBySlug(): %v", err)
			}

			user, err := st.UpsertGitHubUser(ctx, 40, "user-"+tt.role, tt.role+"@example.com", tt.role, "", tt.role)
			if err != nil {
				t.Fatalf("UpsertGitHubUser(): %v", err)
			}
			if err = st.AddOrganizationMember(ctx, defaultOrg.ID, user.ID, tt.orgRole); err != nil {
				t.Fatalf("AddOrganizationMember(): %v", err)
			}

			sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
			token, _, err := sessions.CreateLoginSession(ctx, user.ID, defaultOrg.ID)
			if err != nil {
				t.Fatalf("CreateLoginSession(): %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, "/revues", nil)
			req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}
			body := rec.Body.String()
			if !strings.Contains(body, "empty-state") {
				t.Fatal("expected empty dashboard state")
			}
			if !strings.Contains(body, tt.want) {
				t.Fatalf("expected CTA %q in body", tt.want)
			}
			if tt.notWant != "" && strings.Contains(body, tt.notWant) {
				t.Fatalf("unexpected CTA %q in body", tt.notWant)
			}
		})
	}
}

func TestWizardNouvelle_InlineCreate(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	editor, err := st.UpsertGitHubUser(ctx, 60, "wiz", "wiz@example.com", "Wiz", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, editor.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	form := url.Values{}
	form.Set("csrf_token", auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes"))
	form.Set("name", "Nouveau sujet")
	req := httptest.NewRequest(http.MethodPost, "/revues/nouvelle", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}

	subjects, err := st.ListSubjects(ctx, editor.ID, false, "Nouveau")
	if err != nil || len(subjects) != 1 {
		t.Fatalf("ListSubjects() = %v, %v", subjects, err)
	}
	want := "/subjects/" + strconv.FormatInt(subjects[0].ID, 10) + "/modeles?for_run=1"
	if loc := rec.Header().Get("Location"); loc != want {
		t.Fatalf("Location = %q, want %q", loc, want)
	}
}

func TestWizardNouvelle_ReaderRedirected(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	reader, err := st.UpsertGitHubUser(ctx, 63, "wiz-reader", "wiz-reader@example.com", "Reader", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, reader.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/revues/nouvelle", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("GET status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if loc := rec.Header().Get("Location"); loc != "/revues" {
		t.Fatalf("Location = %q, want /revues", loc)
	}

	form := url.Values{}
	form.Set("csrf_token", auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes"))
	form.Set("name", "Nope")
	postReq := httptest.NewRequest(http.MethodPost, "/revues/nouvelle", strings.NewReader(form.Encode()))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	postRec := httptest.NewRecorder()
	handler.ServeHTTP(postRec, postReq)
	if postRec.Code != http.StatusNotFound {
		t.Fatalf("POST status = %d, want %d", postRec.Code, http.StatusNotFound)
	}
}

func TestWizardNouvelle_SearchByName(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	editor, err := st.UpsertGitHubUser(ctx, 61, "search", "search@example.com", "Search", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	if _, err = st.CreateSubject(ctx, "Alpha infra", "", editor.ID, nil); err != nil {
		t.Fatalf("CreateSubject(alpha): %v", err)
	}
	if _, err = st.CreateSubject(ctx, "Beta mobile", "", editor.ID, nil); err != nil {
		t.Fatalf("CreateSubject(beta): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, editor.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/revues/nouvelle?q=infra", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Alpha infra") {
		t.Fatal("expected matching subject in wizard")
	}
	if strings.Contains(body, "Beta mobile") {
		t.Fatal("non-matching subject must not appear in filtered wizard")
	}
}

func TestWizardNouvelle_WithTemplate(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	editor, err := st.UpsertGitHubUser(ctx, 62, "wiz-tpl", "wiz-tpl@example.com", "Wiz", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	if _, err = st.CreateSubject(ctx, "Alpha", "", editor.ID, nil); err != nil {
		t.Fatalf("CreateSubject(): %v", err)
	}
	if _, err = st.CreateSubject(ctx, "Beta", "", editor.ID, nil); err != nil {
		t.Fatalf("CreateSubject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, "Modèle wizard", editor.ID, nil, []store.TemplateItemInput{{Label: "Point"}})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, editor.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/revues/nouvelle?template="+strconv.FormatInt(template.ID, 10), nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	wantLink := "/subjects/"
	if !strings.Contains(body, wantLink) || !strings.Contains(body, "template="+strconv.FormatInt(template.ID, 10)) {
		t.Fatalf("expected subject links to preserve template param, body=%s", body)
	}

	form := url.Values{}
	form.Set("csrf_token", auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes"))
	form.Set("name", "Inline sujet")
	form.Set("template", strconv.FormatInt(template.ID, 10))
	postReq := httptest.NewRequest(http.MethodPost, "/revues/nouvelle", strings.NewReader(form.Encode()))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	postRec := httptest.NewRecorder()
	handler.ServeHTTP(postRec, postReq)
	if postRec.Code != http.StatusSeeOther {
		t.Fatalf("create status = %d, want %d", postRec.Code, http.StatusSeeOther)
	}
	loc := postRec.Header().Get("Location")
	if !strings.Contains(loc, "for_run=1") || !strings.Contains(loc, "template="+strconv.FormatInt(template.ID, 10)) {
		t.Fatalf("Location = %q, want template preserved", loc)
	}
}

func TestAdminSubjects_RequiresOrgAdmin(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}

	reader, err := st.UpsertGitHubUser(ctx, 80, "reader-admin", "reader-admin@example.com", "Reader", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, reader.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, reader.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/subjects", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}

	owner, err := st.UpsertGitHubUser(ctx, 81, "owner-admin", "owner-admin@example.com", "Owner", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(owner): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, owner.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(owner): %v", err)
	}
	ownerToken, _, err := sessions.CreateLoginSession(ctx, owner.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(owner): %v", err)
	}

	req = httptest.NewRequest(http.MethodGet, "/admin/subjects", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: ownerToken})
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `admin-nav`) || !strings.Contains(body, `href="/admin/subjects"`) {
		t.Fatal("expected admin subjects page with admin nav")
	}
}

func TestSubjects_UpdateRedirect_ByRoute(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}

	owner, err := st.UpsertGitHubUser(ctx, 90, "owner-upd", "owner-upd@example.com", "Owner", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(owner): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, owner.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(owner): %v", err)
	}

	subject, err := st.CreateSubject(ctx, "To update", "", owner.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, owner.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}
	csrf := auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes")

	tests := []struct {
		name     string
		path     string
		wantPath string
	}{
		{
			name:     "member edit route",
			path:     "/subjects/" + strconv.FormatInt(subject.ID, 10),
			wantPath: "/subjects/" + strconv.FormatInt(subject.ID, 10) + "?msg=Sujet+mis+%C3%A0+jour",
		},
		{
			name:     "admin edit route",
			path:     "/admin/subjects/" + strconv.FormatInt(subject.ID, 10),
			wantPath: "/admin/subjects?msg=Sujet+mis+%C3%A0+jour",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Set("csrf_token", csrf)
			form.Set("name", "Updated name")
			req := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusSeeOther {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
			}
			if loc := rec.Header().Get("Location"); loc != tt.wantPath {
				t.Fatalf("Location = %q, want %q", loc, tt.wantPath)
			}
		})
	}
}

func TestSubjects_CreateRedirect_FromAdmin(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}

	owner, err := st.UpsertGitHubUser(ctx, 91, "owner-create", "owner-create@example.com", "Owner", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(owner): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, owner.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(owner): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, owner.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	form := url.Values{}
	form.Set("csrf_token", auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes"))
	form.Set("name", "Admin created")
	req := httptest.NewRequest(http.MethodPost, "/admin/subjects", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}

	subjects, err := st.ListSubjects(ctx, owner.ID, false, "Admin created")
	if err != nil || len(subjects) != 1 {
		t.Fatalf("ListSubjects() = %v, %v", subjects, err)
	}
	want := "/subjects/" + strconv.FormatInt(subjects[0].ID, 10) + "?msg=Sujet+cr%C3%A9%C3%A9"
	if loc := rec.Header().Get("Location"); loc != want {
		t.Fatalf("Location = %q, want %q", loc, want)
	}
}

func TestSubjects_ArchiveRedirect_FromAdmin(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}

	owner, err := st.UpsertGitHubUser(ctx, 92, "owner-arch", "owner-arch@example.com", "Owner", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(owner): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, owner.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(owner): %v", err)
	}

	subject, err := st.CreateSubject(ctx, "To archive", "", owner.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, owner.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	form := url.Values{}
	form.Set("csrf_token", auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes"))
	req := httptest.NewRequest(http.MethodPost, "/admin/subjects/"+strconv.FormatInt(subject.ID, 10)+"/archive", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	want := "/admin/subjects?msg=Sujet+archiv%C3%A9"
	if loc := rec.Header().Get("Location"); loc != want {
		t.Fatalf("Location = %q, want %q", loc, want)
	}
}

func TestSubjects_ShowEditLink_ByOrgRole(t *testing.T) {
	tests := []struct {
		name     string
		orgRole  string
		wantHref string
		notWant  string
	}{
		{
			name:     "org owner",
			orgRole:  store.OrgRoleOwner,
			wantHref: `href="/admin/subjects/`,
			notWant:  `href="/subjects/`,
		},
		{
			name:     "org member editor",
			orgRole:  store.OrgRoleMember,
			wantHref: `href="/subjects/`,
			notWant:  `href="/admin/subjects/`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, db := testRouter(t)
			ctx := context.Background()
			st := store.New(db)
			ctx = testutil.DefaultOrgContext(ctx, st)
			defaultOrg, err := st.OrganizationBySlug(ctx, "default")
			if err != nil {
				t.Fatalf("OrganizationBySlug(): %v", err)
			}

			user, err := st.UpsertGitHubUser(ctx, 93, "user-"+tt.orgRole, tt.orgRole+"@example.com", tt.orgRole, "", auth.RoleEditor)
			if err != nil {
				t.Fatalf("UpsertGitHubUser(): %v", err)
			}
			if err = st.AddOrganizationMember(ctx, defaultOrg.ID, user.ID, tt.orgRole); err != nil {
				t.Fatalf("AddOrganizationMember(): %v", err)
			}

			subject, err := st.CreateSubject(ctx, "Show me", "", user.ID, nil)
			if err != nil {
				t.Fatalf("CreateSubject(): %v", err)
			}

			sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
			token, _, err := sessions.CreateLoginSession(ctx, user.ID, defaultOrg.ID)
			if err != nil {
				t.Fatalf("CreateLoginSession(): %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, "/subjects/"+strconv.FormatInt(subject.ID, 10), nil)
			req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}
			body := rec.Body.String()
			editHref := tt.wantHref + strconv.FormatInt(subject.ID, 10) + `/edit"`
			if !strings.Contains(body, editHref) {
				t.Fatalf("expected edit link %q in body", editHref)
			}
			if strings.Contains(body, tt.notWant+strconv.FormatInt(subject.ID, 10)+`/edit"`) {
				t.Fatalf("unexpected edit link with prefix %q", tt.notWant)
			}
		})
	}
}

func TestSubjectTeams_AddPreviewRemove_RBAC(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}

	owner, err := st.UpsertGitHubUser(ctx, 200, "owner", "owner-teams@example.com", "Owner", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(owner): %v", err)
	}
	lead, err := st.UpsertGitHubUser(ctx, 201, "lead", "lead-teams@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(lead): %v", err)
	}
	member, err := st.UpsertGitHubUser(ctx, 202, "member", "member-teams@example.com", "Member", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(member): %v", err)
	}
	outsider, err := st.UpsertGitHubUser(ctx, 203, "outsider", "outsider-teams@example.com", "Outsider", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(outsider): %v", err)
	}

	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, owner.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(owner): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, lead.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(lead): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, member.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(member): %v", err)
	}

	subject, err := st.CreateSubject(ctx, "Sujet équipes", "", owner.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject(): %v", err)
	}
	if err = st.UpsertDirectSubjectMember(ctx, subject.ID, lead.ID, store.SubjectRoleLead); err != nil {
		t.Fatalf("UpsertDirectSubjectMember(lead): %v", err)
	}
	if err = st.UpsertDirectSubjectMember(ctx, subject.ID, member.ID, store.SubjectRoleContributor); err != nil {
		t.Fatalf("UpsertDirectSubjectMember(member): %v", err)
	}

	team, err := st.CreateTeam(ctx, "Qualité", "qualite", "")
	if err != nil {
		t.Fatalf("CreateTeam(): %v", err)
	}
	if err = st.AddTeamMember(ctx, team.ID, member.ID); err != nil {
		t.Fatalf("AddTeamMember(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	ownerToken, _, err := sessions.CreateLoginSession(ctx, owner.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(owner): %v", err)
	}
	leadToken, _, err := sessions.CreateLoginSession(ctx, lead.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(lead): %v", err)
	}
	memberToken, _, err := sessions.CreateLoginSession(ctx, member.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(member): %v", err)
	}
	outsiderToken, _, err := sessions.CreateLoginSession(ctx, outsider.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(outsider): %v", err)
	}
	ownerCSRF := auth.CSRFToken(ownerToken, "test-secret-at-least-thirty-two-bytes")
	leadCSRF := auth.CSRFToken(leadToken, "test-secret-at-least-thirty-two-bytes")
	memberCSRF := auth.CSRFToken(memberToken, "test-secret-at-least-thirty-two-bytes")

	subjectPath := "/subjects/" + strconv.FormatInt(subject.ID, 10)
	teamIDStr := strconv.FormatInt(team.ID, 10)

	showReq := httptest.NewRequest(http.MethodGet, subjectPath, nil)
	showReq.AddCookie(&http.Cookie{Name: "revues_session", Value: leadToken})
	showRec := httptest.NewRecorder()
	handler.ServeHTTP(showRec, showReq)
	if showRec.Code != http.StatusOK {
		t.Fatalf("show status = %d, want %d", showRec.Code, http.StatusOK)
	}
	showBody := showRec.Body.String()
	if !strings.Contains(showBody, "Équipes") || !strings.Contains(showBody, "Membres directs") {
		t.Fatalf("show missing teams/direct members sections")
	}
	if !strings.Contains(showBody, `tag">direct`) && !strings.Contains(showBody, ">direct<") {
		t.Fatalf("show missing direct access source badge")
	}

	previewURL := subjectPath + "/teams/preview?team_id=" + teamIDStr + "&role=contributor"
	previewReq := httptest.NewRequest(http.MethodGet, previewURL, nil)
	previewReq.AddCookie(&http.Cookie{Name: "revues_session", Value: leadToken})
	previewRec := httptest.NewRecorder()
	handler.ServeHTTP(previewRec, previewReq)
	if previewRec.Code != http.StatusOK {
		t.Fatalf("preview status = %d, want %d", previewRec.Code, http.StatusOK)
	}
	if !strings.Contains(previewRec.Body.String(), "Équipe Qualité : 1 membre aura le rôle Contributeur") {
		t.Fatalf("preview body = %q, want singular aura + role", previewRec.Body.String())
	}

	memberForm := url.Values{}
	memberForm.Set("csrf_token", memberCSRF)
	memberForm.Set("team_id", teamIDStr)
	memberForm.Set("role", "viewer")
	memberAdd := httptest.NewRequest(http.MethodPost, subjectPath+"/teams", strings.NewReader(memberForm.Encode()))
	memberAdd.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	memberAdd.AddCookie(&http.Cookie{Name: "revues_session", Value: memberToken})
	memberRec := httptest.NewRecorder()
	handler.ServeHTTP(memberRec, memberAdd)
	if memberRec.Code != http.StatusNotFound {
		t.Fatalf("contributor add status = %d, want %d", memberRec.Code, http.StatusNotFound)
	}

	outsiderForm := url.Values{}
	outsiderForm.Set("csrf_token", auth.CSRFToken(outsiderToken, "test-secret-at-least-thirty-two-bytes"))
	outsiderForm.Set("team_id", teamIDStr)
	outsiderForm.Set("role", "viewer")
	outsiderAdd := httptest.NewRequest(http.MethodPost, subjectPath+"/teams", strings.NewReader(outsiderForm.Encode()))
	outsiderAdd.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	outsiderAdd.AddCookie(&http.Cookie{Name: "revues_session", Value: outsiderToken})
	outsiderRec := httptest.NewRecorder()
	handler.ServeHTTP(outsiderRec, outsiderAdd)
	if outsiderRec.Code != http.StatusNotFound {
		t.Fatalf("outsider add status = %d, want %d", outsiderRec.Code, http.StatusNotFound)
	}

	leadForm := url.Values{}
	leadForm.Set("csrf_token", leadCSRF)
	leadForm.Set("team_id", teamIDStr)
	leadForm.Set("role", "contributor")
	leadAdd := httptest.NewRequest(http.MethodPost, subjectPath+"/teams", strings.NewReader(leadForm.Encode()))
	leadAdd.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	leadAdd.AddCookie(&http.Cookie{Name: "revues_session", Value: leadToken})
	leadRec := httptest.NewRecorder()
	handler.ServeHTTP(leadRec, leadAdd)
	if leadRec.Code != http.StatusSeeOther {
		t.Fatalf("lead add status = %d, want %d", leadRec.Code, http.StatusSeeOther)
	}

	teams, err := st.ListSubjectTeams(ctx, subject.ID)
	if err != nil || len(teams) != 1 || teams[0].Role != store.SubjectRoleContributor {
		t.Fatalf("ListSubjectTeams() = %+v, %v", teams, err)
	}

	showAfter := httptest.NewRequest(http.MethodGet, subjectPath, nil)
	showAfter.AddCookie(&http.Cookie{Name: "revues_session", Value: ownerToken})
	showAfterRec := httptest.NewRecorder()
	handler.ServeHTTP(showAfterRec, showAfter)
	if showAfterRec.Code != http.StatusOK {
		t.Fatalf("show after add status = %d", showAfterRec.Code)
	}
	if !strings.Contains(showAfterRec.Body.String(), "Qualité") {
		t.Fatalf("expected assigned team name in show body")
	}

	noCSRF := url.Values{}
	noCSRF.Set("team_id", teamIDStr)
	noCSRFReq := httptest.NewRequest(http.MethodPost, subjectPath+"/teams/remove", strings.NewReader(noCSRF.Encode()))
	noCSRFReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	noCSRFReq.AddCookie(&http.Cookie{Name: "revues_session", Value: ownerToken})
	noCSRFRec := httptest.NewRecorder()
	handler.ServeHTTP(noCSRFRec, noCSRFReq)
	if noCSRFRec.Code != http.StatusForbidden && noCSRFRec.Code != http.StatusBadRequest {
		t.Fatalf("missing CSRF status = %d, want 403 or 400", noCSRFRec.Code)
	}

	removeForm := url.Values{}
	removeForm.Set("csrf_token", ownerCSRF)
	removeForm.Set("team_id", teamIDStr)
	removeReq := httptest.NewRequest(http.MethodPost, subjectPath+"/teams/remove", strings.NewReader(removeForm.Encode()))
	removeReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	removeReq.AddCookie(&http.Cookie{Name: "revues_session", Value: ownerToken})
	removeRec := httptest.NewRecorder()
	handler.ServeHTTP(removeRec, removeReq)
	if removeRec.Code != http.StatusSeeOther {
		t.Fatalf("remove status = %d, want %d", removeRec.Code, http.StatusSeeOther)
	}
	teams, err = st.ListSubjectTeams(ctx, subject.ID)
	if err != nil || len(teams) != 0 {
		t.Fatalf("teams after remove = %+v, %v", teams, err)
	}
}

func TestSubjectTeams_IDOR_CrossOrg(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)

	alice, err := st.UpsertGitHubUser(ctx, 210, "alice-t", "alice-t@example.com", "Alice", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(alice): %v", err)
	}
	bob, err := st.UpsertGitHubUser(ctx, 211, "bob-t", "bob-t@example.com", "Bob", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(bob): %v", err)
	}

	orgA, err := st.CreateOrganization(ctx, "Org A Teams", "org-a-teams", alice.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(A): %v", err)
	}
	orgB, err := st.CreateOrganization(ctx, "Org B Teams", "org-b-teams", bob.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(B): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, orgA.ID, alice.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(alice): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, orgB.ID, bob.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(bob): %v", err)
	}

	ctxA := orgctx.WithOrganizationID(ctx, orgA.ID)
	subject, err := st.CreateSubject(ctxA, "Secret teams", "", alice.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject(): %v", err)
	}
	team, err := st.CreateTeam(ctxA, "Alpha", "alpha", "")
	if err != nil {
		t.Fatalf("CreateTeam(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	bobToken, _, err := sessions.CreateLoginSession(ctx, bob.ID, orgB.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(bob): %v", err)
	}
	csrf := auth.CSRFToken(bobToken, "test-secret-at-least-thirty-two-bytes")

	form := url.Values{}
	form.Set("csrf_token", csrf)
	form.Set("team_id", strconv.FormatInt(team.ID, 10))
	form.Set("role", "viewer")
	req := httptest.NewRequest(http.MethodPost, "/subjects/"+strconv.FormatInt(subject.ID, 10)+"/teams", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: bobToken})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("cross-org add status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	preview := httptest.NewRequest(http.MethodGet, "/subjects/"+strconv.FormatInt(subject.ID, 10)+"/teams/preview?team_id="+strconv.FormatInt(team.ID, 10)+"&role=viewer", nil)
	preview.AddCookie(&http.Cookie{Name: "revues_session", Value: bobToken})
	previewRec := httptest.NewRecorder()
	handler.ServeHTTP(previewRec, preview)
	if previewRec.Code != http.StatusNotFound {
		t.Fatalf("cross-org preview status = %d, want %d", previewRec.Code, http.StatusNotFound)
	}
}

func TestOrgPolicies_LeadAssignTeamsDenied(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}
	if err = st.UpdateOrganizationLeadPolicies(ctx, defaultOrg.ID, store.OrgLeadPolicies{
		LeadsMayAssignTeams:     false,
		LeadsMayInviteMembers:   true,
		LeadsMayInviteExternals: false,
	}); err != nil {
		t.Fatalf("UpdateOrganizationLeadPolicies(): %v", err)
	}

	owner, err := st.UpsertGitHubUser(ctx, 300, "owner-pol", "owner-pol@example.com", "Owner", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(owner): %v", err)
	}
	lead, err := st.UpsertGitHubUser(ctx, 301, "lead-pol", "lead-pol@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(lead): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, owner.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(owner): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, lead.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(lead): %v", err)
	}

	subject, err := st.CreateSubject(ctx, "Sujet policies teams", "", owner.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject(): %v", err)
	}
	if err = st.UpsertDirectSubjectMember(ctx, subject.ID, lead.ID, store.SubjectRoleLead); err != nil {
		t.Fatalf("UpsertDirectSubjectMember(): %v", err)
	}
	team, err := st.CreateTeam(ctx, "Ops", "ops-pol", "")
	if err != nil {
		t.Fatalf("CreateTeam(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	leadToken, _, err := sessions.CreateLoginSession(ctx, lead.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(lead): %v", err)
	}
	ownerToken, _, err := sessions.CreateLoginSession(ctx, owner.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(owner): %v", err)
	}
	leadCSRF := auth.CSRFToken(leadToken, "test-secret-at-least-thirty-two-bytes")
	ownerCSRF := auth.CSRFToken(ownerToken, "test-secret-at-least-thirty-two-bytes")
	subjectPath := "/subjects/" + strconv.FormatInt(subject.ID, 10)

	showReq := httptest.NewRequest(http.MethodGet, subjectPath, nil)
	showReq.AddCookie(&http.Cookie{Name: "revues_session", Value: leadToken})
	showRec := httptest.NewRecorder()
	handler.ServeHTTP(showRec, showReq)
	if showRec.Code != http.StatusOK {
		t.Fatalf("show status = %d", showRec.Code)
	}
	body := showRec.Body.String()
	if !strings.Contains(body, "n'autorise pas les leads à affecter des équipes") {
		t.Fatalf("expected teams policy denial message in body")
	}
	if strings.Contains(body, `action="`+subjectPath+`/teams"`) {
		t.Fatalf("lead must not see add-team form when policy denies")
	}

	leadForm := url.Values{}
	leadForm.Set("csrf_token", leadCSRF)
	leadForm.Set("team_id", strconv.FormatInt(team.ID, 10))
	leadForm.Set("role", "viewer")
	leadAdd := httptest.NewRequest(http.MethodPost, subjectPath+"/teams", strings.NewReader(leadForm.Encode()))
	leadAdd.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	leadAdd.AddCookie(&http.Cookie{Name: "revues_session", Value: leadToken})
	leadRec := httptest.NewRecorder()
	handler.ServeHTTP(leadRec, leadAdd)
	if leadRec.Code != http.StatusBadRequest {
		t.Fatalf("lead add status = %d, want %d", leadRec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(leadRec.Body.String(), "autorise pas les leads") ||
		!strings.Contains(leadRec.Body.String(), "affecter des équipes") {
		t.Fatalf("expected denial message on POST")
	}

	ownerForm := url.Values{}
	ownerForm.Set("csrf_token", ownerCSRF)
	ownerForm.Set("team_id", strconv.FormatInt(team.ID, 10))
	ownerForm.Set("role", "viewer")
	ownerAdd := httptest.NewRequest(http.MethodPost, subjectPath+"/teams", strings.NewReader(ownerForm.Encode()))
	ownerAdd.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ownerAdd.AddCookie(&http.Cookie{Name: "revues_session", Value: ownerToken})
	ownerRec := httptest.NewRecorder()
	handler.ServeHTTP(ownerRec, ownerAdd)
	if ownerRec.Code != http.StatusSeeOther {
		t.Fatalf("owner add status = %d, want %d", ownerRec.Code, http.StatusSeeOther)
	}
}

func TestOrgPolicies_LeadInviteMembersDenied(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}
	if err = st.UpdateOrganizationLeadPolicies(ctx, defaultOrg.ID, store.OrgLeadPolicies{
		LeadsMayAssignTeams:     true,
		LeadsMayInviteMembers:   false,
		LeadsMayInviteExternals: false,
	}); err != nil {
		t.Fatalf("UpdateOrganizationLeadPolicies(): %v", err)
	}

	owner, err := st.UpsertGitHubUser(ctx, 310, "owner-inv", "owner-inv@example.com", "Owner", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(owner): %v", err)
	}
	lead, err := st.UpsertGitHubUser(ctx, 311, "lead-inv", "lead-inv@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(lead): %v", err)
	}
	invitee, err := st.UpsertGitHubUser(ctx, 312, "invitee-inv", "invitee-inv@example.com", "Invitee", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(invitee): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, owner.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(owner): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, lead.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(lead): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, invitee.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(invitee): %v", err)
	}

	subject, err := st.CreateSubject(ctx, "Sujet policies invite", "", owner.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject(): %v", err)
	}
	if err = st.UpsertDirectSubjectMember(ctx, subject.ID, lead.ID, store.SubjectRoleLead); err != nil {
		t.Fatalf("UpsertDirectSubjectMember(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	leadToken, _, err := sessions.CreateLoginSession(ctx, lead.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(lead): %v", err)
	}
	ownerToken, _, err := sessions.CreateLoginSession(ctx, owner.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(owner): %v", err)
	}
	leadCSRF := auth.CSRFToken(leadToken, "test-secret-at-least-thirty-two-bytes")
	ownerCSRF := auth.CSRFToken(ownerToken, "test-secret-at-least-thirty-two-bytes")
	subjectPath := "/subjects/" + strconv.FormatInt(subject.ID, 10)

	showReq := httptest.NewRequest(http.MethodGet, subjectPath, nil)
	showReq.AddCookie(&http.Cookie{Name: "revues_session", Value: leadToken})
	showRec := httptest.NewRecorder()
	handler.ServeHTTP(showRec, showReq)
	if showRec.Code != http.StatusOK {
		t.Fatalf("show status = %d", showRec.Code)
	}
	if !strings.Contains(showRec.Body.String(), "n'autorise pas les leads à inviter des membres") {
		t.Fatalf("expected members policy denial message")
	}

	leadForm := url.Values{}
	leadForm.Set("csrf_token", leadCSRF)
	leadForm.Set("email", invitee.Email)
	leadForm.Set("role", "viewer")
	leadAdd := httptest.NewRequest(http.MethodPost, subjectPath+"/members", strings.NewReader(leadForm.Encode()))
	leadAdd.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	leadAdd.AddCookie(&http.Cookie{Name: "revues_session", Value: leadToken})
	leadRec := httptest.NewRecorder()
	handler.ServeHTTP(leadRec, leadAdd)
	if leadRec.Code != http.StatusBadRequest {
		t.Fatalf("lead invite status = %d, want %d", leadRec.Code, http.StatusBadRequest)
	}

	ownerForm := url.Values{}
	ownerForm.Set("csrf_token", ownerCSRF)
	ownerForm.Set("email", invitee.Email)
	ownerForm.Set("role", "contributor")
	ownerAdd := httptest.NewRequest(http.MethodPost, subjectPath+"/members", strings.NewReader(ownerForm.Encode()))
	ownerAdd.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ownerAdd.AddCookie(&http.Cookie{Name: "revues_session", Value: ownerToken})
	ownerRec := httptest.NewRecorder()
	handler.ServeHTTP(ownerRec, ownerAdd)
	if ownerRec.Code != http.StatusSeeOther {
		t.Fatalf("owner invite status = %d, want %d; body=%s", ownerRec.Code, http.StatusSeeOther, ownerRec.Body.String())
	}
}

func TestOrgPolicies_LeadInviteExternals(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}
	if err = st.UpdateOrganizationLeadPolicies(ctx, defaultOrg.ID, store.OrgLeadPolicies{
		LeadsMayAssignTeams:     true,
		LeadsMayInviteMembers:   false,
		LeadsMayInviteExternals: true,
	}); err != nil {
		t.Fatalf("UpdateOrganizationLeadPolicies(): %v", err)
	}

	owner, err := st.UpsertGitHubUser(ctx, 320, "owner-ext", "owner-ext@example.com", "Owner", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(owner): %v", err)
	}
	lead, err := st.UpsertGitHubUser(ctx, 321, "lead-ext", "lead-ext@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(lead): %v", err)
	}
	external, err := st.UpsertGitHubUser(ctx, 322, "external-ext", "external-ext@example.com", "External", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(external): %v", err)
	}
	orgMember, err := st.UpsertGitHubUser(ctx, 323, "member-ext", "member-ext@example.com", "Member", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(member): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, owner.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(owner): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, lead.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(lead): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, orgMember.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(orgMember): %v", err)
	}

	subject, err := st.CreateSubject(ctx, "Sujet policies externals", "", owner.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject(): %v", err)
	}
	if err = st.UpsertDirectSubjectMember(ctx, subject.ID, lead.ID, store.SubjectRoleLead); err != nil {
		t.Fatalf("UpsertDirectSubjectMember(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	leadToken, _, err := sessions.CreateLoginSession(ctx, lead.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(lead): %v", err)
	}
	leadCSRF := auth.CSRFToken(leadToken, "test-secret-at-least-thirty-two-bytes")
	subjectPath := "/subjects/" + strconv.FormatInt(subject.ID, 10)

	denyOrgMember := url.Values{}
	denyOrgMember.Set("csrf_token", leadCSRF)
	denyOrgMember.Set("email", orgMember.Email)
	denyOrgMember.Set("role", "viewer")
	denyReq := httptest.NewRequest(http.MethodPost, subjectPath+"/members", strings.NewReader(denyOrgMember.Encode()))
	denyReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	denyReq.AddCookie(&http.Cookie{Name: "revues_session", Value: leadToken})
	denyRec := httptest.NewRecorder()
	handler.ServeHTTP(denyRec, denyReq)
	if denyRec.Code != http.StatusBadRequest {
		t.Fatalf("invite org member status = %d, want %d", denyRec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(denyRec.Body.String(), "autorise pas les leads") ||
		!strings.Contains(denyRec.Body.String(), "inviter des membres") {
		t.Fatalf("expected members policy denial, body=%q", denyRec.Body.String())
	}

	allowExternal := url.Values{}
	allowExternal.Set("csrf_token", leadCSRF)
	allowExternal.Set("email", external.Email)
	allowExternal.Set("role", "viewer")
	allowReq := httptest.NewRequest(http.MethodPost, subjectPath+"/members", strings.NewReader(allowExternal.Encode()))
	allowReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	allowReq.AddCookie(&http.Cookie{Name: "revues_session", Value: leadToken})
	allowRec := httptest.NewRecorder()
	handler.ServeHTTP(allowRec, allowReq)
	if allowRec.Code != http.StatusSeeOther {
		t.Fatalf("invite external status = %d, want %d; body=%s", allowRec.Code, http.StatusSeeOther, allowRec.Body.String())
	}
}
