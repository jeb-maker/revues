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
	atlassianAuthorizeURL           = "https://auth.atlassian.com/authorize"
	atlassianTokenURL               = "https://auth.atlassian.com/oauth/token"
	atlassianAccessibleResourcesURL = "https://api.atlassian.com/oauth/token/accessible-resources"
	atlassianOAuthScopes            = "read:jira-work write:jira-work read:jira-user offline_access"
	atlassianAudience               = "api.atlassian.com"
)

// AtlassianTokenResponse is returned by the Atlassian token endpoint.
type AtlassianTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// AtlassianResource is one site from accessible-resources.
type AtlassianResource struct {
	ID     string   `json:"id"`
	URL    string   `json:"url"`
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

// AtlassianOAuth performs Authorization Code (3LO) against Atlassian.
// Atlassian does not require PKCE; state + client_secret are used.
type AtlassianOAuth struct {
	ClientID     string
	ClientSecret string
	BaseURL      string
	HTTPClient   *http.Client
	// Optional overrides for tests (empty = production Atlassian URLs).
	AuthorizeURL           string
	TokenURL               string
	AccessibleResourcesURL string
}

// AuthURL builds the Atlassian authorization redirect URL.
func (a *AtlassianOAuth) AuthURL(state string) string {
	q := url.Values{}
	q.Set("audience", atlassianAudience)
	q.Set("client_id", a.ClientID)
	q.Set("scope", atlassianOAuthScopes)
	q.Set("redirect_uri", a.callbackURL())
	q.Set("state", state)
	q.Set("response_type", "code")
	q.Set("prompt", "consent")

	return a.authorizeURL() + "?" + q.Encode()
}

func (a *AtlassianOAuth) callbackURL() string {
	return strings.TrimRight(a.BaseURL, "/") + "/auth/atlassian/callback"
}

// ExchangeCode trades an authorization code for access + refresh tokens.
func (a *AtlassianOAuth) ExchangeCode(ctx context.Context, code string) (*AtlassianTokenResponse, error) {
	body := map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     a.ClientID,
		"client_secret": a.ClientSecret,
		"code":          code,
		"redirect_uri":  a.callbackURL(),
	}
	return a.postToken(ctx, body)
}

// RefreshAccessToken exchanges a refresh token for a new access token.
func (a *AtlassianOAuth) RefreshAccessToken(ctx context.Context, refreshToken string) (*AtlassianTokenResponse, error) {
	body := map[string]string{
		"grant_type":    "refresh_token",
		"client_id":     a.ClientID,
		"client_secret": a.ClientSecret,
		"refresh_token": refreshToken,
	}
	return a.postToken(ctx, body)
}

func (a *AtlassianOAuth) postToken(ctx context.Context, body map[string]string) (*AtlassianTokenResponse, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal token request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL(), bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("token request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client().Do(req)
	if err != nil {
		return nil, fmt.Errorf("token exchange: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange status %d: %s", resp.StatusCode, bytes.TrimSpace(respBody))
	}

	var token AtlassianTokenResponse
	if err := json.Unmarshal(respBody, &token); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}
	if token.AccessToken == "" {
		return nil, fmt.Errorf("empty access token")
	}
	return &token, nil
}

// AccessibleResources lists Atlassian Cloud sites the token can access.
func (a *AtlassianOAuth) AccessibleResources(ctx context.Context, accessToken string) ([]AtlassianResource, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.accessibleResourcesURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("accessible-resources request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client().Do(req)
	if err != nil {
		return nil, fmt.Errorf("accessible-resources: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("accessible-resources status %d: %s", resp.StatusCode, bytes.TrimSpace(body))
	}

	var resources []AtlassianResource
	if err := json.NewDecoder(resp.Body).Decode(&resources); err != nil {
		return nil, fmt.Errorf("decode accessible-resources: %w", err)
	}
	return resources, nil
}

// PickCloudResource prefers a site matching preferBaseURL, else the first with write:jira-work.
func PickCloudResource(resources []AtlassianResource, preferBaseURL string) (AtlassianResource, error) {
	prefer := strings.TrimRight(strings.TrimSpace(strings.ToLower(preferBaseURL)), "/")
	if prefer != "" {
		for _, r := range resources {
			if strings.TrimRight(strings.ToLower(r.URL), "/") == prefer {
				return r, nil
			}
		}
	}
	for _, r := range resources {
		if resourceHasScope(r, "write:jira-work") {
			return r, nil
		}
	}
	if len(resources) > 0 {
		return resources[0], nil
	}
	return AtlassianResource{}, fmt.Errorf("aucun site Atlassian accessible")
}

func resourceHasScope(r AtlassianResource, want string) bool {
	for _, s := range r.Scopes {
		if s == want {
			return true
		}
	}
	return false
}

func (a *AtlassianOAuth) client() *http.Client {
	if a.HTTPClient != nil {
		return a.HTTPClient
	}
	return &http.Client{Timeout: 15 * time.Second}
}

func (a *AtlassianOAuth) authorizeURL() string {
	if a.AuthorizeURL != "" {
		return a.AuthorizeURL
	}
	return atlassianAuthorizeURL
}

func (a *AtlassianOAuth) tokenURL() string {
	if a.TokenURL != "" {
		return a.TokenURL
	}
	return atlassianTokenURL
}

func (a *AtlassianOAuth) accessibleResourcesURL() string {
	if a.AccessibleResourcesURL != "" {
		return a.AccessibleResourcesURL
	}
	return atlassianAccessibleResourcesURL
}
