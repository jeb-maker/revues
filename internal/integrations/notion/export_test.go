package notion_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/integrations/notion"
	"github.com/jeb-maker/revues/internal/items"
	"github.com/jeb-maker/revues/internal/store"
)

func TestCreateReviewPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"page-id","url":"https://notion.so/page-id"}`))
	}))
	t.Cleanup(srv.Close)
	client := &notion.Client{HTTPClient: srv.Client(), APIBaseURL: srv.URL + "/v1"}
	result, err := client.CreateReviewPage(context.Background(), notion.Config{APIToken: "tok", DefaultDatabaseID: "abc123def4567890abc123def4567890"}, notion.CreatePageInput{DatabaseID: "abc123def4567890abc123def4567890", Title: "Revue"})
	if err != nil || result.URL != "https://notion.so/page-id" {
		t.Fatalf("CreateReviewPage() = %+v err=%v", result, err)
	}
}

func TestExportServiceExportRun(t *testing.T) {
	ctx := context.Background()
	svc := testNotionService(t)
	_ = svc.Save(ctx, notion.Config{APIToken: "tok", DefaultDatabaseID: "abc123def4567890abc123def4567890"})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"page-id","url":"https://notion.so/revue-export"}`))
	}))
	t.Cleanup(srv.Close)
	st := svc.Store
	lead, _ := st.UpsertGitHubUser(ctx, 1, "lead", "lead@example.com", "Lead", "", "editor")
	project, _ := st.CreateProject(ctx, "Alpha", "", lead.ID)
	template, _, _ := st.CreateChecklistTemplate(ctx, project.ID, "Modèle", lead.ID, []store.TemplateItemInput{{Label: "Backup", Required: true}})
	run, _ := st.CreateChecklistRun(ctx, project.ID, template.ID, "Revue Q1", lead.ID, sql.NullString{})
	_ = st.StartRun(ctx, run.ID)
	runItems, _ := st.ListRunItems(ctx, run.ID)
	_ = st.UpdateRunItemStatus(ctx, run.ID, runItems[0].ID, lead.ID, items.StatusOK, "")
	_ = st.CompleteRun(ctx, run.ID, "Clôturée")
	exportSvc := &notion.ExportService{Store: st, EncryptionKey: svc.EncryptionKey, Client: &notion.Client{HTTPClient: srv.Client(), APIBaseURL: srv.URL + "/v1"}, BaseURL: "http://example.com"}
	url, err := exportSvc.ExportRun(ctx, run.ID)
	if err != nil || url != "https://notion.so/revue-export" {
		t.Fatalf("ExportRun() = %q err=%v", url, err)
	}
}

func TestExportServiceAlreadyExported(t *testing.T) {
	ctx := context.Background()
	svc := testNotionService(t)
	_ = svc.Save(ctx, notion.Config{APIToken: "tok", DefaultDatabaseID: "abc123def4567890abc123def4567890"})
	st := svc.Store
	lead, _ := st.UpsertGitHubUser(ctx, 2, "lead", "lead@example.com", "Lead", "", "editor")
	project, _ := st.CreateProject(ctx, "Alpha", "", lead.ID)
	template, _, _ := st.CreateChecklistTemplate(ctx, project.ID, "Modèle", lead.ID, []store.TemplateItemInput{{Label: "P", Required: true}})
	run, _ := st.CreateChecklistRun(ctx, project.ID, template.ID, "Revue", lead.ID, sql.NullString{})
	_ = st.StartRun(ctx, run.ID)
	_ = st.CompleteRun(ctx, run.ID, "Note")
	_ = st.SetRunNotionURL(ctx, run.ID, "https://notion.so/existing")
	_, err := (&notion.ExportService{Store: st, EncryptionKey: svc.EncryptionKey}).ExportRun(ctx, run.ID)
	if err == nil || !strings.Contains(err.Error(), "already exported") {
		t.Fatalf("ExportRun() err = %v", err)
	}
}
