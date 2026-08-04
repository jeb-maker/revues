package organizations_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
	appweb "github.com/jeb-maker/revues/internal/web"
)

const hubTestSessionSecret = "test-secret-at-least-thirty-two-bytes"

func TestAdminHub_RBAC(t *testing.T) {
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

	globalAdmin, err := st.UpsertGitHubUser(ctx, 1, "globaladmin", "globaladmin@example.com", "Global", "", auth.RoleAdmin)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(globalAdmin): %v", err)
	}
	orgAdmin, err := st.UpsertGitHubUser(ctx, 2, "orgadmin", "orgadmin@example.com", "OrgAdmin", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(orgAdmin): %v", err)
	}
	member, err := st.UpsertGitHubUser(ctx, 3, "member", "member@example.com", "Member", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(member): %v", err)
	}

	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, globalAdmin.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(globalAdmin): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, orgAdmin.ID, store.OrgRoleAdmin); err != nil {
		t.Fatalf("AddOrganizationMember(orgAdmin): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, member.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(member): %v", err)
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
	tokens := map[string]string{}
	for key, userID := range map[string]int64{
		"globalAdmin": globalAdmin.ID,
		"orgAdmin":    orgAdmin.ID,
		"member":      member.ID,
	} {
		token, _, err := sessions.CreateLoginSession(ctx, userID, defaultOrg.ID)
		if err != nil {
			t.Fatalf("CreateLoginSession(%s): %v", key, err)
		}
		tokens[key] = token
	}

	tests := []struct {
		name       string
		tokenKey   string
		wantStatus int
		wantBody   []string
		notWant    []string
	}{
		{
			name:       "global admin ok with integrations link",
			tokenKey:   "globalAdmin",
			wantStatus: http.StatusOK,
			wantBody:   []string{"Inviter", "/admin/users", "Équipes", "/admin/teams", "Mes sujets", "/admin/subjects", "Libellé sujet", "/admin/settings/labels", "Politiques", "/admin/settings/policies", "Intégrations", "/admin/integrations"},
		},
		{
			name:       "org admin ok with integrations link",
			tokenKey:   "orgAdmin",
			wantStatus: http.StatusOK,
			wantBody:   []string{"Inviter", "/admin/users", "Équipes", "/admin/teams", "Mes sujets", "/admin/subjects", "Libellé sujet", "/admin/settings/labels", "Politiques", "/admin/settings/policies", "Intégrations", "/admin/integrations"},
		},
		{
			name:       "org member denied",
			tokenKey:   "member",
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			req.AddCookie(&http.Cookie{Name: "revues_session", Value: tokens[tt.tokenKey]})

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			body := rec.Body.String()
			for _, want := range tt.wantBody {
				if !strings.Contains(body, want) {
					t.Errorf("body missing %q", want)
				}
			}
			for _, not := range tt.notWant {
				if strings.Contains(body, not) {
					t.Errorf("body must not contain %q", not)
				}
			}
		})
	}
}

// Scenarios (B3 / decisions.md):
// 1. Solo (1 member, whitelist ≤1): header Organisation → hub Inviter + Mes sujets only.
// 2. After 2nd whitelist email: full hub + Organisation nav tab.
func TestAdminHub_SoloMinimalThenWhitelistUnlock(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/solo-hub.db", 0)
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err = store.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate(): %v", err)
	}

	st := store.New(db)
	owner, err := st.UpsertGitHubUser(ctx, 40, "soloowner", "soloowner@example.com", "Solo", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	org, err := st.CreateOrganization(ctx, "Solo Hub", "solo-hub", owner.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, org.ID, owner.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}
	orgCtx := testutil.OrgContext(ctx, org.ID)
	if err = st.InsertAllowedEmail(orgCtx, "soloowner@example.com", auth.RoleEditor); err != nil {
		t.Fatalf("InsertAllowedEmail(owner): %v", err)
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
	token, _, err := sessions.CreateLoginSession(ctx, owner.ID, org.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	getHub := func() (int, string) {
		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec.Code, rec.Body.String()
	}

	code, body := getHub()
	if code != http.StatusOK {
		t.Fatalf("solo hub status = %d", code)
	}
	for _, want := range []string{"Inviter", "/admin/users", "Mes sujets", "/admin/subjects"} {
		if !strings.Contains(body, want) {
			t.Errorf("solo hub missing %q", want)
		}
	}
	for _, not := range []string{">Équipes<", "/admin/teams", "Libellé", "Politiques", "Intégrations"} {
		if strings.Contains(body, not) {
			t.Errorf("solo hub must not contain %q", not)
		}
	}
	// Header ghost link (not primary nav tab) while Organisation nav is hidden.
	if !strings.Contains(body, `href="/admin"`) {
		t.Error("solo must expose Organisation header link to /admin")
	}
	if strings.Contains(body, `>Organisation</a>`) {
		t.Error("solo must not show Organisation primary nav tab")
	}

	if err = st.InsertAllowedEmail(orgCtx, "colleague@example.com", auth.RoleReader); err != nil {
		t.Fatalf("InsertAllowedEmail(colleague): %v", err)
	}

	code, body = getHub()
	if code != http.StatusOK {
		t.Fatalf("unlocked hub status = %d", code)
	}
	for _, want := range []string{
		"Inviter", "/admin/users",
		"Équipes", "/admin/teams",
		"Mes sujets", "/admin/subjects",
		"Libellé", "/admin/settings/labels",
		"Politiques", "/admin/settings/policies",
		"Intégrations", "/admin/integrations",
		">Organisation</a>",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("unlocked hub/nav missing %q", want)
		}
	}
	if !strings.Contains(body, "Un second e-mail a été autorisé") {
		t.Error("expected whitelist→org unlock flash after second email")
	}
}
