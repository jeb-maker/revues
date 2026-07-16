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
			wantBody:   []string{"Inviter", "/admin/users", "Mes sujets", "/admin/subjects", "Libellé sujet", "/admin/settings/labels", "Intégrations", "/admin/integrations"},
		},
		{
			name:       "org admin ok with integrations link",
			tokenKey:   "orgAdmin",
			wantStatus: http.StatusOK,
			wantBody:   []string{"Inviter", "/admin/users", "Mes sujets", "/admin/subjects", "Libellé sujet", "/admin/settings/labels", "Intégrations", "/admin/integrations"},
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
