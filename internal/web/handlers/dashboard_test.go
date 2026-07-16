package handlers_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/testutil"

	"github.com/jeb-maker/revues/internal/auth"
	runs "github.com/jeb-maker/revues/internal/features/runs"
	"github.com/jeb-maker/revues/internal/store"
)

func TestDashboard_ShowsActiveRunProgress(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 80, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "Alpha", "desc", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, "Modèle", lead.ID, nil, []store.TemplateItemInput{
		{Label: "A", Required: true},
		{Label: "B", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, lead.ID)
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
	if err = st.UpdateRunItemStatus(ctx, run.ID, runItems[0].ID, lead.ID, runs.StatusOK, ""); err != nil {
		t.Fatalf("UpdateRunItemStatus(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, lead.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/revues", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Alpha") {
		t.Fatal("expected active run display label")
	}
	if !strings.Contains(body, "50%") {
		t.Fatal("expected run progress percent")
	}
	if !strings.Contains(body, `href="/revues"`) || !strings.Contains(body, `aria-current="page"`) {
		t.Fatal("expected active Revues nav link on revues home")
	}
}

func TestDashboard_ShowsRecentCompletedRuns(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 86, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "Zeta", "desc", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, "Modèle", lead.ID, nil, []store.TemplateItemInput{
		{Label: "A", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, lead.ID)
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
	if err = st.UpdateRunItemStatus(ctx, run.ID, runItems[0].ID, lead.ID, runs.StatusOK, ""); err != nil {
		t.Fatalf("UpdateRunItemStatus(): %v", err)
	}
	if err = st.CompleteRun(ctx, run.ID, ""); err != nil {
		t.Fatalf("CompleteRun(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, lead.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/revues", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if strings.Contains(body, "Terminées récemment") {
		t.Fatal("unexpected legacy completed runs section")
	}
	if !strings.Contains(body, "Zeta") {
		t.Fatal("expected completed run display label")
	}
	if !strings.Contains(body, "100%") {
		t.Fatal("expected completed run progress")
	}
	if strings.Contains(body, "Aucune revue active") {
		t.Fatal("unexpected legacy empty active runs message")
	}
	if !strings.Contains(body, `id="filter-query"`) {
		t.Fatal("expected search input on revues page")
	}
	if !strings.Contains(body, ">Filtrer<") {
		t.Fatal("expected filter submit button")
	}
}

func TestDashboard_ShowsNewRunAsInProgress(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 87, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "DraftCo", "desc", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, "Modèle", lead.ID, nil, []store.TemplateItemInput{
		{Label: "A", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	if _, err = st.CreateChecklistRun(ctx, project.ID, template.ID, lead.ID); err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, lead.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/revues", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "DraftCo") {
		t.Fatal("expected run display label on dashboard")
	}
	if !strings.Contains(body, "En cours") {
		t.Fatal("expected in-progress status label on dashboard")
	}
	if strings.Contains(body, "Brouillon") {
		t.Fatal("new runs must not appear as draft")
	}
}

func TestDashboard_ShowsRunDueDate(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 84, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "Delta", "desc", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, "Modèle", lead.ID, nil, []store.TemplateItemInput{
		{Label: "A", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, lead.ID)
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}
	if err = st.SetRunDueDate(ctx, run.ID, sql.NullString{String: "2026-07-01T00:00:00Z", Valid: true}); err != nil {
		t.Fatalf("SetRunDueDate(): %v", err)
	}
	if err = st.StartRun(ctx, run.ID); err != nil {
		t.Fatalf("StartRun(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, lead.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/revues", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "01/07/2026") {
		t.Fatal("expected formatted due date on dashboard")
	}
	if !strings.Contains(body, "Modèle") {
		t.Fatal("expected computed run label on dashboard")
	}
}

func TestProjectShow_ShowsRunDueDate(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 85, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "Epsilon", "", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, "Modèle", lead.ID, nil, nil)
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	if _, err = st.CreateChecklistRun(ctx, project.ID, template.ID, lead.ID); err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}
	runs, err := st.ListRunsBySubject(ctx, project.ID)
	if err != nil || len(runs) == 0 {
		t.Fatalf("ListRunsBySubject() = %v, %v", runs, err)
	}
	if err = st.SetRunDueDate(ctx, runs[0].ID, sql.NullString{String: "2026-09-30T00:00:00Z", Valid: true}); err != nil {
		t.Fatalf("SetRunDueDate(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, lead.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/subjects/"+strconv.FormatInt(project.ID, 10), nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "30/09/2026") {
		t.Fatal("expected formatted due date on project page")
	}
	if !strings.Contains(body, "Modèle") {
		t.Fatal("expected computed run label on project page")
	}
}

func TestProjectShow_ShowsNokItems(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 81, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "Beta", "", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, "Modèle", lead.ID, nil, []store.TemplateItemInput{
		{Label: "Blocage", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, lead.ID)
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
	if err = st.UpdateRunItemStatus(ctx, run.ID, runItems[0].ID, lead.ID, runs.StatusNOK, "à corriger"); err != nil {
		t.Fatalf("UpdateRunItemStatus(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, lead.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/subjects/"+strconv.FormatInt(project.ID, 10), nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Points non validés") {
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
	ctx = testutil.DefaultOrgContext(ctx, st)

	user, err := st.UpsertGitHubUser(ctx, 82, "user", "user@example.com", "User", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	_, err = st.CreateProject(ctx, "Gamma", "", user.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	if _, _, err = st.CreateChecklistTemplate(ctx, "Checklist QA", user.ID, nil, nil); err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, user.ID, 0)
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
	if !strings.Contains(body, "tous sujets") {
		t.Fatal("expected global template marker in index")
	}
	if !strings.Contains(body, "list-toolbar") {
		t.Fatal("expected list toolbar on templates page")
	}
}

func TestHome_RedirectsAuthenticatedUserToDashboard(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	user, err := st.UpsertGitHubUser(ctx, 83, "user", "user@example.com", "User", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, user.ID, 0)
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
	if loc := rec.Header().Get("Location"); loc != "/revues" {
		t.Fatalf("Location = %q, want /revues", loc)
	}
}
