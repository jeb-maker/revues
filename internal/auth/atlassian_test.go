package auth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
)

func TestAtlassianOAuth_AuthURL(t *testing.T) {
	t.Parallel()
	oauth := &auth.AtlassianOAuth{
		ClientID: "cid",
		BaseURL:  "https://revues.example",
	}
	u := oauth.AuthURL("state-xyz")
	if !strings.Contains(u, "client_id=cid") {
		t.Fatalf("AuthURL missing client_id: %s", u)
	}
	if !strings.Contains(u, "audience=api.atlassian.com") {
		t.Fatalf("AuthURL missing audience: %s", u)
	}
	if !strings.Contains(u, "prompt=consent") {
		t.Fatalf("AuthURL missing prompt: %s", u)
	}
	if !strings.Contains(u, "redirect_uri=") || !strings.Contains(u, "%2Fauth%2Fatlassian%2Fcallback") {
		t.Fatalf("AuthURL missing callback: %s", u)
	}
	if !strings.Contains(u, "state=state-xyz") {
		t.Fatalf("AuthURL missing state: %s", u)
	}
	if strings.Contains(u, "code_challenge") {
		t.Fatal("AuthURL must not use PKCE")
	}
}

func TestAtlassianOAuth_ExchangeAndRefresh(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		grant       string
		wantAccess  string
		wantRefresh string
		wantScope   string
	}{
		{
			name:        "authorization_code",
			grant:       "authorization_code",
			wantAccess:  "access-1",
			wantRefresh: "refresh-1",
			wantScope:   "read:jira-work write:jira-work",
		},
		{
			name:        "refresh_token",
			grant:       "refresh_token",
			wantAccess:  "access-2",
			wantRefresh: "refresh-2",
			wantScope:   "read:jira-work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/oauth/token" {
					http.NotFound(w, r)
					return
				}
				var body map[string]string
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if body["grant_type"] != tt.grant {
					http.Error(w, "bad grant", http.StatusBadRequest)
					return
				}
				if body["client_id"] == "" || body["client_secret"] == "" {
					http.Error(w, "missing client", http.StatusBadRequest)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"access_token":  tt.wantAccess,
					"refresh_token": tt.wantRefresh,
					"expires_in":    3600,
					"scope":         tt.wantScope,
				})
			}))
			t.Cleanup(srv.Close)

			oauth := &auth.AtlassianOAuth{
				ClientID:     "cid",
				ClientSecret: "secret",
				BaseURL:      "https://revues.example",
				TokenURL:     srv.URL + "/oauth/token",
			}

			var got *auth.AtlassianTokenResponse
			var err error
			switch tt.grant {
			case "authorization_code":
				got, err = oauth.ExchangeCode(context.Background(), "auth-code")
			case "refresh_token":
				got, err = oauth.RefreshAccessToken(context.Background(), "old-refresh")
			}
			if err != nil {
				t.Fatalf("token: %v", err)
			}
			if got.AccessToken != tt.wantAccess || got.RefreshToken != tt.wantRefresh || got.Scope != tt.wantScope {
				t.Fatalf("got = %+v", got)
			}
			if got.ExpiresIn != 3600 {
				t.Fatalf("ExpiresIn = %d", got.ExpiresIn)
			}
		})
	}
}

func TestAtlassianOAuth_AccessibleResources(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"id": "c1", "url": "https://a.atlassian.net", "name": "A", "scopes": []string{"read:jira-work"}},
			{"id": "c2", "url": "https://b.atlassian.net", "name": "B", "scopes": []string{"write:jira-work"}},
		})
	}))
	t.Cleanup(srv.Close)

	oauth := &auth.AtlassianOAuth{AccessibleResourcesURL: srv.URL}
	got, err := oauth.AccessibleResources(context.Background(), "tok")
	if err != nil {
		t.Fatalf("AccessibleResources(): %v", err)
	}
	if len(got) != 2 || got[1].ID != "c2" {
		t.Fatalf("got = %+v", got)
	}
}

func TestPickCloudResource(t *testing.T) {
	t.Parallel()
	resources := []auth.AtlassianResource{
		{ID: "1", URL: "https://a.atlassian.net", Scopes: []string{"read:jira-work"}},
		{ID: "2", URL: "https://b.atlassian.net", Scopes: []string{"write:jira-work"}},
		{ID: "3", URL: "https://c.atlassian.net", Scopes: []string{"read:jira-work", "write:jira-work"}},
	}

	got, err := auth.PickCloudResource(resources, "https://c.atlassian.net/")
	if err != nil || got.ID != "3" {
		t.Fatalf("prefer match = %+v, %v", got, err)
	}

	got, err = auth.PickCloudResource(resources, "https://unknown.atlassian.net")
	if err != nil || got.ID != "2" {
		t.Fatalf("first write = %+v, %v", got, err)
	}

	_, err = auth.PickCloudResource(nil, "")
	if err == nil {
		t.Fatal("expected error for empty resources")
	}
}
