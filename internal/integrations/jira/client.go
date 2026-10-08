package jira

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jeb-maker/revues/internal/safehttp"
)

const (
	cloudMyselfPath  = "/rest/api/3/myself"
	serverMyselfPath = "/rest/api/2/myself"
	cloudIssuePath   = "/rest/api/3/issue/"
	serverIssuePath  = "/rest/api/2/issue/"
)

// ErrConnectionFailed is returned when Jira rejects credentials or is unreachable.
var ErrConnectionFailed = errors.New("jira connection failed")

// ErrIssueNotFound is returned when Jira has no issue for the given key.
var ErrIssueNotFound = errors.New("jira issue not found")

// ErrCreateFailed is returned when Jira rejects issue creation.
var ErrCreateFailed = errors.New("jira issue creation failed")

// CreateIssueInput holds fields for a new Jira issue.
type CreateIssueInput struct {
	ProjectKey  string
	IssueType   string
	Summary     string
	Description string
}

// Client tests Jira API connectivity.
type Client struct {
	HTTPClient *http.Client
}

// TestConnection verifies credentials against the Jira REST API.
func (c *Client) TestConnection(ctx context.Context, cfg Config) error {
	if !cfg.Configured() {
		return errors.New("configuration Jira incomplète")
	}

	client := c.httpClient(cfg.APIBaseURL(), 10*time.Second)

	var req *http.Request
	var err error

	switch {
	case cfg.UsesOAuth() || cfg.InstanceType == InstanceCloud:
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, cfg.APIBaseURL()+cloudMyselfPath, nil)
		if err != nil {
			return fmt.Errorf("build jira request: %w", err)
		}
		if setErr := c.setAuth(req, cfg); setErr != nil {
			return setErr
		}
		req.Header.Set("Accept", "application/json")
	case cfg.InstanceType == InstanceServer:
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, cfg.APIBaseURL()+serverMyselfPath, nil)
		if err != nil {
			return fmt.Errorf("build jira request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+cfg.PAT)
		req.Header.Set("Accept", "application/json")
	default:
		return errors.New("type d'instance Jira invalide")
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrConnectionFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("%w: status %d %s", ErrConnectionFailed, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var myself struct {
		AccountID string `json:"accountId"`
		Name      string `json:"name"`
		Key       string `json:"key"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&myself); err != nil {
		return fmt.Errorf("%w: invalid response", ErrConnectionFailed)
	}

	return nil
}

// MyselfEmail fetches the authenticated user's email from /myself (OAuth Cloud).
func (c *Client) MyselfEmail(ctx context.Context, cfg Config) (string, error) {
	if !cfg.Configured() {
		return "", errors.New("configuration Jira incomplète")
	}
	client := c.httpClient(cfg.APIBaseURL(), 10*time.Second)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.APIBaseURL()+cloudMyselfPath, nil)
	if err != nil {
		return "", fmt.Errorf("build jira myself request: %w", err)
	}
	if setErr := c.setAuth(req, cfg); setErr != nil {
		return "", setErr
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrConnectionFailed, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("%w: status %d %s", ErrConnectionFailed, resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var myself struct {
		EmailAddress string `json:"emailAddress"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&myself); err != nil {
		return "", fmt.Errorf("%w: invalid response", ErrConnectionFailed)
	}
	return strings.TrimSpace(myself.EmailAddress), nil
}

// IssueInfo is a minimal Jira issue snapshot (key + workflow status).
type IssueInfo struct {
	Key    string
	Status string
}

// GetIssue verifies that an issue exists in Jira and returns its key.
func (c *Client) GetIssue(ctx context.Context, cfg Config, key string) (string, error) {
	info, err := c.GetIssueInfo(ctx, cfg, key)
	if err != nil {
		return "", err
	}
	return info.Key, nil
}

// GetIssueInfo fetches issue key and current status name from Jira (on demand).
func (c *Client) GetIssueInfo(ctx context.Context, cfg Config, key string) (IssueInfo, error) {
	if !cfg.Configured() {
		return IssueInfo{}, errors.New("configuration Jira incomplète")
	}

	client := c.httpClient(cfg.APIBaseURL(), 10*time.Second)

	issueKey := strings.ToUpper(strings.TrimSpace(key))
	if issueKey == "" {
		return IssueInfo{}, ErrIssueNotFound
	}

	var req *http.Request
	var err error

	switch {
	case cfg.UsesOAuth() || cfg.InstanceType == InstanceCloud:
		u := cfg.APIBaseURL() + cloudIssuePath + url.PathEscape(issueKey) + "?fields=status"
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return IssueInfo{}, fmt.Errorf("build jira issue request: %w", err)
		}
		if setErr := c.setAuth(req, cfg); setErr != nil {
			return IssueInfo{}, setErr
		}
		req.Header.Set("Accept", "application/json")
	case cfg.InstanceType == InstanceServer:
		u := cfg.APIBaseURL() + serverIssuePath + url.PathEscape(issueKey) + "?fields=status"
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return IssueInfo{}, fmt.Errorf("build jira issue request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+cfg.PAT)
		req.Header.Set("Accept", "application/json")
	default:
		return IssueInfo{}, errors.New("type d'instance Jira invalide")
	}

	resp, err := client.Do(req)
	if err != nil {
		return IssueInfo{}, fmt.Errorf("%w: %w", ErrConnectionFailed, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var issue struct {
			Key    string `json:"key"`
			Fields struct {
				Status struct {
					Name string `json:"name"`
				} `json:"status"`
			} `json:"fields"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
			return IssueInfo{}, fmt.Errorf("%w: invalid response", ErrConnectionFailed)
		}
		keyOut := issueKey
		if issue.Key != "" {
			keyOut = strings.ToUpper(issue.Key)
		}
		return IssueInfo{
			Key:    keyOut,
			Status: strings.TrimSpace(issue.Fields.Status.Name),
		}, nil
	case http.StatusNotFound:
		return IssueInfo{}, ErrIssueNotFound
	default:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return IssueInfo{}, fmt.Errorf("%w: status %d %s", ErrConnectionFailed, resp.StatusCode, strings.TrimSpace(string(body)))
	}
}

// CreateIssue creates a Jira issue and returns its key.
func (c *Client) CreateIssue(ctx context.Context, cfg Config, input CreateIssueInput) (string, error) {
	if !cfg.Configured() {
		return "", errors.New("configuration Jira incomplète")
	}

	projectKey := strings.ToUpper(strings.TrimSpace(input.ProjectKey))
	if projectKey == "" {
		return "", errors.New("clé projet Jira requise")
	}
	summary := strings.TrimSpace(input.Summary)
	if summary == "" {
		return "", errors.New("titre Jira requis")
	}
	issueType := strings.TrimSpace(input.IssueType)
	if issueType == "" {
		issueType = DefaultIssueType
	}

	client := c.httpClient(cfg.APIBaseURL(), 15*time.Second)

	var body []byte
	var err error
	useCloudAPI := cfg.UsesOAuth() || cfg.InstanceType == InstanceCloud

	switch {
	case useCloudAPI:
		body, err = json.Marshal(map[string]any{
			"fields": map[string]any{
				"project":     map[string]string{"key": projectKey},
				"summary":     summary,
				"description": cloudDescriptionADF(input.Description),
				"issuetype":   map[string]string{"name": issueType},
			},
		})
	case cfg.InstanceType == InstanceServer:
		body, err = json.Marshal(map[string]any{
			"fields": map[string]any{
				"project":     map[string]string{"key": projectKey},
				"summary":     summary,
				"description": input.Description,
				"issuetype":   map[string]string{"name": issueType},
			},
		})
	default:
		return "", errors.New("type d'instance Jira invalide")
	}
	if err != nil {
		return "", fmt.Errorf("marshal jira create payload: %w", err)
	}

	issuePath := cloudIssuePath
	if cfg.InstanceType == InstanceServer && !cfg.UsesOAuth() {
		issuePath = serverIssuePath
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.APIBaseURL()+issuePath, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build jira create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if setErr := c.setAuth(req, cfg); setErr != nil {
		return "", setErr
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrConnectionFailed, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("%w: status %d %s", ErrCreateFailed, resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var created struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(respBody, &created); err != nil {
		return "", fmt.Errorf("%w: invalid response", ErrCreateFailed)
	}
	if created.Key == "" {
		return "", fmt.Errorf("%w: missing issue key", ErrCreateFailed)
	}
	return strings.ToUpper(created.Key), nil
}

func (c *Client) setAuth(req *http.Request, cfg Config) error {
	if cfg.UsesOAuth() {
		req.Header.Set("Authorization", "Bearer "+cfg.AccessToken)
		return nil
	}
	switch cfg.InstanceType {
	case InstanceCloud:
		token := base64.StdEncoding.EncodeToString([]byte(cfg.Email + ":" + cfg.APIToken))
		req.Header.Set("Authorization", "Basic "+token)
		return nil
	case InstanceServer:
		req.Header.Set("Authorization", "Bearer "+cfg.PAT)
		return nil
	default:
		return errors.New("type d'instance Jira invalide")
	}
}

func cloudDescriptionADF(text string) map[string]any {
	text = strings.TrimSpace(text)
	if text == "" {
		return map[string]any{
			"type":    "doc",
			"version": 1,
			"content": []any{},
		}
	}

	paragraphs := strings.Split(text, "\n")
	content := make([]any, 0, len(paragraphs))
	for _, line := range paragraphs {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		content = append(content, map[string]any{
			"type": "paragraph",
			"content": []any{
				map[string]string{"type": "text", "text": line},
			},
		})
	}
	if len(content) == 0 {
		content = append(content, map[string]any{
			"type": "paragraph",
			"content": []any{
				map[string]string{"type": "text", "text": text},
			},
		})
	}

	return map[string]any{
		"type":    "doc",
		"version": 1,
		"content": content,
	}
}

func (c *Client) httpClient(baseURL string, timeout time.Duration) *http.Client {
	if c != nil && c.HTTPClient != nil {
		return c.HTTPClient
	}
	host := ""
	if u, err := url.Parse(NormalizeBaseURL(baseURL)); err == nil {
		host = u.Hostname()
	}
	return safehttp.NewClient(safehttp.Options{
		Timeout:           timeout,
		DialTimeout:       timeout,
		MaxRedirects:      1,
		AllowDevLocalhost: safehttp.IsLocalhostHost(host),
	})
}
