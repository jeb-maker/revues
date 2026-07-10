package handlers_test

import (
	"context"
	"github.com/jeb-maker/revues/internal/testutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

func TestAdminUsers_AddAndRemove(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}

	admin, err := st.UpsertGitHubUser(ctx, 99, "admin", "admin@example.com", "Admin", "", auth.RoleAdmin)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	if err := st.AddOrganizationMember(ctx, defaultOrg.ID, admin.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}
	if insertErr := st.InsertAllowedEmail(ctx, "admin@example.com", auth.RoleAdmin); insertErr != nil {
		t.Fatalf("InsertAllowedEmail(): %v", insertErr)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, admin.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	csrf := auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes")

	form := url.Values{}
	form.Set("csrf_token", csrf)
	form.Set("email", "new@example.com")
	form.Set("role", auth.RoleEditor)
	req := httptest.NewRequest(http.MethodPost, "/admin/users", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("add status = %d, want %d", rec.Code, http.StatusSeeOther)
	}

	role, ok, err := st.AllowedRole(ctx, "new@example.com")
	if err != nil || !ok || role != auth.RoleEditor {
		t.Fatalf("AllowedRole() = %q, %v, %v", role, ok, err)
	}

	removeForm := url.Values{}
	removeForm.Set("csrf_token", csrf)
	removeForm.Set("email", "new@example.com")
	removeReq := httptest.NewRequest(http.MethodPost, "/admin/users/remove", strings.NewReader(removeForm.Encode()))
	removeReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	removeReq.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	removeRec := httptest.NewRecorder()
	handler.ServeHTTP(removeRec, removeReq)
	if removeRec.Code != http.StatusSeeOther {
		t.Fatalf("remove status = %d, want %d", removeRec.Code, http.StatusSeeOther)
	}
}

func TestAdminUsers_ReaderForbidden(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}

	reader, err := st.UpsertGitHubUser(ctx, 1, "reader", "reader@example.com", "Reader", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	if err := st.AddOrganizationMember(ctx, defaultOrg.ID, reader.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, reader.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestAdminUsers_OrgAdminAllowed(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}

	orgAdmin, err := st.UpsertGitHubUser(ctx, 2, "orgadmin", "orgadmin@example.com", "OrgAdmin", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	if err := st.AddOrganizationMember(ctx, defaultOrg.ID, orgAdmin.ID, store.OrgRoleAdmin); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, orgAdmin.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
