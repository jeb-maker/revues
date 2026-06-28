package notion_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jeb-maker/revues/internal/integrations/notion"
)

func TestParseDatabaseRef(t *testing.T) {
	const id = "a1b2c3d4e5f6478990abcdef12345678"
	got, err := notion.ParseDatabaseRef(id)
	if err != nil || got != id {
		t.Fatalf("got=%q err=%v", got, err)
	}
}

func TestBuildImportPreview(t *testing.T) {
	db := notion.DatabaseInfo{Properties: []notion.PropertyInfo{{Name: "Name", Type: "title"}}}
	pages := []notion.DatabasePage{{Properties: map[string]json.RawMessage{
		"Name": json.RawMessage(`{"title":[{"plain_text":"P"}]}`),
	}}}
	preview, err := notion.BuildImportPreview(db, pages, notion.ColumnMapping{Label: "Name"}, "")
	if err != nil || len(preview.Items) != 1 {
		t.Fatalf("preview=%+v err=%v", preview, err)
	}
}

func TestGetDatabaseAndQuery(t *testing.T) {
	const dbID = "a1b2c3d4e5f6478990abcdef12345678"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(map[string]any{"id": dbID, "title": []map[string]string{{"plain_text": "C"}}, "properties": map[string]any{"Name": map[string]string{"type": "title"}}})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"results": []map[string]any{{"properties": map[string]any{"Name": map[string]any{"title": []map[string]string{{"plain_text": "P"}}}}}}, "has_more": false})
	}))
	t.Cleanup(srv.Close)
	client := &notion.Client{HTTPClient: srv.Client(), APIBaseURL: srv.URL + "/v1"}
	cfg := notion.Config{APIToken: "secret"}
	if _, err := client.GetDatabase(context.Background(), cfg, dbID); err != nil {
		t.Fatal(err)
	}
}
