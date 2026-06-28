package jira_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jeb-maker/revues/internal/integrations/jira"
)

func TestParseIssueReference(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"key", "rev-42", "REV-42", false},
		{"browse url", "https://example.atlassian.net/browse/PROJ-7", "PROJ-7", false},
		{"invalid", "not-a-key", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := jira.ParseIssueReference(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseIssueReference() = %v", err)
			}
			if got != tt.want {
				t.Fatalf("key = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBrowseURL(t *testing.T) {
	t.Parallel()
	got := jira.BrowseURL("https://example.atlassian.net/", "rev-1")
	want := "https://example.atlassian.net/browse/REV-1"
	if got != want {
		t.Fatalf("BrowseURL() = %q, want %q", got, want)
	}
}

func TestValidateBrowseURLHost(t *testing.T) {
	t.Parallel()
	cfg := jira.Config{BaseURL: "https://example.atlassian.net"}
	if err := jira.ValidateBrowseURL(cfg, "https://other.example.net/browse/REV-1"); err == nil {
		t.Fatal("expected host mismatch error")
	}
	if err := jira.ValidateBrowseURL(cfg, "https://example.atlassian.net/browse/REV-1"); err != nil {
		t.Fatalf("ValidateBrowseURL() = %v", err)
	}
}

func TestClientGetIssueCloud(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/3/issue/REV-9" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Basic dXNlckBleGFtcGxlLmNvbTpzZWNyZXQ=" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"key":"REV-9"}`))
	}))
	t.Cleanup(srv.Close)

	client := &jira.Client{HTTPClient: srv.Client()}
	cfg := jira.Config{
		InstanceType: jira.InstanceCloud,
		BaseURL:      srv.URL,
		Email:        "user@example.com",
		APIToken:     "secret",
	}
	key, err := client.GetIssue(context.Background(), cfg, "REV-9")
	if err != nil {
		t.Fatalf("GetIssue(): %v", err)
	}
	if key != "REV-9" {
		t.Fatalf("key = %q", key)
	}
}

func TestClientGetIssueServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/2/issue/SRV-1" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer server-pat" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"key":"SRV-1"}`))
	}))
	t.Cleanup(srv.Close)

	client := &jira.Client{HTTPClient: srv.Client()}
	cfg := jira.Config{
		InstanceType: jira.InstanceServer,
		BaseURL:      srv.URL,
		PAT:          "server-pat",
	}
	key, err := client.GetIssue(context.Background(), cfg, "SRV-1")
	if err != nil {
		t.Fatalf("GetIssue(): %v", err)
	}
	if key != "SRV-1" {
		t.Fatalf("key = %q", key)
	}
}

func TestClientGetIssueNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	client := &jira.Client{HTTPClient: srv.Client()}
	cfg := jira.Config{
		InstanceType: jira.InstanceCloud,
		BaseURL:      srv.URL,
		Email:        "user@example.com",
		APIToken:     "secret",
	}
	_, err := client.GetIssue(context.Background(), cfg, "MISSING-1")
	if err == nil {
		t.Fatal("expected not found")
	}
}
