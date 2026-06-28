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

type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

type githubMockConfig struct {
	AccessToken string
	User        map[string]any
	Emails      []githubEmail
}

func startGitHubMock(t *testing.T, cfg githubMockConfig) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/access_token"):
			if err := r.ParseForm(); err != nil {
				http.Error(w, "bad form", http.StatusBadRequest)
				return
			}
			if r.Form.Get("client_id") == "" || r.Form.Get("code") == "" {
				http.Error(w, "missing fields", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": cfg.AccessToken})

		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/user"):
			if r.Header.Get("Authorization") != "Bearer "+cfg.AccessToken {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(cfg.User)

		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/user/emails"):
			if r.Header.Get("Authorization") != "Bearer "+cfg.AccessToken {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(cfg.Emails)

		default:
			http.NotFound(w, r)
		}
	}))

	t.Cleanup(srv.Close)
	return srv
}

func TestGitHubOAuth_ExchangeCode(t *testing.T) {
	t.Parallel()

	const token = "gho_mock_access_token"
	srv := startGitHubMock(t, githubMockConfig{AccessToken: token})

	oauth := &auth.GitHubOAuth{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		BaseURL:      "http://example.com",
		TokenURL:     srv.URL + "/login/oauth/access_token",
	}

	got, err := oauth.ExchangeCode(context.Background(), "auth-code", "pkce-verifier")
	if err != nil {
		t.Fatalf("ExchangeCode() error = %v", err)
	}
	if got != token {
		t.Errorf("token = %q, want %q", got, token)
	}
}

func TestGitHubOAuth_FetchProfile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		emails  []githubEmail
		wantErr string
		want    *auth.GitHubProfile
	}{
		{
			name: "primary verified email",
			emails: []githubEmail{
				{Email: "user@example.com", Primary: true, Verified: true},
			},
			want: &auth.GitHubProfile{
				ID:          42,
				Login:       "octocat",
				DisplayName: "Octo Cat",
				AvatarURL:   "https://example.com/avatar.png",
				Email:       "user@example.com",
			},
		},
		{
			name: "fallback verified non-primary email",
			emails: []githubEmail{
				{Email: "primary@example.com", Primary: true, Verified: false},
				{Email: "verified@example.com", Primary: false, Verified: true},
			},
			want: &auth.GitHubProfile{
				ID:          42,
				Login:       "octocat",
				DisplayName: "Octo Cat",
				AvatarURL:   "https://example.com/avatar.png",
				Email:       "verified@example.com",
			},
		},
		{
			name: "no verified email",
			emails: []githubEmail{
				{Email: "unverified@example.com", Primary: true, Verified: false},
			},
			wantErr: "no verified email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			const token = "gho_mock_access_token"
			srv := startGitHubMock(t, githubMockConfig{
				AccessToken: token,
				User: map[string]any{
					"id":         42,
					"login":      "octocat",
					"name":       "Octo Cat",
					"avatar_url": "https://example.com/avatar.png",
				},
				Emails: tt.emails,
			})

			oauth := &auth.GitHubOAuth{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				BaseURL:      "http://example.com",
				TokenURL:     srv.URL + "/login/oauth/access_token",
				UserURL:      srv.URL + "/user",
				EmailsURL:    srv.URL + "/user/emails",
			}

			gotToken, err := oauth.ExchangeCode(context.Background(), "auth-code", "pkce-verifier")
			if err != nil {
				t.Fatalf("ExchangeCode() error = %v", err)
			}
			if gotToken != token {
				t.Fatalf("token = %q, want %q", gotToken, token)
			}

			profile, err := oauth.FetchProfile(context.Background(), gotToken)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("FetchProfile() error = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("FetchProfile() error = %v", err)
			}
			if profile.ID != tt.want.ID || profile.Login != tt.want.Login ||
				profile.DisplayName != tt.want.DisplayName || profile.AvatarURL != tt.want.AvatarURL ||
				profile.Email != tt.want.Email {
				t.Errorf("FetchProfile() = %+v, want %+v", profile, tt.want)
			}
		})
	}
}
