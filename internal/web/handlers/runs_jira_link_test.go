package handlers_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/features/projects"
	"github.com/jeb-maker/revues/internal/integrations/jira"
	"github.com/jeb-maker/revues/internal/store"
)

func TestIDOR_CrossProjectJiraLink(t *testing.T) {
	handler, db := testRouterWithEncryptionKey(t, config.TestEncryptionKey())
	ctx := context.Background()
	st := store.New(db)

	alice, err := st.UpsertGitHubUser(ctx, 60, "alice", "alice@example.com", "Alice", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(alice): %v", err)
	}
	bob, err := st.UpsertGitHubUser(ctx, 61, "bob", "bob@example.com", "Bob", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(bob): %v", err)
	}

	projectA, err := st.CreateProject(ctx, "A", "", alice.ID)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	projectB, err := st.CreateProject(ctx, "B", "", bob.ID)
	if err != nil {
		t.Fatalf("CreateProject(bob): %v", err)
	}

	templateA, _, err := st.CreateChecklistTemplate(ctx, projectA.ID, "Modèle", alice.ID, []store.TemplateItemInput{
		{Label: "Point", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	runA, err := st.CreateChecklistRun(ctx, projectA.ID, templateA.ID, "Revue A", alice.ID, sql.NullString{})
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}
	itemsA, err := st.ListRunItems(ctx, runA.ID)
	if err != nil || len(itemsA) != 1 {
		t.Fatalf("ListRunItems(): %v", err)
	}

	templateB, _, err := st.CreateChecklistTemplate(ctx, projectB.ID, "Modèle B", bob.ID, []store.TemplateItemInput{
		{Label: "Point B", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	runB, err := st.CreateChecklistRun(ctx, projectB.ID, templateB.ID, "Revue B", bob.ID, sql.NullString{})
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}
	itemsB, err := st.ListRunItems(ctx, runB.ID)
	if err != nil || len(itemsB) != 1 {
		t.Fatalf("ListRunItems(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	bobToken, _, err := sessions.CreateLoginSession(ctx, bob.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(bob): %v", err)
	}

	form := url.Values{}
	form.Set("csrf_token", auth.CSRFToken(bobToken, "test-secret-at-least-thirty-two-bytes"))
	form.Set("jira_issue", "REV-1")
	req := httptest.NewRequest(http.MethodPost, "/runs/"+strconv.FormatInt(runA.ID, 10)+"/items/"+strconv.FormatInt(itemsB[0].ID, 10)+"/jira-link", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: bobToken})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestJiraLink_ViewerForbidden(t *testing.T) {
	handler, db := testRouterWithEncryptionKey(t, config.TestEncryptionKey())
	ctx := context.Background()
	st := store.New(db)

	lead, err := st.UpsertGitHubUser(ctx, 70, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	viewer, err := st.UpsertGitHubUser(ctx, 71, "viewer", "viewer@example.com", "Viewer", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(viewer): %v", err)
	}

	project, err := st.CreateProject(ctx, "P", "", lead.ID)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	if err = st.AddProjectMember(ctx, project.ID, viewer.ID, projects.LocalRoleViewer); err != nil {
		t.Fatalf("AddProjectMember(): %v", err)
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
	items, err := st.ListRunItems(ctx, run.ID)
	if err != nil || len(items) != 1 {
		t.Fatalf("ListRunItems(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, viewer.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	form := url.Values{}
	form.Set("csrf_token", auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes"))
	form.Set("jira_issue", "REV-1")
	req := httptest.NewRequest(http.MethodPost, "/runs/"+strconv.FormatInt(run.ID, 10)+"/items/"+strconv.FormatInt(items[0].ID, 10)+"/jira-link", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestJiraLink_Success(t *testing.T) {
	jiraSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/3/issue/REV-99" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"key":"REV-99"}`))
	}))
	t.Cleanup(jiraSrv.Close)

	handler, db := testRouterWithEncryptionKey(t, config.TestEncryptionKey())
	ctx := context.Background()
	st := store.New(db)

	lead, err := st.UpsertGitHubUser(ctx, 80, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "P", "", lead.ID)
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
	items, err := st.ListRunItems(ctx, run.ID)
	if err != nil || len(items) != 1 {
		t.Fatalf("ListRunItems(): %v", err)
	}

	svc := &jira.Service{Store: st, EncryptionKey: mustTestKey(t)}
	saveErr := svc.Save(ctx, jira.Config{
		InstanceType: jira.InstanceCloud,
		BaseURL:      jiraSrv.URL,
		Email:        "user@example.com",
		APIToken:     "secret",
	})
	if saveErr != nil {
		t.Fatalf("Save jira config: %v", saveErr)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, lead.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	form := url.Values{}
	form.Set("csrf_token", auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes"))
	form.Set("jira_issue", "REV-99")
	req := httptest.NewRequest(http.MethodPost, "/runs/"+strconv.FormatInt(run.ID, 10)+"/items/"+strconv.FormatInt(items[0].ID, 10)+"/jira-link", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d, body=%q", rec.Code, http.StatusSeeOther, rec.Body.String())
	}

	link, err := st.IntegrationLinkByRunItemAndType(ctx, items[0].ID, store.IntegrationTypeJira)
	if err != nil {
		t.Fatalf("IntegrationLinkByRunItemAndType(): %v", err)
	}
	if link.ExternalKey != "REV-99" {
		t.Fatalf("ExternalKey = %q", link.ExternalKey)
	}
}

func mustTestKey(t *testing.T) []byte {
	t.Helper()
	key, err := config.Config{EncryptionKey: config.TestEncryptionKey()}.EncryptionKeyBytes()
	if err != nil {
		t.Fatalf("EncryptionKeyBytes(): %v", err)
	}
	return key
}
