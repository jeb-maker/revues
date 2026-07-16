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

func TestRuns_CreateAndStart(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 30, "lead2", "lead2@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	subject, err := st.CreateSubject(ctx, "Alpha", "", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject(): %v", err)
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
	req := httptest.NewRequest(http.MethodPost, "/subjects/"+strconv.FormatInt(subject.ID, 10)+"/revues", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("create status = %d, want %d", rec.Code, http.StatusSeeOther)
	}

	runs, err := st.ListRunsBySubject(ctx, subject.ID)
	if err != nil || len(runs) != 1 {
		t.Fatalf("ListRunsBySubject() = %v, %v", runs, err)
	}
	if runs[0].Status != store.RunStatusInProgress {
		t.Fatalf("status = %q, want in_progress", runs[0].Status)
	}

	items, err := st.ListRunItems(ctx, runs[0].ID)
	if err != nil || len(items) != 1 || items[0].Label != "Check" {
		t.Fatalf("ListRunItems() = %v, %v", items, err)
	}

	// Start remains available for legacy drafts; already started runs are a no-op.
	startForm := url.Values{}
	startForm.Set("csrf_token", csrf)
	startReq := httptest.NewRequest(http.MethodPost, "/runs/"+strconv.FormatInt(runs[0].ID, 10)+"/start", strings.NewReader(startForm.Encode()))
	startReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	startReq.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	startRec := httptest.NewRecorder()
	handler.ServeHTTP(startRec, startReq)
	if startRec.Code != http.StatusSeeOther {
		t.Fatalf("start status = %d, want %d", startRec.Code, http.StatusSeeOther)
	}

	updated, err := st.RunByID(ctx, runs[0].ID)
	if err != nil {
		t.Fatalf("RunByID(): %v", err)
	}
	if updated.Status != store.RunStatusInProgress {
		t.Fatalf("status = %q, want in_progress", updated.Status)
	}
}

func TestRuns_WizardSubjectsRedirectsWhenSingleSubject(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 30, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	subject, err := st.CreateSubject(ctx, "Solo", "", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, lead.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/revues/nouvelle", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	want := "/subjects/" + strconv.FormatInt(subject.ID, 10) + "/modeles?for_run=1"
	if loc := rec.Header().Get("Location"); loc != want {
		t.Fatalf("Location = %q, want %q", loc, want)
	}
}

func TestRuns_CreateFromTemplateList(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 31, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	subject, err := st.CreateSubject(ctx, "Team", "", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject(): %v", err)
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
	req := httptest.NewRequest(http.MethodPost, "/subjects/"+strconv.FormatInt(subject.ID, 10)+"/revues", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if !strings.HasPrefix(rec.Header().Get("Location"), "/runs/") {
		t.Fatalf("Location = %q, want run redirect", rec.Header().Get("Location"))
	}

	listReq := httptest.NewRequest(http.MethodGet, "/subjects/"+strconv.FormatInt(subject.ID, 10)+"/modeles?for_run=1", nil)
	listReq.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	listRec := httptest.NewRecorder()
	handler.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("template list status = %d, want %d", listRec.Code, http.StatusOK)
	}
	if !strings.Contains(listRec.Body.String(), `action="/subjects/`+strconv.FormatInt(subject.ID, 10)+`/revues"`) {
		t.Fatal("expected POST create form on template list")
	}
}
