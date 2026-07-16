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

func TestLeadPoliciesSave_UpdatesFlags(t *testing.T) {
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

	orgAdmin, err := st.UpsertGitHubUser(ctx, 21, "poladmin", "poladmin@example.com", "PolAdmin", "", auth.RoleEditor)
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

	getReq := httptest.NewRequest(http.MethodGet, "/admin/settings/policies", nil)
	getReq.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	getRec := httptest.NewRecorder()
	handler.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("GET status = %d", getRec.Code)
	}
	if !strings.Contains(getRec.Body.String(), "Délégation aux leads") {
		t.Fatalf("GET body missing policies form")
	}

	form := url.Values{
		"csrf_token":                 {auth.CSRFToken(token, hubTestSessionSecret)},
		"leads_may_invite_externals": {"on"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/settings/policies", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusFound, rec.Body.String())
	}
	if loc := rec.Header().Get("Location"); !strings.HasPrefix(loc, "/admin/settings/policies") {
		t.Fatalf("Location = %q", loc)
	}

	got, err := st.OrganizationByID(ctx, defaultOrg.ID)
	if err != nil {
		t.Fatalf("OrganizationByID(): %v", err)
	}
	if got.LeadsMayAssignTeams || got.LeadsMayInviteMembers || !got.LeadsMayInviteExternals {
		t.Fatalf("policies = assign=%v invite=%v externals=%v, want false/false/true",
			got.LeadsMayAssignTeams, got.LeadsMayInviteMembers, got.LeadsMayInviteExternals)
	}
}

func TestLeadPolicies_MemberForbidden(t *testing.T) {
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

	member, err := st.UpsertGitHubUser(ctx, 22, "polmember", "polmember@example.com", "PolMember", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, member.ID, store.OrgRoleMember); err != nil {
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
	token, _, err := sessions.CreateLoginSession(ctx, member.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/settings/policies", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}
