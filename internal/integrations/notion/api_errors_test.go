package notion_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/integrations/notion"
)

func TestCreateReviewPageAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object": "error", "status": 403, "code": "restricted_resource", "message": "Can't access",
		})
	}))
	t.Cleanup(srv.Close)
	client := &notion.Client{HTTPClient: srv.Client(), APIBaseURL: srv.URL + "/v1"}
	_, err := client.CreateReviewPage(context.Background(), notion.Config{APIToken: "tok"}, notion.CreatePageInput{
		DatabaseID: "abc123def4567890abc123def4567890", Title: "Revue",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var api *notion.APIError
	if !errors.As(err, &api) || api.Status != http.StatusForbidden {
		t.Fatalf("err=%v want APIError 403", err)
	}
	if !errors.Is(err, notion.ErrExportFailed) {
		t.Fatalf("err=%v want ErrExportFailed", err)
	}
	msg := notion.UserMessage(err)
	if !strings.Contains(msg, "Partagez la base") {
		t.Fatalf("UserMessage() = %q", msg)
	}
}

func TestGetDatabaseUnauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":"unauthorized","message":"bad token"}`))
	}))
	t.Cleanup(srv.Close)
	client := &notion.Client{HTTPClient: srv.Client(), APIBaseURL: srv.URL + "/v1"}
	_, err := client.GetDatabase(context.Background(), notion.Config{APIToken: "tok"}, "a1b2c3d4e5f6478990abcdef12345678")
	var api *notion.APIError
	if !errors.As(err, &api) || api.Status != http.StatusUnauthorized {
		t.Fatalf("err=%v", err)
	}
}

func TestQueryDatabaseRateLimited(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"code":"rate_limited","message":"slow down"}`))
	}))
	t.Cleanup(srv.Close)
	client := &notion.Client{HTTPClient: srv.Client(), APIBaseURL: srv.URL + "/v1"}
	_, err := client.QueryDatabase(context.Background(), notion.Config{APIToken: "tok"}, "a1b2c3d4e5f6478990abcdef12345678")
	var api *notion.APIError
	if !errors.As(err, &api) || api.Status != http.StatusTooManyRequests {
		t.Fatalf("err=%v", err)
	}
	if got := notion.UserMessage(err); !strings.Contains(got, "limite temporairement") {
		t.Fatalf("UserMessage()=%q", got)
	}
}
