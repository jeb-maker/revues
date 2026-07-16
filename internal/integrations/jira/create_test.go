package jira_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/integrations/jira"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
)

func TestClientCreateIssueCloud(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/rest/api/3/issue/" {
			http.NotFound(w, r)
			return
		}
		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fields, _ := payload["fields"].(map[string]any)
		project, _ := fields["project"].(map[string]any)
		if project["key"] != "REV" {
			http.Error(w, "bad project", http.StatusBadRequest)
			return
		}
		if fields["summary"] != "Point nok" {
			http.Error(w, "bad summary", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"key":"REV-42"}`))
	}))
	t.Cleanup(srv.Close)

	client := &jira.Client{HTTPClient: srv.Client()}
	cfg := jira.Config{
		InstanceType: jira.InstanceCloud,
		BaseURL:      srv.URL,
		Email:        "user@example.com",
		APIToken:     "secret",
	}
	key, err := client.CreateIssue(context.Background(), cfg, jira.CreateIssueInput{
		ProjectKey:  "REV",
		IssueType:   "Task",
		Summary:     "Point nok",
		Description: "Description\nligne 2",
	})
	if err != nil {
		t.Fatalf("CreateIssue(): %v", err)
	}
	if key != "REV-42" {
		t.Fatalf("key = %q", key)
	}
}

func TestClientCreateIssueServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/rest/api/2/issue/" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"key":"SRV-7"}`))
	}))
	t.Cleanup(srv.Close)

	client := &jira.Client{HTTPClient: srv.Client()}
	cfg := jira.Config{
		InstanceType: jira.InstanceServer,
		BaseURL:      srv.URL,
		PAT:          "server-pat",
	}
	key, err := client.CreateIssue(context.Background(), cfg, jira.CreateIssueInput{
		ProjectKey:  "SRV",
		Summary:     "Issue",
		Description: "Details",
	})
	if err != nil {
		t.Fatalf("CreateIssue(): %v", err)
	}
	if key != "SRV-7" {
		t.Fatalf("key = %q", key)
	}
}

func TestClientCreateIssueAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "project not found", http.StatusBadRequest)
	}))
	t.Cleanup(srv.Close)

	client := &jira.Client{HTTPClient: srv.Client()}
	cfg := jira.Config{
		InstanceType: jira.InstanceCloud,
		BaseURL:      srv.URL,
		Email:        "user@example.com",
		APIToken:     "secret",
	}
	_, err := client.CreateIssue(context.Background(), cfg, jira.CreateIssueInput{
		ProjectKey: "MISSING",
		Summary:    "Issue",
	})
	if err == nil || !strings.Contains(err.Error(), "jira issue creation failed") {
		t.Fatalf("CreateIssue() = %v", err)
	}
}

func TestDefaultIssueContent(t *testing.T) {
	title, desc := jira.DefaultIssueContent(&store.RunItem{
		Label:   "Contrôle sécurité",
		Section: "Infra",
		Comment: "Port ouvert",
	}, jira.RunItemContext{
		SubjectName: "Alpha",
		RunTitle:    "Revue Q1",
		ItemURL:     "https://revues.example/runs/1/items/2",
	})
	if title != "Contrôle sécurité" {
		t.Fatalf("title = %q", title)
	}
	if !strings.Contains(desc, "Alpha") || !strings.Contains(desc, "Revue Q1") || !strings.Contains(desc, "Port ouvert") {
		t.Fatalf("description = %q", desc)
	}
}

func TestCreateServiceRunItem(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/rest/api/3/issue/":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"key":"REV-99"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	ctx := context.Background()
	svc, st := testJiraService(t)
	ctx = testutil.DefaultOrgContext(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 1, "lead", "lead@example.com", "Lead", "", "editor")
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	project, err := st.CreateProject(ctx, "Alpha", "", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}
	template, _, err := st.CreateChecklistTemplate(ctx, "Modèle", lead.ID, nil, []store.TemplateItemInput{
		{Label: "Point nok", Required: true},
	})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	run, err := st.CreateChecklistRun(ctx, project.ID, template.ID, lead.ID)
	if err != nil {
		t.Fatalf("CreateChecklistRun(): %v", err)
	}
	items, err := st.ListRunItems(ctx, run.ID)
	if err != nil || len(items) != 1 {
		t.Fatalf("ListRunItems(): %v", err)
	}
	if err = st.StartRun(ctx, run.ID); err != nil {
		t.Fatalf("StartRun(): %v", err)
	}
	if err = st.UpdateRunItemStatus(ctx, run.ID, items[0].ID, lead.ID, store.RunItemStatusNOK, "Problème"); err != nil {
		t.Fatalf("UpdateRunItemStatus(): %v", err)
	}

	jiraSvc := &jira.Service{Store: st, EncryptionKey: svc.EncryptionKey}
	if err = jiraSvc.Save(ctx, jira.Config{
		InstanceType: jira.InstanceCloud,
		BaseURL:      srv.URL,
		Email:        "user@example.com",
		APIToken:     "secret",
		ProjectKey:   "REV",
	}); err != nil {
		t.Fatalf("Save(): %v", err)
	}

	createSvc := &jira.CreateService{
		Store:         st,
		EncryptionKey: svc.EncryptionKey,
		Client:        &jira.Client{HTTPClient: srv.Client()},
	}
	link, err := createSvc.CreateRunItem(ctx, run.ID, items[0].ID, jira.CreateInput{}, jira.RunItemContext{
		SubjectName: project.Name,
		RunTitle:    store.RunDisplayLabel("Modèle", project.Name, run.CreatedAt, run.ID),
		ItemURL:     "https://revues.example/runs/1/items/2",
	})
	if err != nil {
		t.Fatalf("CreateRunItem(): %v", err)
	}
	if link.ExternalKey != "REV-99" {
		t.Fatalf("ExternalKey = %q", link.ExternalKey)
	}
}
