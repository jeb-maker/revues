package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/handlers"
	"github.com/jeb-maker/revues/internal/web/templates"
)

const oauthTestSecret = "test-secret-at-least-thirty-two-bytes"

type oauthGitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

type oauthGitHubMock struct {
	AccessToken string
	Login       string
	Email       string
	Verified    bool
}

func startOAuthGitHubMock(t *testing.T, mock oauthGitHubMock) (*httptest.Server, *auth.GitHubOAuth) {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/access_token"):
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": mock.AccessToken})

		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/user"):
			if r.Header.Get("Authorization") != "Bearer "+mock.AccessToken {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":         1001,
				"login":      mock.Login,
				"name":       "Test User",
				"avatar_url": "https://example.com/avatar.png",
			})

		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/user/emails"):
			if r.Header.Get("Authorization") != "Bearer "+mock.AccessToken {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]oauthGitHubEmail{
				{Email: mock.Email, Primary: true, Verified: mock.Verified},
			})

		default:
			http.NotFound(w, r)
		}
	}))

	t.Cleanup(srv.Close)

	github := &auth.GitHubOAuth{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		BaseURL:      "http://example.com",
		TokenURL:     srv.URL + "/login/oauth/access_token",
		UserURL:      srv.URL + "/user",
		EmailsURL:    srv.URL + "/user/emails",
	}

	return srv, github
}

func newOAuthAuthHandler(t *testing.T, github *auth.GitHubOAuth) (*handlers.Auth, *store.Store, *auth.SessionManager) {
	t.Helper()

	db := mustMemoryDB(t)
	st := store.New(db)
	sessions := &auth.SessionManager{Store: st, SessionSecret: oauthTestSecret}

	tpl, err := templates.Parse()
	if err != nil {
		t.Fatalf("Parse templates: %v", err)
	}

	handler := &handlers.Auth{
		Templates: tpl,
		Store:     st,
		Sessions:  sessions,
		GitHub:    github,
		Config: config.Config{
			BaseURL:       "http://example.com",
			SessionSecret: oauthTestSecret,
		},
	}

	return handler, st, sessions
}

func oauthCookie(t *testing.T, sessions *auth.SessionManager, state, verifier string) *http.Cookie {
	t.Helper()

	payload, signature := sessions.BuildOAuthCookiePayload(state, verifier)
	return &http.Cookie{
		Name:  "revues_oauth",
		Value: payload + "|" + signature,
	}
}

func sessionCookieFromResponse(t *testing.T, rec *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()

	for _, c := range rec.Result().Cookies() {
		if c.Name == "revues_session" && c.Value != "" {
			return c
		}
	}
	return nil
}

func TestOAuthCallback_SuccessVerifiedWhitelisted(t *testing.T) {
	const (
		state    = "oauth-state-abc"
		verifier = "oauth-verifier-xyz"
		email    = "allowed@example.com"
	)

	_, github := startOAuthGitHubMock(t, oauthGitHubMock{
		AccessToken: "gho_success_token",
		Login:       "allowed-user",
		Email:       email,
		Verified:    true,
	})

	handler, st, sessions := newOAuthAuthHandler(t, github)
	ctx := context.Background()

	if err := st.InsertAllowedEmail(ctx, email, auth.RoleEditor); err != nil {
		t.Fatalf("InsertAllowedEmail(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/github/callback?code=auth-code&state="+state, nil)
	req.AddCookie(oauthCookie(t, sessions, state, verifier))
	rec := httptest.NewRecorder()
	handler.Callback(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if loc := rec.Header().Get("Location"); loc != "/" {
		t.Errorf("Location = %q, want %q", loc, "/")
	}

	sessionCookie := sessionCookieFromResponse(t, rec)
	if sessionCookie == nil {
		t.Fatal("expected revues_session cookie")
	}

	userID, err := st.UserIDByTokenHash(ctx, auth.HashToken(sessionCookie.Value))
	if err != nil {
		t.Fatalf("UserIDByTokenHash(): %v", err)
	}

	user, err := st.UserByID(ctx, userID)
	if err != nil {
		t.Fatalf("UserByID(): %v", err)
	}
	if user.Email != email || user.Login != "allowed-user" || user.Role != auth.RoleEditor {
		t.Errorf("user = %+v, want email=%q role=%q", user, email, auth.RoleEditor)
	}

	for _, c := range rec.Result().Cookies() {
		if c.Name == "revues_oauth" && c.MaxAge >= 0 {
			t.Error("expected revues_oauth cookie to be cleared")
		}
	}
}

func TestOAuthCallback_RefuseUnauthorizedEmail(t *testing.T) {
	const (
		state    = "oauth-state-unauth"
		verifier = "oauth-verifier-unauth"
	)

	_, github := startOAuthGitHubMock(t, oauthGitHubMock{
		AccessToken: "gho_unauth_token",
		Login:       "blocked-user",
		Email:       "blocked@example.com",
		Verified:    true,
	})

	handler, st, sessions := newOAuthAuthHandler(t, github)
	ctx := context.Background()

	if err := st.InsertAllowedEmail(ctx, "other@example.com", auth.RoleReader); err != nil {
		t.Fatalf("InsertAllowedEmail(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/github/callback?code=auth-code&state="+state, nil)
	req.AddCookie(oauthCookie(t, sessions, state, verifier))
	rec := httptest.NewRecorder()
	handler.Callback(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if loc := rec.Header().Get("Location"); !strings.Contains(loc, "error=email+non+autoris") {
		t.Errorf("Location = %q, want unauthorized email error", loc)
	}
	if sessionCookieFromResponse(t, rec) != nil {
		t.Error("expected no session cookie for unauthorized email")
	}
}

func TestOAuthCallback_RefuseUnverifiedEmail(t *testing.T) {
	const (
		state    = "oauth-state-unverified"
		verifier = "oauth-verifier-unverified"
	)

	_, github := startOAuthGitHubMock(t, oauthGitHubMock{
		AccessToken: "gho_unverified_token",
		Login:       "unverified-user",
		Email:       "unverified@example.com",
		Verified:    false,
	})

	handler, st, sessions := newOAuthAuthHandler(t, github)
	ctx := context.Background()

	if err := st.InsertAllowedEmail(ctx, "unverified@example.com", auth.RoleReader); err != nil {
		t.Fatalf("InsertAllowedEmail(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/github/callback?code=auth-code&state="+state, nil)
	req.AddCookie(oauthCookie(t, sessions, state, verifier))
	rec := httptest.NewRecorder()
	handler.Callback(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if loc := rec.Header().Get("Location"); !strings.Contains(loc, "error=profil+github") {
		t.Errorf("Location = %q, want profil github error", loc)
	}
	if sessionCookieFromResponse(t, rec) != nil {
		t.Error("expected no session cookie for unverified email")
	}

	if _, err := st.UserByEmail(ctx, "unverified@example.com"); err == nil {
		t.Error("user should not be created for unverified email")
	} else if !errors.Is(err, store.ErrUserNotFound) {
		t.Fatalf("UserByEmail() error = %v, want ErrUserNotFound", err)
	}
}

func TestOAuthCallback_ErrorInvalidState(t *testing.T) {
	const verifier = "oauth-verifier-bad-state"

	_, github := startOAuthGitHubMock(t, oauthGitHubMock{
		AccessToken: "gho_unused_token",
		Login:       "user",
		Email:       "user@example.com",
		Verified:    true,
	})

	handler, _, sessions := newOAuthAuthHandler(t, github)

	req := httptest.NewRequest(http.MethodGet, "/auth/github/callback?code=auth-code&state=wrong-state", nil)
	req.AddCookie(oauthCookie(t, sessions, "expected-state", verifier))
	rec := httptest.NewRecorder()
	handler.Callback(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if loc := rec.Header().Get("Location"); !strings.Contains(loc, "error=state+invalide") {
		t.Errorf("Location = %q, want state invalide error", loc)
	}
	if sessionCookieFromResponse(t, rec) != nil {
		t.Error("expected no session cookie for invalid state")
	}
}

func TestOAuthCallback_ErrorExpiredOAuthCookie(t *testing.T) {
	tests := []struct {
		name   string
		cookie *http.Cookie
	}{
		{
			name:   "missing cookie",
			cookie: nil,
		},
		{
			name: "invalid signature",
			cookie: &http.Cookie{
				Name:  "revues_oauth",
				Value: "state|verifier|bad-signature",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, github := startOAuthGitHubMock(t, oauthGitHubMock{
				AccessToken: "gho_unused_token",
				Login:       "user",
				Email:       "user@example.com",
				Verified:    true,
			})

			handler, _, _ := newOAuthAuthHandler(t, github)

			req := httptest.NewRequest(
				http.MethodGet,
				"/auth/github/callback?"+url.Values{
					"code":  {"auth-code"},
					"state": {"any-state"},
				}.Encode(),
				nil,
			)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			rec := httptest.NewRecorder()
			handler.Callback(rec, req)

			if rec.Code != http.StatusFound {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
			}
			if loc := rec.Header().Get("Location"); !strings.Contains(loc, "error=session+oauth+invalide") {
				t.Errorf("Location = %q, want session oauth invalide error", loc)
			}
			if sessionCookieFromResponse(t, rec) != nil {
				t.Error("expected no session cookie without valid oauth cookie")
			}
		})
	}
}
