package handlers_test

import (
	"context"
	"encoding/json"
	"github.com/jeb-maker/revues/internal/testutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/crypto"
	"github.com/jeb-maker/revues/internal/features/checklisttemplates"
	"github.com/jeb-maker/revues/internal/integrations/notion"
	"github.com/jeb-maker/revues/internal/store"
	appmiddleware "github.com/jeb-maker/revues/internal/web/middleware"
	viewtemplates "github.com/jeb-maker/revues/internal/web/templates"
)

func TestNotionImport_CreateTemplateV1(t *testing.T) {
	const dbID = "a1b2c3d4e5f6478990abcdef12345678"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": dbID, "title": []map[string]string{{"plain_text": "Checklist"}},
				"properties": map[string]any{"Name": map[string]string{"type": "title"}},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": []map[string]any{{"properties": map[string]any{
				"Name": map[string]any{"title": []map[string]string{{"plain_text": "Point A"}}},
			}}}, "has_more": false,
		})
	}))
	t.Cleanup(srv.Close)

	ctx := context.Background()
	db := mustMemoryDB(t)
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	key, _ := crypto.DecodeKey(config.TestEncryptionKey())
	_ = (&notion.Service{Store: st, EncryptionKey: key}).Save(ctx, notion.Config{APIToken: "secret"})
	tpl, _ := viewtemplates.Parse()
	secret := "test-secret-at-least-thirty-two-bytes"
	h := &checklisttemplates.ChecklistTemplates{
		Deps:          checklisttemplates.Deps{Templates: tpl, Store: st, SessionSecret: secret},
		EncryptionKey: key,
		NotionClient:  &notion.Client{HTTPClient: srv.Client(), APIBaseURL: srv.URL + "/v1"},
	}
	r := chi.NewRouter()
	r.Use(appmiddleware.LoadUser(st), appmiddleware.LoadActiveOrganization(st), appmiddleware.CSRF(secret))
	r.Post("/modeles/notion-import", h.NotionImport)

	lead, _ := st.UpsertGitHubUser(ctx, 40, "lead-notion", "lead-notion@example.com", "Lead", "", auth.RoleEditor)
	project, _ := st.CreateProject(ctx, "Import", "", lead.ID, []string{"qa"})
	sessions := &auth.SessionManager{Store: st, SessionSecret: secret}
	token, _, _ := sessions.CreateLoginSession(ctx, lead.ID, 0)
	form := url.Values{"csrf_token": {auth.CSRFToken(token, secret)}, "action": {"import"}, "database_id": {dbID}, "template_name": {"Modèle importé"}, "map_label": {"Name"}, "tags": {"qa"}}
	req := httptest.NewRequest(http.MethodPost, "/modeles/notion-import", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	templates, _ := st.ListChecklistTemplates(ctx, project.ID)
	if len(templates) != 1 || templates[0].LatestVersion != 1 {
		t.Fatalf("templates=%+v", templates)
	}
}

func TestNotionImport_RequiresEditorRole(t *testing.T) {
	handler, db := testRouterWithEncryptionKey(t, config.TestEncryptionKey())
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	lead, _ := st.UpsertGitHubUser(ctx, 41, "lead2", "lead2@example.com", "Lead", "", auth.RoleAdmin)
	_, _ = st.CreateProject(ctx, "Team", "", lead.ID, nil)
	key, _ := crypto.DecodeKey(config.TestEncryptionKey())
	_ = (&notion.Service{Store: st, EncryptionKey: key}).Save(ctx, notion.Config{APIToken: "secret"})

	// Unauthenticated users are redirected away from notion import.
	req := httptest.NewRequest(http.MethodGet, "/modeles/notion-import", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Errorf("status=%d want 302 redirect to login", rec.Code)
	}
}
