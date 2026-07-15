package organizations_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
	appweb "github.com/jeb-maker/revues/internal/web"
)

func TestSubjectLabelPreset_CibleShownInAdminNav(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/test.db", 0)
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err = store.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate(): %v", err)
	}

	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}

	orgAdmin, err := st.UpsertGitHubUser(ctx, 10, "orgadmin", "orgadmin@example.com", "OrgAdmin", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, orgAdmin.ID, store.OrgRoleAdmin); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}
	if err = st.UpdateOrganizationUISubjectLabel(ctx, defaultOrg.ID, store.UISubjectLabelCible); err != nil {
		t.Fatalf("UpdateOrganizationUISubjectLabel(): %v", err)
	}

	handler, _, err := appweb.NewRouter(appweb.Deps{
		Config: config.Config{
			Addr:          ":8080",
			BaseURL:       "http://example.com",
			SessionSecret: hubTestSessionSecret,
			Env:           "development",
		},
		DB: db,
	})
	if err != nil {
		t.Fatalf("NewRouter(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: hubTestSessionSecret}
	token, _, err := sessions.CreateLoginSession(ctx, orgAdmin.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/subjects", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Cibles") {
		t.Fatalf("admin nav missing preset plural %q in body", "Cibles")
	}
	if strings.Contains(body, ">Sujets<") {
		t.Fatalf("admin nav still shows default Sujets with cible preset")
	}
}

func TestSubjectLabelsSave_UpdatesPreset(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/test.db", 0)
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err = store.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate(): %v", err)
	}

	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}

	orgAdmin, err := st.UpsertGitHubUser(ctx, 11, "labeladmin", "labeladmin@example.com", "LabelAdmin", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, orgAdmin.ID, store.OrgRoleAdmin); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}

	handler, _, err := appweb.NewRouter(appweb.Deps{
		Config: config.Config{
			Addr:          ":8080",
			BaseURL:       "http://example.com",
			SessionSecret: hubTestSessionSecret,
			Env:           "development",
		},
		DB: db,
	})
	if err != nil {
		t.Fatalf("NewRouter(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: hubTestSessionSecret}
	token, _, err := sessions.CreateLoginSession(ctx, orgAdmin.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	form := url.Values{
		"csrf_token":       {auth.CSRFToken(token, hubTestSessionSecret)},
		"ui_subject_label": {store.UISubjectLabelAsset},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/settings/labels", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusFound, rec.Body.String())
	}
	if loc := rec.Header().Get("Location"); !strings.HasPrefix(loc, "/admin/settings/labels") {
		t.Fatalf("Location = %q", loc)
	}

	got, err := st.OrganizationByID(ctx, defaultOrg.ID)
	if err != nil {
		t.Fatalf("OrganizationByID(): %v", err)
	}
	if got.UISubjectLabel != store.UISubjectLabelAsset {
		t.Fatalf("UISubjectLabel = %q, want %q", got.UISubjectLabel, store.UISubjectLabelAsset)
	}
}
