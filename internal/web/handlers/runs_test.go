package handlers_test

import (
	"context"
	"database/sql"
	"github.com/jeb-maker/revues/internal/testutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/features/projects"
	"github.com/jeb-maker/revues/internal/store"
)

func TestIDOR_CrossProjectRun(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	alice, err := st.UpsertGitHubUser(ctx, 10, "alice", "alice@example.com", "Alice", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(alice): %v", err)
	}
	bob, err := st.UpsertGitHubUser(ctx, 11, "bob", "bob@example.com", "Bob", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(bob): %v", err)
	}

	projectA, err := st.CreateProject(ctx, "Secret", "", alice.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	_, err = st.CreateProject(ctx, "Other", "", bob.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(bob): %v", err)
	}

	template, _, err := st.CreateChecklistTemplate(ctx, "Modèle", alice.ID, nil, []store.TemplateItemInput{
		{Label: "Point", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, projectA.ID, template.ID, "Revue secrète", alice.ID, sql.NullString{})
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	bobToken, _, err := sessions.CreateLoginSession(ctx, bob.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(bob): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/runs/"+strconv.FormatInt(run.ID, 10), nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: bobToken})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d (IDOR must return 404)", rec.Code, http.StatusNotFound)
	}
}

func TestRuns_ViewerCannotLaunch(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 20, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(lead): %v", err)
	}
	viewer, err := st.UpsertGitHubUser(ctx, 21, "viewer", "viewer@example.com", "Viewer", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(viewer): %v", err)
	}

	project, err := st.CreateProject(ctx, "Team", "", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	if err = st.AddProjectMember(ctx, project.ID, viewer.ID, projects.LocalRoleViewer); err != nil {
		t.Fatalf("AddProjectMember(): %v", err)
	}

	template, _, err := st.CreateChecklistTemplate(ctx, "Modèle", lead.ID, nil, []store.TemplateItemInput{
		{Label: "Point", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, viewer.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	form := url.Values{}
	form.Set("csrf_token", auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes"))
	form.Set("template_id", strconv.FormatInt(template.ID, 10))
	form.Set("title", "Nope")
	req := httptest.NewRequest(http.MethodPost, "/projects/"+strconv.FormatInt(project.ID, 10)+"/runs", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestRuns_WizardCreateAndStart(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 30, "lead2", "lead2@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "Alpha", "", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, "Modèle", lead.ID, nil, []store.TemplateItemInput{
		{Section: "S", Label: "Check", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, lead.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}
	csrf := auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes")

	form := url.Values{}
	form.Set("csrf_token", csrf)
	form.Set("template_id", strconv.FormatInt(template.ID, 10))
	form.Set("title", "Revue sprint")
	req := httptest.NewRequest(http.MethodPost, "/projects/"+strconv.FormatInt(project.ID, 10)+"/runs", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("create status = %d, want %d", rec.Code, http.StatusSeeOther)
	}

	run, err := st.ListRunsByProject(ctx, project.ID)
	if err != nil || len(run) != 1 {
		t.Fatalf("ListRunsByProject() = %v, %v", run, err)
	}
	if run[0].Status != store.RunStatusDraft {
		t.Fatalf("status = %q, want draft", run[0].Status)
	}

	items, err := st.ListRunItems(ctx, run[0].ID)
	if err != nil || len(items) != 1 || items[0].Label != "Check" {
		t.Fatalf("ListRunItems() = %v, %v", items, err)
	}

	startForm := url.Values{}
	startForm.Set("csrf_token", csrf)
	startReq := httptest.NewRequest(http.MethodPost, "/runs/"+strconv.FormatInt(run[0].ID, 10)+"/start", strings.NewReader(startForm.Encode()))
	startReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	startReq.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	startRec := httptest.NewRecorder()
	handler.ServeHTTP(startRec, startReq)
	if startRec.Code != http.StatusSeeOther {
		t.Fatalf("start status = %d, want %d", startRec.Code, http.StatusSeeOther)
	}

	updated, err := st.RunByID(ctx, run[0].ID)
	if err != nil {
		t.Fatalf("RunByID(): %v", err)
	}
	if updated.Status != store.RunStatusInProgress {
		t.Fatalf("status = %q, want in_progress", updated.Status)
	}
}

func TestRuns_CreateWithDueDate(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 31, "lead3", "lead3@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "Beta", "", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, "Modèle", lead.ID, nil, []store.TemplateItemInput{
		{Label: "Point", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, lead.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}
	csrf := auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes")

	form := url.Values{}
	form.Set("csrf_token", csrf)
	form.Set("template_id", strconv.FormatInt(template.ID, 10))
	form.Set("title", "Revue avec échéance")
	form.Set("due_date", "2026-08-01")
	req := httptest.NewRequest(http.MethodPost, "/projects/"+strconv.FormatInt(project.ID, 10)+"/runs", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("create status = %d, want %d", rec.Code, http.StatusSeeOther)
	}

	runs, err := st.ListRunsByProject(ctx, project.ID)
	if err != nil || len(runs) != 1 {
		t.Fatalf("ListRunsByProject() = %v, %v", runs, err)
	}
	if !runs[0].DueDate.Valid || runs[0].DueDate.String != "2026-08-01T00:00:00Z" {
		t.Fatalf("due_date = %+v", runs[0].DueDate)
	}
}
