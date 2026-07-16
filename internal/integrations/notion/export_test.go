package notion_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	runs "github.com/jeb-maker/revues/internal/features/runs"
	"github.com/jeb-maker/revues/internal/integrations/notion"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
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
	ctx = testutil.DefaultOrgContext(ctx, svc.Store.(*store.Store))
	_ = svc.Save(ctx, notion.Config{APIToken: "tok", DefaultDatabaseID: "abc123def4567890abc123def4567890"})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"page-id","url":"https://notion.so/revue-export"}`))
	}))
	t.Cleanup(srv.Close)
	st := svc.Store
	storeSt := st.(*store.Store)
	lead, _ := storeSt.UpsertGitHubUser(ctx, 1, "lead", "lead@example.com", "Lead", "", "editor")
	project, _ := storeSt.CreateProject(ctx, "Alpha", "", lead.ID, nil)
	template, _, _ := storeSt.CreateChecklistTemplate(ctx, "Modèle", lead.ID, nil, []store.TemplateItemInput{{Label: "Backup", Required: true}})
	run, _ := storeSt.CreateChecklistRun(ctx, project.ID, template.ID, lead.ID)
	_ = storeSt.StartRun(ctx, run.ID)
	runItems, _ := storeSt.ListRunItems(ctx, run.ID)
	_ = storeSt.UpdateRunItemStatus(ctx, run.ID, runItems[0].ID, lead.ID, runs.StatusOK, "")
	_ = storeSt.CompleteRun(ctx, run.ID, "Clôturée")
	exportSvc := &notion.ExportService{Store: storeSt, EncryptionKey: svc.EncryptionKey, Client: &notion.Client{HTTPClient: srv.Client(), APIBaseURL: srv.URL + "/v1"}, BaseURL: "http://example.com"}
	url, err := exportSvc.ExportRun(ctx, run.ID)
	if err != nil || url != "https://notion.so/revue-export" {
		t.Fatalf("ExportRun() = %q err=%v", url, err)
	}
}

func TestExportServiceAlreadyExported(t *testing.T) {
	ctx := context.Background()
	svc := testNotionService(t)
	ctx = testutil.DefaultOrgContext(ctx, svc.Store.(*store.Store))
	_ = svc.Save(ctx, notion.Config{APIToken: "tok", DefaultDatabaseID: "abc123def4567890abc123def4567890"})
	st := svc.Store
	storeSt := st.(*store.Store)
	lead, _ := storeSt.UpsertGitHubUser(ctx, 2, "lead", "lead@example.com", "Lead", "", "editor")
	project, _ := storeSt.CreateProject(ctx, "Alpha", "", lead.ID, nil)
	template, _, _ := storeSt.CreateChecklistTemplate(ctx, "Modèle", lead.ID, nil, []store.TemplateItemInput{{Label: "P", Required: true}})
	run, _ := storeSt.CreateChecklistRun(ctx, project.ID, template.ID, lead.ID)
	_ = storeSt.StartRun(ctx, run.ID)
	_ = storeSt.CompleteRun(ctx, run.ID, "Note")
	_ = storeSt.SetRunNotionURL(ctx, run.ID, "https://notion.so/existing")
	_, err := (&notion.ExportService{Store: storeSt, EncryptionKey: svc.EncryptionKey}).ExportRun(ctx, run.ID)
	if err == nil || !strings.Contains(err.Error(), "already exported") {
		t.Fatalf("ExportRun() err = %v", err)
	}
}
