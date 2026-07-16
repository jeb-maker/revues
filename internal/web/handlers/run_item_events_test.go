package handlers_test

import (
	"context"
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

func TestRunItemShow_DisplaysAuditHistory(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 90, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "P", "", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, "Modèle", lead.ID, nil, []store.TemplateItemInput{
		{Label: "Point audit", Required: true},
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
		t.Fatalf("ListRunItems(): %v", err)
	}
	if err = st.UpdateRunItemStatus(ctx, run.ID, runItems[0].ID, lead.ID, runs.StatusOK, ""); err != nil {
		t.Fatalf("UpdateRunItemStatus(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, lead.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/runs/"+strconv.FormatInt(run.ID, 10)+"/items/"+strconv.FormatInt(runItems[0].ID, 10), nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Historique") || !strings.Contains(body, "En attente") || !strings.Contains(body, ">OK<") {
		t.Fatal("expected audit history in item detail page")
	}
}

func TestIDOR_CrossProjectRunItemShow(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)

	alice, err := st.UpsertGitHubUser(ctx, 91, "alice", "alice@example.com", "Alice", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(alice): %v", err)
	}
	bob, err := st.UpsertGitHubUser(ctx, 92, "bob", "bob@example.com", "Bob", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(bob): %v", err)
	}
	ctx = testutil.SetupIsolatedOrg(ctx, st, "Alice Org", "alice-item-show", alice.ID)

	projectA, err := st.CreateProject(ctx, "Secret", "", alice.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, "Modèle", alice.ID, nil, []store.TemplateItemInput{
		{Label: "Secret point", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, projectA.ID, template.ID, alice.ID)
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}
	if err = st.StartRun(ctx, run.ID); err != nil {
		t.Fatalf("StartRun(): %v", err)
	}
	items, err := st.ListRunItems(ctx, run.ID)
	if err != nil || len(items) != 1 {
		t.Fatalf("ListRunItems(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	bobToken, _, err := sessions.CreateLoginSession(ctx, bob.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession(bob): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/runs/"+strconv.FormatInt(run.ID, 10)+"/items/"+strconv.FormatInt(items[0].ID, 10), nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: bobToken})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
