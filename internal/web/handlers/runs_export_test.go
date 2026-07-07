package handlers_test

import (
	"context"
	"database/sql"
	"encoding/csv"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/features/projects"
	runs "github.com/jeb-maker/revues/internal/features/runs"
	"github.com/jeb-maker/revues/internal/store"
)

func setupDoneRun(t *testing.T, st *store.Store, ctx context.Context, lead *store.User, project *store.Project) *store.ChecklistRun {
	t.Helper()

	template, _, err := st.CreateChecklistTemplate(ctx, project.ID, "Modèle", lead.ID, []store.TemplateItemInput{
		{Section: "S", Label: "Backup", Required: true},
		{Label: "Logs", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, "Revue Q2", lead.ID, sql.NullString{})
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
	if err = st.UpdateRunItemStatus(ctx, run.ID, runItems[1].ID, lead.ID, runs.StatusNOK, "Rotation manquante"); err != nil {
		t.Fatalf("UpdateRunItemStatus(): %v", err)
	}
	if err = st.CompleteRun(ctx, run.ID, "Clôturée"); err != nil {
		t.Fatalf("CompleteRun(): %v", err)
	}
	return run
}

func TestRuns_ExportCSV(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)

	lead, err := st.UpsertGitHubUser(ctx, 50, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "Alpha", "", lead.ID)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	run := setupDoneRun(t, st, ctx, lead, project)

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, lead.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/runs/"+strconv.FormatInt(run.ID, 10)+"/export.csv", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want text/csv; charset=utf-8", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, `attachment; filename="Revue-Q2.csv"`) {
		t.Fatalf("Content-Disposition = %q", cd)
	}

	records, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("len(records) = %d, want 3", len(records))
	}
	//nolint:misspell // French CSV column headers per issue #31
	wantHeader := []string{"projet", "revue", "date", "points", "statuts", "commentaires", "auteur"}
	for i, col := range wantHeader {
		if records[0][i] != col {
			t.Fatalf("header[%d] = %q, want %q", i, records[0][i], col)
		}
	}
	if records[1][0] != "Alpha" || records[1][1] != "Revue Q2" || records[1][3] != "Backup" {
		t.Fatalf("first data row = %v", records[1])
	}
	if records[1][4] != runs.StatusOK || records[1][6] != "lead" {
		t.Fatalf("first data row status/author = %v", records[1])
	}
	if records[2][3] != "Logs" || records[2][4] != runs.StatusNOK || records[2][5] != "Rotation manquante" {
		t.Fatalf("second data row = %v", records[2])
	}
	if records[1][2] == "" {
		t.Fatal("expected run date on export row")
	}
}

func TestRuns_ExportCSV_NotDone(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)

	lead, err := st.UpsertGitHubUser(ctx, 51, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "Alpha", "", lead.ID)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, project.ID, "Modèle", lead.ID, []store.TemplateItemInput{
		{Label: "Point", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, "Revue", lead.ID, sql.NullString{})
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}
	if err = st.StartRun(ctx, run.ID); err != nil {
		t.Fatalf("StartRun(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, lead.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/runs/"+strconv.FormatInt(run.ID, 10)+"/export.csv", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestRuns_ExportCSV_IDOR(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)

	alice, err := st.UpsertGitHubUser(ctx, 52, "alice", "alice@example.com", "Alice", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(alice): %v", err)
	}
	bob, err := st.UpsertGitHubUser(ctx, 53, "bob", "bob@example.com", "Bob", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(bob): %v", err)
	}

	projectA, err := st.CreateProject(ctx, "Secret", "", alice.ID)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	_, err = st.CreateProject(ctx, "Other", "", bob.ID)
	if err != nil {
		t.Fatalf("CreateProject(bob): %v", err)
	}
	run := setupDoneRun(t, st, ctx, alice, projectA)

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	bobToken, _, err := sessions.CreateLoginSession(ctx, bob.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(bob): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/runs/"+strconv.FormatInt(run.ID, 10)+"/export.csv", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: bobToken})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d (IDOR must return 404)", rec.Code, http.StatusNotFound)
	}
}

func TestRuns_ExportCSV_ViewerCanExport(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)

	lead, err := st.UpsertGitHubUser(ctx, 54, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(lead): %v", err)
	}
	viewer, err := st.UpsertGitHubUser(ctx, 55, "viewer", "viewer@example.com", "Viewer", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(viewer): %v", err)
	}
	project, err := st.CreateProject(ctx, "Team", "", lead.ID)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	if err = st.AddProjectMember(ctx, project.ID, viewer.ID, projects.LocalRoleViewer); err != nil {
		t.Fatalf("AddProjectMember(): %v", err)
	}
	run := setupDoneRun(t, st, ctx, lead, project)

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, viewer.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/runs/"+strconv.FormatInt(run.ID, 10)+"/export.csv", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRuns_ShowDoneIncludesExportButton(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)

	lead, err := st.UpsertGitHubUser(ctx, 56, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "Alpha", "", lead.ID)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	run := setupDoneRun(t, st, ctx, lead, project)

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, lead.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/runs/"+strconv.FormatInt(run.ID, 10), nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	exportURL := "/runs/" + strconv.FormatInt(run.ID, 10) + "/export.csv"
	if !strings.Contains(rec.Body.String(), exportURL) {
		t.Fatalf("expected export link %q in page", exportURL)
	}
	if !strings.Contains(rec.Body.String(), "Exporter CSV") {
		t.Fatal("expected export button label")
	}
}

func TestRuns_ShowInProgressOmitsExportButton(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)

	lead, err := st.UpsertGitHubUser(ctx, 57, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "Alpha", "", lead.ID)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, project.ID, "Modèle", lead.ID, []store.TemplateItemInput{
		{Label: "Point", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, "Revue", lead.ID, sql.NullString{})
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}
	if err = st.StartRun(ctx, run.ID); err != nil {
		t.Fatalf("StartRun(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, lead.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/runs/"+strconv.FormatInt(run.ID, 10), nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if strings.Contains(rec.Body.String(), "/export.csv") {
		t.Fatal("export link should not appear for in-progress run")
	}
}
