package confluence_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/jeb-maker/revues/internal/crypto"
	"github.com/jeb-maker/revues/internal/integrations/confluence"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
)

func testConfluenceService(t *testing.T) (*confluence.Service, *store.Store) {
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
	return &confluence.Service{Store: st, EncryptionKey: key}, st
}

func TestServiceSaveLoadClear(t *testing.T) {
	ctx := context.Background()
	svc, st := testConfluenceService(t)
	ctx = testutil.DefaultOrgContext(ctx, st)

	cfg := confluence.Config{
		BaseURL:      "https://example.atlassian.net",
		Email:        "user@example.com",
		APIToken:     "secret-token",
		SpaceKey:     "rev",
		ParentPageID: "12345",
	}
	if err := svc.Save(ctx, cfg); err != nil {
		t.Fatalf("Save(): %v", err)
	}

	got, ok, err := svc.Load(ctx)
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	if !ok {
		t.Fatal("expected configured confluence")
	}
	if got.BaseURL != cfg.BaseURL || got.Email != cfg.Email || got.APIToken != cfg.APIToken {
		t.Fatalf("Load() = %+v", got)
	}
	if got.SpaceKey != "REV" {
		t.Fatalf("SpaceKey = %q want REV", got.SpaceKey)
	}
	if got.ParentPageID != "12345" {
		t.Fatalf("ParentPageID = %q", got.ParentPageID)
	}

	if clearErr := svc.Clear(ctx); clearErr != nil {
		t.Fatalf("Clear(): %v", clearErr)
	}
	_, ok, err = svc.Load(ctx)
	if err != nil {
		t.Fatalf("Load after Clear(): %v", err)
	}
	if ok {
		t.Fatal("expected no confluence config after Clear")
	}
}

func TestValidateBaseURLRejectsPrivateIP(t *testing.T) {
	t.Parallel()
	err := confluence.ValidateBaseURL("https://10.0.0.8")
	if err == nil || !strings.Contains(err.Error(), "non autorisée") {
		t.Fatalf("ValidateBaseURL() = %v", err)
	}
}

func TestClientTestConnectionAndCreatePage(t *testing.T) {
	t.Parallel()

	var sawAuth bool
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Basic ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		sawAuth = true
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/wiki/rest/api/user/current":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"accountId":"abc","type":"known"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/wiki/rest/api/content":
			body, _ := io.ReadAll(r.Body)
			var payload map[string]any
			_ = json.Unmarshal(body, &payload)
			if payload["type"] != "page" {
				http.Error(w, "bad type", http.StatusBadRequest)
				return
			}
			space, _ := payload["space"].(map[string]any)
			if space["key"] != "REV" {
				http.Error(w, "bad space", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"id":"999",
				"_links":{"base":"` + r.Host + `","webui":"/wiki/spaces/REV/pages/999/Title"}
			}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(mock.Close)

	cfg := confluence.Config{
		BaseURL:  mock.URL,
		Email:    "bot@example.com",
		APIToken: "tok",
		SpaceKey: "REV",
	}
	client := &confluence.Client{HTTPClient: mock.Client()}

	if err := client.TestConnection(context.Background(), cfg); err != nil {
		t.Fatalf("TestConnection(): %v", err)
	}
	if !sawAuth {
		t.Fatal("expected auth header")
	}

	created, err := client.CreatePage(context.Background(), cfg, confluence.CreatePageInput{
		Title:    "Revue test",
		SpaceKey: "REV",
		BodyHTML: "<p>hello</p>",
	})
	if err != nil {
		t.Fatalf("CreatePage(): %v", err)
	}
	if created.ID != "999" {
		t.Fatalf("ID = %q", created.ID)
	}
	if !strings.Contains(created.URL, "/wiki/spaces/REV/pages/999") {
		t.Fatalf("URL = %q", created.URL)
	}
}

func TestBuildPageHTMLEscapesAndIncludesJira(t *testing.T) {
	t.Parallel()

	html := confluence.BuildPageHTML(confluence.PageContent{
		Title:            `Revue <script>`,
		SubjectName:      "Sujet & Co",
		CompletedByLogin: "alice",
		CompletedAt:      "2026-01-02T15:04:05Z",
		ClosingNote:      "Note <b>",
		Items: []store.RunItem{
			{ID: 1, Section: "A", Label: "Point ok", Status: store.RunItemStatusOK, Comment: "fine"},
			{ID: 2, Section: "A", Label: "Point nok", Status: store.RunItemStatusNOK, Comment: "bad & ugly"},
		},
		JiraByItemID: map[int64]store.IntegrationLink{
			2: {ExternalKey: "REV-1", ExternalURL: "https://example.atlassian.net/browse/REV-1"},
		},
	})
	if strings.Contains(html, "<script>") {
		t.Fatal("script tag not escaped")
	}
	if !strings.Contains(html, "Sujet &amp; Co") {
		t.Fatalf("subject not escaped: %s", html)
	}
	if !strings.Contains(html, `href="https://example.atlassian.net/browse/REV-1"`) {
		t.Fatalf("missing jira link: %s", html)
	}
	if !strings.Contains(html, "Non validé") {
		t.Fatalf("missing nok label: %s", html)
	}
}
