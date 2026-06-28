package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	githubAuthorizeURL = "https://github.com/login/oauth/authorize"
	githubTokenURL     = "https://github.com/login/oauth/access_token"
	githubUserURL      = "https://api.github.com/user"
	githubEmailsURL    = "https://api.github.com/user/emails"
)

// GitHubProfile holds identity data fetched from GitHub.
type GitHubProfile struct {
	ID          int64
	Login       string
	DisplayName string
	AvatarURL   string
	Email       string
}

// GitHubOAuth performs Authorization Code + PKCE against GitHub.
type GitHubOAuth struct {
	ClientID     string
	ClientSecret string
	BaseURL      string
	HTTPClient   *http.Client
	// Optional overrides for tests (empty = production GitHub URLs).
	TokenURL  string
	UserURL   string
	EmailsURL string
}

// AuthURL builds the GitHub authorization redirect URL.
func (g *GitHubOAuth) AuthURL(state, codeChallenge string) string {
	q := url.Values{}
	q.Set("client_id", g.ClientID)
	q.Set("redirect_uri", g.callbackURL())
	q.Set("state", state)
	q.Set("scope", "read:user user:email")
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")

	return githubAuthorizeURL + "?" + q.Encode()
}

func (g *GitHubOAuth) callbackURL() string {
	return strings.TrimRight(g.BaseURL, "/") + "/auth/github/callback"
}

// ExchangeCode trades authorization code + verifier for an access token.
func (g *GitHubOAuth) ExchangeCode(ctx context.Context, code, verifier string) (string, error) {
	body := url.Values{}
	body.Set("client_id", g.ClientID)
	body.Set("client_secret", g.ClientSecret)
	body.Set("code", code)
	body.Set("redirect_uri", g.callbackURL())
	body.Set("code_verifier", verifier)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.tokenURL(), strings.NewReader(body.Encode()))
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := g.client().Do(req)
	if err != nil {
		return "", fmt.Errorf("token exchange: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token exchange status %d", resp.StatusCode)
	}

	var payload struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}
	if payload.Error != "" {
		return "", fmt.Errorf("token error: %s", payload.Error)
	}
	if payload.AccessToken == "" {
		return "", fmt.Errorf("empty access token")
	}

	return payload.AccessToken, nil
}

// FetchProfile loads GitHub user profile and primary verified email.
func (g *GitHubOAuth) FetchProfile(ctx context.Context, accessToken string) (*GitHubProfile, error) {
	user, err := g.fetchUser(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	email, err := g.fetchPrimaryVerifiedEmail(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	return &GitHubProfile{
		ID:          user.ID,
		Login:       user.Login,
		DisplayName: user.Name,
		AvatarURL:   user.AvatarURL,
		Email:       email,
	}, nil
}

type githubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

func (g *GitHubOAuth) fetchUser(ctx context.Context, accessToken string) (*githubUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.userURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("user request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := g.client().Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("fetch user status %d: %s", resp.StatusCode, bytes.TrimSpace(body))
	}

	var user githubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("decode user: %w", err)
	}

	return &user, nil
}

func (g *GitHubOAuth) fetchPrimaryVerifiedEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.emailsURL(), nil)
	if err != nil {
		return "", fmt.Errorf("emails request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := g.client().Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch emails: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch emails status %d", resp.StatusCode)
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", fmt.Errorf("decode emails: %w", err)
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	for _, e := range emails {
		if e.Verified {
			return e.Email, nil
		}
	}

	return "", fmt.Errorf("no verified email on GitHub account")
}

func (g *GitHubOAuth) client() *http.Client {
	if g.HTTPClient != nil {
		return g.HTTPClient
	}
	return &http.Client{Timeout: 10 * time.Second}
}

func (g *GitHubOAuth) tokenURL() string {
	if g.TokenURL != "" {
		return g.TokenURL
	}
	return githubTokenURL
}

func (g *GitHubOAuth) userURL() string {
	if g.UserURL != "" {
		return g.UserURL
	}
	return githubUserURL
}

func (g *GitHubOAuth) emailsURL() string {
	if g.EmailsURL != "" {
		return g.EmailsURL
	}
	return githubEmailsURL
}
