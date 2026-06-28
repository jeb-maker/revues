package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/items"
	"github.com/jeb-maker/revues/internal/store"
)

func TestDashboard_ShowsActiveRunProgress(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)

	lead, err := st.UpsertGitHubUser(ctx, 80, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "Alpha", "desc", lead.ID)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, project.ID, "Modèle", lead.ID, []store.TemplateItemInput{
		{Label: "A", Required: true},
		{Label: "B", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, "Sprint", lead.ID)
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}
	if err = st.StartRun(ctx, run.ID); err != nil {
		t.Fatalf("StartRun(): %v", err)
	}
	runItems, err := st.ListRunItems(ctx, run.ID)
	if err != nil || len(runItems) != 2 {
		t.Fatalf("ListRunItems() = %v, %v", runItems, err)
	}
	if err = st.UpdateRunItemStatus(ctx, run.ID, runItems[0].ID, lead.ID, items.StatusOK, ""); err != nil {
		t.Fatalf("UpdateRunItemStatus(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, lead.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Tableau de bord") {
		t.Fatal("expected dashboard title")
	}
	if !strings.Contains(body, "Sprint") {
		t.Fatal("expected active run title")
	}
	if !strings.Contains(body, "50%") {
		t.Fatal("expected run progress percent")
	}
	if !strings.Contains(body, "nav-tab is-active") || !strings.Contains(body, ">Projets</a>") {
		t.Fatal("expected active projects tab")
	}
}

func TestProjectShow_ShowsNokItems(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)

	lead, err := st.UpsertGitHubUser(ctx, 81, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "Beta", "", lead.ID)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, project.ID, "Modèle", lead.ID, []store.TemplateItemInput{
		{Label: "Blocage", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, "Revue", lead.ID)
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}
	if err = st.StartRun(ctx, run.ID); err != nil {
		t.Fatalf("StartRun(): %v", err)
	}
	runItems, err := st.ListRunItems(ctx, run.ID)
	if err != nil || len(runItems) != 1 {
		t.Fatalf("ListRunItems() = %v, %v", runItems, err)
	}
	if err = st.UpdateRunItemStatus(ctx, run.ID, runItems[0].ID, lead.ID, items.StatusNOK, "à corriger"); err != nil {
		t.Fatalf("UpdateRunItemStatus(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, lead.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/projects/"+strconv.FormatInt(project.ID, 10), nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Points bloquants") {
		t.Fatal("expected nok section")
	}
	if !strings.Contains(body, "Blocage") || !strings.Contains(body, "à corriger") {
		t.Fatal("expected nok item details")
	}
}

func TestTemplatesIndex_ListsVisibleTemplates(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)

	user, err := st.UpsertGitHubUser(ctx, 82, "user", "user@example.com", "User", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "Gamma", "", user.ID)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	if _, _, err = st.CreateChecklistTemplate(ctx, project.ID, "Checklist QA", user.ID, nil); err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, user.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/modeles", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Checklist QA") {
		t.Fatal("expected template in index")
	}
	if !strings.Contains(body, "Gamma") {
		t.Fatal("expected project name in index")
	}
}

func TestHome_RedirectsAuthenticatedUserToDashboard(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)

	user, err := st.UpsertGitHubUser(ctx, 83, "user", "user@example.com", "User", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, user.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if loc := rec.Header().Get("Location"); loc != "/projects" {
		t.Fatalf("Location = %q, want /projects", loc)
	}
}
