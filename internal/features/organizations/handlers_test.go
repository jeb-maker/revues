package organizations_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/features/organizations"
	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
	appmiddleware "github.com/jeb-maker/revues/internal/web/middleware"
	"github.com/jeb-maker/revues/internal/web/templates"
)

const testSessionSecret = "test-secret-at-least-thirty-two-bytes"

func TestPostLoginRoute(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	st := store.New(db)

	userNoOrg, err := st.UpsertGitHubUser(ctx, 10, "solo", "solo@example.com", "Solo", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	userOne, err := st.UpsertGitHubUser(ctx, 11, "one", "one@example.com", "One", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	orgOne, err := st.CreateOrganization(ctx, "Solo Org", "solo-org", userOne.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, orgOne.ID, userOne.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}

	userMany, err := st.UpsertGitHubUser(ctx, 12, "many", "many@example.com", "Many", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	for _, slug := range []string{"alpha", "beta"} {
		org, err := st.CreateOrganization(ctx, strings.ToUpper(slug), slug, userMany.ID)
		if err != nil {
			t.Fatalf("CreateOrganization(%s): %v", slug, err)
		}
		if err := st.AddOrganizationMember(ctx, org.ID, userMany.ID, store.OrgRoleMember); err != nil {
			t.Fatalf("AddOrganizationMember(%s): %v", slug, err)
		}
	}

	tests := []struct {
		name       string
		userID     int64
		wantOrgArg int64
		wantPath   string
	}{
		{name: "zero organizations", userID: userNoOrg.ID, wantOrgArg: auth.SessionOrgPending, wantPath: "/org/new"},
		{name: "one organization", userID: userOne.ID, wantOrgArg: orgOne.ID, wantPath: "/projects"},
		{name: "many organizations", userID: userMany.ID, wantOrgArg: auth.SessionOrgPending, wantPath: "/org/select"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOrg, gotPath, err := organizations.PostLoginRoute(ctx, st, tt.userID)
			if err != nil {
				t.Fatalf("PostLoginRoute() error = %v", err)
			}
			if gotOrg != tt.wantOrgArg || gotPath != tt.wantPath {
				t.Fatalf("PostLoginRoute() = (%d, %q), want (%d, %q)", gotOrg, gotPath, tt.wantOrgArg, tt.wantPath)
			}
		})
	}
}

func TestCreateOrganizationSelfService(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	st := store.New(db)
	sessions := &auth.SessionManager{Store: st, SessionSecret: testSessionSecret}

	user, err := st.UpsertGitHubUser(ctx, 20, "creator", "creator@example.com", "Creator", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	token, _, err := sessions.CreateLoginSession(ctx, user.ID, auth.SessionOrgPending)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	tpl, err := templates.Parse()
	if err != nil {
		t.Fatalf("Parse templates: %v", err)
	}

	handler := &organizations.Organizations{Deps: organizations.Deps{
		Templates:     tpl,
		Store:         st,
		Sessions:      sessions,
		SessionSecret: testSessionSecret,
	}}

	router := chi.NewRouter()
	router.Use(appmiddleware.LoadUser(st))
	router.Use(appmiddleware.CSRF(testSessionSecret))
	router.Post("/org/new", handler.Create)

	form := url.Values{
		"name":       {"Acme Corp"},
		"slug":       {"acme-corp"},
		"csrf_token": {auth.CSRFToken(token, testSessionSecret)},
	}
	req := httptest.NewRequest(http.MethodPost, "/org/new", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if loc := rec.Header().Get("Location"); loc != "/projects" {
		t.Fatalf("Location = %q, want /projects", loc)
	}

	_, orgID, err := st.SessionByTokenHash(ctx, auth.HashToken(token))
	if err != nil {
		t.Fatalf("SessionByTokenHash(): %v", err)
	}
	org, err := st.OrganizationBySlug(ctx, "acme-corp")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}
	if orgID != org.ID {
		t.Fatalf("session org = %d, want %d", orgID, org.ID)
	}
	role, ok, err := st.OrganizationMemberRole(ctx, org.ID, user.ID)
	if err != nil || !ok || role != store.OrgRoleOwner {
		t.Fatalf("OrganizationMemberRole() = %q, %v, %v", role, ok, err)
	}
}

func TestSelectOrganization(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	st := store.New(db)
	sessions := &auth.SessionManager{Store: st, SessionSecret: testSessionSecret}

	user, err := st.UpsertGitHubUser(ctx, 30, "picker", "picker@example.com", "Picker", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	orgA, err := st.CreateOrganization(ctx, "Alpha", "alpha", user.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(alpha): %v", err)
	}
	orgB, err := st.CreateOrganization(ctx, "Beta", "beta", user.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(beta): %v", err)
	}
	for _, org := range []*store.Organization{orgA, orgB} {
		if err = st.AddOrganizationMember(ctx, org.ID, user.ID, store.OrgRoleMember); err != nil {
			t.Fatalf("AddOrganizationMember(): %v", err)
		}
	}

	token, _, err := sessions.CreateLoginSession(ctx, user.ID, auth.SessionOrgPending)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	tpl, err := templates.Parse()
	if err != nil {
		t.Fatalf("Parse templates: %v", err)
	}

	handler := &organizations.Organizations{Deps: organizations.Deps{
		Templates:     tpl,
		Store:         st,
		Sessions:      sessions,
		SessionSecret: testSessionSecret,
	}}

	router := chi.NewRouter()
	router.Use(appmiddleware.LoadUser(st))
	router.Use(appmiddleware.CSRF(testSessionSecret))
	router.Post("/org/select", handler.Select)

	form := url.Values{
		"organization_id": {strconv.FormatInt(orgB.ID, 10)},
		"csrf_token":      {auth.CSRFToken(token, testSessionSecret)},
	}

	req := httptest.NewRequest(http.MethodPost, "/org/select", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}

	_, orgID, err := st.SessionByTokenHash(ctx, auth.HashToken(token))
	if err != nil {
		t.Fatalf("SessionByTokenHash(): %v", err)
	}
	if orgID != orgB.ID {
		t.Fatalf("session org = %d, want %d", orgID, orgB.ID)
	}
}

func TestSwitchOrganization(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	st := store.New(db)
	sessions := &auth.SessionManager{Store: st, SessionSecret: testSessionSecret}

	user, err := st.UpsertGitHubUser(ctx, 40, "switcher", "switcher@example.com", "Switcher", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	orgA, err := st.CreateOrganization(ctx, "Alpha", "alpha-switch", user.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(alpha): %v", err)
	}
	orgB, err := st.CreateOrganization(ctx, "Beta", "beta-switch", user.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(beta): %v", err)
	}
	for _, org := range []*store.Organization{orgA, orgB} {
		if err := st.AddOrganizationMember(ctx, org.ID, user.ID, store.OrgRoleMember); err != nil {
			t.Fatalf("AddOrganizationMember(): %v", err)
		}
	}

	token, _, err := sessions.CreateLoginSession(ctx, user.ID, orgA.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	tpl, err := templates.Parse()
	if err != nil {
		t.Fatalf("Parse templates: %v", err)
	}

	handler := &organizations.Organizations{Deps: organizations.Deps{
		Templates:     tpl,
		Store:         st,
		Sessions:      sessions,
		SessionSecret: testSessionSecret,
	}}

	router := chi.NewRouter()
	router.Use(appmiddleware.LoadUser(st))
	router.Use(appmiddleware.LoadActiveOrganization(st))
	router.Use(appmiddleware.LoadHeaderData(st))
	router.Use(appmiddleware.CSRF(testSessionSecret))
	router.Post("/org/switch", handler.Switch)

	form := url.Values{
		"organization_id": {strconv.FormatInt(orgB.ID, 10)},
		"csrf_token":      {auth.CSRFToken(token, testSessionSecret)},
	}

	req := httptest.NewRequest(http.MethodPost, "/org/switch", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}

	_, orgID, err := st.SessionByTokenHash(ctx, auth.HashToken(token))
	if err != nil {
		t.Fatalf("SessionByTokenHash(): %v", err)
	}
	if orgID != orgB.ID {
		t.Fatalf("session org = %d, want %d", orgID, orgB.ID)
	}
}

func TestAcceptOrganizationInvitation(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	st := store.New(db)
	sessions := &auth.SessionManager{Store: st, SessionSecret: testSessionSecret}

	owner, err := st.UpsertGitHubUser(ctx, 50, "owner", "owner2@example.com", "Owner", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(owner): %v", err)
	}
	invitee, err := st.UpsertGitHubUser(ctx, 51, "invitee", "pending@example.com", "Pending", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(invitee): %v", err)
	}
	org, err := st.CreateOrganization(ctx, "Gamma", "gamma", owner.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(): %v", err)
	}
	if err := st.AddOrganizationMember(ctx, org.ID, owner.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(owner): %v", err)
	}
	orgCtx := orgctx.WithOrganizationID(ctx, org.ID)
	project, err := st.CreateProject(orgCtx, "App", "desc", owner.ID)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	if err := st.CreateOrganizationInvitation(ctx, "pending@example.com", org.ID, project.ID, "contributor"); err != nil {
		t.Fatalf("CreateOrganizationInvitation(): %v", err)
	}
	invites, err := st.ListPendingInvitationsByEmail(ctx, "pending@example.com")
	if err != nil || len(invites) != 1 {
		t.Fatalf("ListPendingInvitationsByEmail() = %v, %v", invites, err)
	}

	token, _, err := sessions.CreateLoginSession(ctx, invitee.ID, auth.SessionOrgPending)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	tpl, err := templates.Parse()
	if err != nil {
		t.Fatalf("Parse templates: %v", err)
	}

	handler := &organizations.Organizations{Deps: organizations.Deps{
		Templates:     tpl,
		Store:         st,
		Sessions:      sessions,
		SessionSecret: testSessionSecret,
	}}

	router := chi.NewRouter()
	router.Use(appmiddleware.LoadUser(st))
	router.Use(appmiddleware.CSRF(testSessionSecret))
	router.Post("/org/invitations/{id}/accept", handler.AcceptInvitation)

	path := "/org/invitations/" + strconv.FormatInt(invites[0].ID, 10) + "/accept"
	form := url.Values{"csrf_token": {auth.CSRFToken(token, testSessionSecret)}}
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d, body=%q", rec.Code, http.StatusSeeOther, rec.Body.String())
	}
	if loc := rec.Header().Get("Location"); loc != "/projects/"+strconv.FormatInt(project.ID, 10) {
		t.Fatalf("Location = %q, want project page", loc)
	}

	_, orgID, err := st.SessionByTokenHash(ctx, auth.HashToken(token))
	if err != nil {
		t.Fatalf("SessionByTokenHash(): %v", err)
	}
	if orgID != org.ID {
		t.Fatalf("session org = %d, want %d", orgID, org.ID)
	}
	role, ok, err := st.OrganizationMemberRole(ctx, org.ID, invitee.ID)
	if err != nil || !ok || role != store.OrgRoleMember {
		t.Fatalf("OrganizationMemberRole() = %q, %v, %v", role, ok, err)
	}
	projRole, ok, err := st.MemberRole(orgCtx, project.ID, invitee.ID)
	if err != nil || !ok || projRole != "contributor" {
		t.Fatalf("MemberRole() = %q, %v, %v", projRole, ok, err)
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/organizations.db")
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
