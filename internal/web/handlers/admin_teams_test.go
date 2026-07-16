package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
)

func TestAdminTeams_MemberForbidden(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}

	member, err := st.UpsertGitHubUser(ctx, 10, "member", "member@example.com", "Member", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, member.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, member.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/teams", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestAdminTeams_OrgAdminOK(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}

	orgAdmin, err := st.UpsertGitHubUser(ctx, 11, "orgadmin", "orgadmin@example.com", "OrgAdmin", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, orgAdmin.ID, store.OrgRoleAdmin); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, orgAdmin.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/teams", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Créer une équipe") || !strings.Contains(body, `href="/admin/teams"`) {
		t.Errorf("body missing teams UI markers")
	}
}

func TestAdminTeams_CreateAddRemoveMember(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}

	admin, err := st.UpsertGitHubUser(ctx, 12, "admin", "admin@example.com", "Admin", "", auth.RoleAdmin)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(admin): %v", err)
	}
	alice, err := st.UpsertGitHubUser(ctx, 13, "alice", "alice@example.com", "Alice", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(alice): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, admin.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(admin): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, alice.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(alice): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, admin.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}
	csrf := auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes")

	form := url.Values{}
	form.Set("csrf_token", csrf)
	form.Set("name", "Qualité")
	form.Set("slug", "qualite")
	form.Set("description", "équipe QA")
	req := httptest.NewRequest(http.MethodPost, "/admin/teams", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("create status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	loc := rec.Header().Get("Location")
	if !strings.HasPrefix(loc, "/admin/teams/") {
		t.Fatalf("create location = %q", loc)
	}

	teams, err := st.ListOrganizationTeams(ctx)
	if err != nil || len(teams) != 1 {
		t.Fatalf("ListOrganizationTeams() = %+v, %v", teams, err)
	}
	teamID := teams[0].ID
	teamPath := "/admin/teams/" + strconv.FormatInt(teamID, 10)

	addForm := url.Values{}
	addForm.Set("csrf_token", csrf)
	addForm.Set("user_id", strconv.FormatInt(alice.ID, 10))
	addReq := httptest.NewRequest(http.MethodPost, teamPath+"/members", strings.NewReader(addForm.Encode()))
	addReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addReq.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	addRec := httptest.NewRecorder()
	handler.ServeHTTP(addRec, addReq)
	if addRec.Code != http.StatusSeeOther {
		t.Fatalf("add member status = %d, want %d", addRec.Code, http.StatusSeeOther)
	}

	showReq := httptest.NewRequest(http.MethodGet, teamPath, nil)
	showReq.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	showRec := httptest.NewRecorder()
	handler.ServeHTTP(showRec, showReq)
	if showRec.Code != http.StatusOK {
		t.Fatalf("show status = %d, want %d", showRec.Code, http.StatusOK)
	}
	body := showRec.Body.String()
	if !strings.Contains(body, "alice") || !strings.Contains(body, "alice@example.com") {
		t.Fatalf("show body missing member login/email: %s", body)
	}

	removeForm := url.Values{}
	removeForm.Set("csrf_token", csrf)
	removeForm.Set("user_id", strconv.FormatInt(alice.ID, 10))
	removeReq := httptest.NewRequest(http.MethodPost, teamPath+"/members/remove", strings.NewReader(removeForm.Encode()))
	removeReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	removeReq.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	removeRec := httptest.NewRecorder()
	handler.ServeHTTP(removeRec, removeReq)
	if removeRec.Code != http.StatusSeeOther {
		t.Fatalf("remove member status = %d, want %d", removeRec.Code, http.StatusSeeOther)
	}

	members, err := st.ListTeamMembers(ctx, teamID)
	if err != nil {
		t.Fatalf("ListTeamMembers(): %v", err)
	}
	if len(members) != 0 {
		t.Fatalf("members after remove = %+v", members)
	}
}
