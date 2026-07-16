package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/crypto"
	adminsettings "github.com/jeb-maker/revues/internal/features/admin/settings"
	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
)

const integrationsOrgSessionSecret = "test-secret-at-least-thirty-two-bytes"

func TestAdminIntegrations_OrgOwnerSaveSMTP(t *testing.T) {
	handler, db := testRouterWithEncryptionKey(t, config.TestEncryptionKey())
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}
	owner, err := st.UpsertGitHubUser(ctx, 10, "owner", "owner@example.com", "Owner", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, owner.ID, store.OrgRoleOwner); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: integrationsOrgSessionSecret}
	token, _, err := sessions.CreateLoginSession(ctx, owner.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}
	csrf := auth.CSRFToken(token, integrationsOrgSessionSecret)

	form := url.Values{}
	form.Set("csrf_token", csrf)
	form.Set("action", "save")
	form.Set("host", "smtp.owner.example.com")
	form.Set("port", "587")
	form.Set("from", "owner@example.com")
	form.Set("password", "secret")

	req := httptest.NewRequest(http.MethodPost, "/admin/settings/smtp", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d, body=%q", rec.Code, http.StatusSeeOther, rec.Body.String())
	}

	key, err := crypto.DecodeKey(config.TestEncryptionKey())
	if err != nil {
		t.Fatalf("DecodeKey(): %v", err)
	}
	svc := &adminsettings.SettingsService{Store: st, EncryptionKey: key}
	cfg, ok, err := svc.LoadSMTP(orgctx.WithOrganizationID(ctx, defaultOrg.ID))
	if err != nil || !ok {
		t.Fatalf("LoadSMTP() = ok=%v err=%v", ok, err)
	}
	if cfg.Host != "smtp.owner.example.com" {
		t.Fatalf("LoadSMTP().Host = %q", cfg.Host)
	}
}

func TestAdminIntegrations_OrgMemberPOSTForbidden(t *testing.T) {
	handler, db := testRouterWithEncryptionKey(t, config.TestEncryptionKey())
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}
	member, err := st.UpsertGitHubUser(ctx, 11, "member", "member@example.com", "Member", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, member.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: integrationsOrgSessionSecret}
	token, _, err := sessions.CreateLoginSession(ctx, member.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}
	csrf := auth.CSRFToken(token, integrationsOrgSessionSecret)

	form := url.Values{}
	form.Set("csrf_token", csrf)
	form.Set("action", "save")
	form.Set("host", "smtp.evil.example.com")
	form.Set("port", "587")
	form.Set("from", "evil@example.com")

	req := httptest.NewRequest(http.MethodPost, "/admin/settings/smtp", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestAdminIntegrations_CrossOrgIsolation(t *testing.T) {
	handler, db := testRouterWithEncryptionKey(t, config.TestEncryptionKey())
	ctx := context.Background()
	st := store.New(db)
	_ = testutil.DefaultOrgContext(ctx, st)

	orgA, err := st.CreateOrganization(ctx, "Org A", "org-a", 0)
	if err != nil {
		t.Fatalf("CreateOrganization(A): %v", err)
	}
	orgB, err := st.CreateOrganization(ctx, "Org B", "org-b", 0)
	if err != nil {
		t.Fatalf("CreateOrganization(B): %v", err)
	}

	admin, err := st.UpsertGitHubUser(ctx, 12, "multi", "multi@example.com", "Multi", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	for _, orgID := range []int64{orgA.ID, orgB.ID} {
		if err = st.AddOrganizationMember(ctx, orgID, admin.ID, store.OrgRoleAdmin); err != nil {
			t.Fatalf("AddOrganizationMember(%d): %v", orgID, err)
		}
	}

	key, err := crypto.DecodeKey(config.TestEncryptionKey())
	if err != nil {
		t.Fatalf("DecodeKey(): %v", err)
	}
	svc := &adminsettings.SettingsService{Store: st, EncryptionKey: key}
	if err = svc.SaveSMTP(orgctx.WithOrganizationID(ctx, orgA.ID), adminsettings.SMTPConfig{
		Host: "smtp.a.example.com",
		Port: 587,
		From: "a@example.com",
	}); err != nil {
		t.Fatalf("SaveSMTP(A): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: integrationsOrgSessionSecret}
	tokenB, _, err := sessions.CreateLoginSession(ctx, admin.ID, orgB.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(B): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/settings/smtp", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: tokenB})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if strings.Contains(body, "smtp.a.example.com") {
		t.Fatal("org B session must not see org A SMTP host")
	}

	_, ok, err := svc.LoadSMTP(orgctx.WithOrganizationID(ctx, orgB.ID))
	if err != nil {
		t.Fatalf("LoadSMTP(B): %v", err)
	}
	if ok {
		t.Fatal("org B should have no SMTP config")
	}

	tokenA, _, err := sessions.CreateLoginSession(ctx, admin.ID, orgA.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(A): %v", err)
	}
	reqA := httptest.NewRequest(http.MethodGet, "/admin/settings/smtp", nil)
	reqA.AddCookie(&http.Cookie{Name: "revues_session", Value: tokenA})
	recA := httptest.NewRecorder()
	handler.ServeHTTP(recA, reqA)
	if recA.Code != http.StatusOK {
		t.Fatalf("org A status = %d, want %d", recA.Code, http.StatusOK)
	}
	if !strings.Contains(recA.Body.String(), "smtp.a.example.com") {
		t.Fatal("org A session must see its SMTP host")
	}
}

func TestAdminIntegrations_OrgAdminSeesNav(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("OrganizationBySlug(): %v", err)
	}
	orgAdmin, err := st.UpsertGitHubUser(ctx, 13, "navadmin", "navadmin@example.com", "NavAdmin", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, orgAdmin.ID, store.OrgRoleAdmin); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: integrationsOrgSessionSecret}
	token, _, err := sessions.CreateLoginSession(ctx, orgAdmin.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/integrations", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	for _, want := range []string{
		"organisation active",
		"/admin/settings/smtp",
		"/admin/integrations/jira",
		"/admin/integrations/notion",
		"/admin/settings/webhooks",
		`aria-current="page"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q", want)
		}
	}
}
