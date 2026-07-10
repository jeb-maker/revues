package handlers_test

import (
	"context"
	"encoding/json"
	"github.com/jeb-maker/revues/internal/testutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/crypto"
	"github.com/jeb-maker/revues/internal/features/checklisttemplates"
	"github.com/jeb-maker/revues/internal/features/projects"
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
	r.Post("/projects/{id}/templates/notion-import", h.NotionImport)

	lead, _ := st.UpsertGitHubUser(ctx, 40, "lead-notion", "lead-notion@example.com", "Lead", "", auth.RoleEditor)
	project, _ := st.CreateProject(ctx, "Import", "", lead.ID)
	sessions := &auth.SessionManager{Store: st, SessionSecret: secret}
	token, _, _ := sessions.CreateLoginSession(ctx, lead.ID, 0)
	form := url.Values{"csrf_token": {auth.CSRFToken(token, secret)}, "action": {"import"}, "database_id": {dbID}, "template_name": {"Modèle importé"}, "map_label": {"Name"}}
	req := httptest.NewRequest(http.MethodPost, "/projects/"+strconv.FormatInt(project.ID, 10)+"/templates/notion-import", strings.NewReader(form.Encode()))
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

func TestNotionImport_ContributorForbidden(t *testing.T) {
	handler, db := testRouterWithEncryptionKey(t, config.TestEncryptionKey())
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	lead, _ := st.UpsertGitHubUser(ctx, 41, "lead2", "lead2@example.com", "Lead", "", auth.RoleEditor)
	contrib, _ := st.UpsertGitHubUser(ctx, 42, "contrib2", "contrib2@example.com", "Contrib", "", auth.RoleEditor)
	project, _ := st.CreateProject(ctx, "Team", "", lead.ID)
	_ = st.AddProjectMember(ctx, project.ID, contrib.ID, projects.LocalRoleContributor)
	key, _ := crypto.DecodeKey(config.TestEncryptionKey())
	_ = (&notion.Service{Store: st, EncryptionKey: key}).Save(ctx, notion.Config{APIToken: "secret"})
	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, _ := sessions.CreateLoginSession(ctx, contrib.ID, 0)
	req := httptest.NewRequest(http.MethodGet, "/projects/"+strconv.FormatInt(project.ID, 10)+"/templates/notion-import", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status=%d want 404", rec.Code)
	}
}
