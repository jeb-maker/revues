package jira_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/jeb-maker/revues/internal/crypto"
	"github.com/jeb-maker/revues/internal/integrations/jira"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
)

func testJiraService(t *testing.T) (*jira.Service, *store.Store) {
	t.Helper()

	ctx := context.Background()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("Close(): %v", err)
		}
	})
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("foreign_keys: %v", err)
	}
	if err := store.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate(): %v", err)
	}

	key := make([]byte, crypto.KeySize)
	st := store.New(db)
	ctx = testutil.DefaultOrgContext(ctx, st)
	return &jira.Service{Store: st, EncryptionKey: key}, st
}

func TestServiceSaveLoadCloud(t *testing.T) {
	ctx := context.Background()
	svc, st := testJiraService(t)
	ctx = testutil.DefaultOrgContext(ctx, st)

	cfg := jira.Config{
		InstanceType: jira.InstanceCloud,
		BaseURL:      "https://example.atlassian.net",
		Email:        "user@example.com",
		APIToken:     "secret-token",
	}
	if err := svc.Save(ctx, cfg); err != nil {
		t.Fatalf("Save(): %v", err)
	}

	got, ok, err := svc.Load(ctx)
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	if !ok {
		t.Fatal("expected configured jira")
	}
	if got.InstanceType != cfg.InstanceType || got.BaseURL != cfg.BaseURL || got.Email != cfg.Email || got.APIToken != cfg.APIToken {
		t.Fatalf("Load() = %+v, want %+v", got, cfg)
	}
}

func TestServiceSaveLoadServer(t *testing.T) {
	ctx := context.Background()
	svc, st := testJiraService(t)
	ctx = testutil.DefaultOrgContext(ctx, st)

	cfg := jira.Config{
		InstanceType: jira.InstanceServer,
		BaseURL:      "https://jira.example.com",
		PAT:          "personal-access-token",
	}
	if err := svc.Save(ctx, cfg); err != nil {
		t.Fatalf("Save(): %v", err)
	}

	got, ok, err := svc.Load(ctx)
	if err != nil || !ok {
		t.Fatalf("Load() = ok=%v err=%v", ok, err)
	}
	if got.PAT != cfg.PAT || got.InstanceType != jira.InstanceServer {
		t.Fatalf("Load() = %+v", got)
	}
}

func TestValidateCloudRequiresEmailAndToken(t *testing.T) {
	err := jira.Validate(jira.Config{
		InstanceType: jira.InstanceCloud,
		BaseURL:      "https://example.atlassian.net",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidateServerRequiresPAT(t *testing.T) {
	err := jira.Validate(jira.Config{
		InstanceType: jira.InstanceServer,
		BaseURL:      "https://jira.example.com",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidateBaseURLRequiresHTTPS(t *testing.T) {
	if err := jira.ValidateBaseURL("http://jira.example.com"); err == nil {
		t.Fatal("expected https requirement")
	}
	if err := jira.ValidateBaseURL("http://localhost:8080"); err != nil {
		t.Fatalf("localhost http should be allowed: %v", err)
	}
}

func TestClientTestConnectionCloud(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/3/myself" {
			http.NotFound(w, r)
			return
		}
		auth := r.Header.Get("Authorization")
		if auth != "Basic dXNlckBleGFtcGxlLmNvbTpzZWNyZXQ=" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"accountId":"abc"}`))
	}))
	t.Cleanup(srv.Close)

	client := &jira.Client{HTTPClient: srv.Client()}
	cfg := jira.Config{
		InstanceType: jira.InstanceCloud,
		BaseURL:      srv.URL,
		Email:        "user@example.com",
		APIToken:     "secret",
	}
	if err := client.TestConnection(context.Background(), cfg); err != nil {
		t.Fatalf("TestConnection(): %v", err)
	}
}

func TestClientTestConnectionServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/2/myself" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer server-pat" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"admin"}`))
	}))
	t.Cleanup(srv.Close)

	client := &jira.Client{HTTPClient: srv.Client()}
	cfg := jira.Config{
		InstanceType: jira.InstanceServer,
		BaseURL:      srv.URL,
		PAT:          "server-pat",
	}
	if err := client.TestConnection(context.Background(), cfg); err != nil {
		t.Fatalf("TestConnection(): %v", err)
	}
}

func TestClientTestConnectionUnauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	t.Cleanup(srv.Close)

	client := &jira.Client{HTTPClient: srv.Client()}
	cfg := jira.Config{
		InstanceType: jira.InstanceCloud,
		BaseURL:      srv.URL,
		Email:        "user@example.com",
		APIToken:     "bad",
	}
	if err := client.TestConnection(context.Background(), cfg); err == nil {
		t.Fatal("expected connection error")
	}
}

func TestMergeSecret(t *testing.T) {
	if got := jira.MergeSecret("stored", ""); got != "stored" {
		t.Fatalf("MergeSecret(empty) = %q", got)
	}
	if got := jira.MergeSecret("stored", "new"); got != "new" {
		t.Fatalf("MergeSecret(new) = %q", got)
	}
}

func TestConfigured(t *testing.T) {
	cfg := jira.Config{InstanceType: jira.InstanceCloud, BaseURL: "https://x.atlassian.net", Email: "a@b.c", APIToken: "tok"}
	if !cfg.Configured() {
		t.Fatal("expected configured cloud")
	}
	raw, _ := json.Marshal(cfg)
	if len(raw) == 0 {
		t.Fatal("expected json marshal")
	}
}
