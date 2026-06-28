package notion_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/jeb-maker/revues/internal/crypto"
	"github.com/jeb-maker/revues/internal/integrations/notion"
	"github.com/jeb-maker/revues/internal/store"
)

func testNotionService(t *testing.T) *notion.Service {
	t.Helper()
	ctx := context.Background()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	_, _ = db.ExecContext(ctx, "PRAGMA foreign_keys = ON")
	if err := store.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate(): %v", err)
	}
	return &notion.Service{Store: store.New(db), EncryptionKey: make([]byte, crypto.KeySize)}
}

func TestServiceSaveLoad(t *testing.T) {
	ctx := context.Background()
	svc := testNotionService(t)
	cfg := notion.Config{APIToken: "secret-token", WorkspaceName: "Acme"}
	if err := svc.Save(ctx, cfg); err != nil {
		t.Fatalf("Save(): %v", err)
	}
	got, ok, err := svc.Load(ctx)
	if err != nil || !ok || got.APIToken != cfg.APIToken {
		t.Fatalf("Load() = %+v ok=%v err=%v", got, ok, err)
	}
}

func TestValidateRequiresToken(t *testing.T) {
	if err := notion.Validate(notion.Config{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestClientTestConnection(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/users/me" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"Bot","bot":{"workspace_name":"Acme"}}`))
	}))
	t.Cleanup(srv.Close)
	client := &notion.Client{HTTPClient: srv.Client(), APIBaseURL: srv.URL + "/v1"}
	info, err := client.TestConnection(context.Background(), notion.Config{APIToken: "tok"})
	if err != nil || info.WorkspaceName != "Acme" {
		t.Fatalf("TestConnection() = %+v err=%v", info, err)
	}
}
