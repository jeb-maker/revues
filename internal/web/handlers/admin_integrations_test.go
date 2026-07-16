package handlers_test

import (
	"context"
	"github.com/jeb-maker/revues/internal/testutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/crypto"
	adminsettings "github.com/jeb-maker/revues/internal/features/admin/settings"
	"github.com/jeb-maker/revues/internal/store"
)

func TestAdminIntegrations_ReaderForbidden(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	reader, err := st.UpsertGitHubUser(ctx, 1, "reader", "reader@example.com", "Reader", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, reader.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/integrations", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestAdminIntegrations_OrgAdminOK(t *testing.T) {
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
	if err = st.AddOrganizationMember(ctx, defaultOrg.ID, orgAdmin.ID, store.OrgRoleAdmin); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, orgAdmin.ID, defaultOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/integrations", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestAdminIntegrations_ShowStatus(t *testing.T) {
	handler, db := testRouterWithEncryptionKey(t, config.TestEncryptionKey())
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	adminUser, err := st.UpsertGitHubUser(ctx, 1, "admin", "admin@example.com", "Admin", "", auth.RoleAdmin)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	key, err := crypto.DecodeKey(config.TestEncryptionKey())
	if err != nil {
		t.Fatalf("DecodeKey(): %v", err)
	}
	settings := &adminsettings.SettingsService{Store: st, EncryptionKey: key}
	if saveErr := settings.SaveSMTP(ctx, adminsettings.SMTPConfig{
		Host: "smtp.example.com",
		Port: 587,
		From: "revues@example.com",
	}); saveErr != nil {
		t.Fatalf("SaveSMTP(): %v", saveErr)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, adminUser.ID, 0)
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
	for _, want := range []string{"organisation active", "SMTP", "Jira", "Notion", "Webhooks", "Activé", "Désactivé", "/admin/settings/smtp", "/admin/integrations/jira", "/admin/integrations/notion", "/admin/settings/webhooks"} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q", want)
		}
	}
	if !strings.Contains(body, "status-enabled") {
		t.Error("expected at least one integration marked enabled")
	}
	if !strings.Contains(body, "status-disabled") {
		t.Error("expected at least one integration marked disabled")
	}
}
