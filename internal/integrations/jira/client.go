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
	"strings"
	"time"
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

	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	baseURL := NormalizeBaseURL(cfg.BaseURL)
	var req *http.Request
	var err error

	switch cfg.InstanceType {
	case InstanceCloud:
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, baseURL+cloudMyselfPath, nil)
		if err != nil {
			return fmt.Errorf("build jira request: %w", err)
		}
		token := base64.StdEncoding.EncodeToString([]byte(cfg.Email + ":" + cfg.APIToken))
		req.Header.Set("Authorization", "Basic "+token)
		req.Header.Set("Accept", "application/json")
	case InstanceServer:
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, baseURL+serverMyselfPath, nil)
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

// GetIssue verifies that an issue exists in Jira and returns its key.
func (c *Client) GetIssue(ctx context.Context, cfg Config, key string) (string, error) {
	if !cfg.Configured() {
		return "", errors.New("configuration Jira incomplète")
	}

	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	baseURL := NormalizeBaseURL(cfg.BaseURL)
	issueKey := strings.ToUpper(strings.TrimSpace(key))
	if issueKey == "" {
		return "", ErrIssueNotFound
	}

	var req *http.Request
	var err error

	switch cfg.InstanceType {
	case InstanceCloud:
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, baseURL+cloudIssuePath+issueKey, nil)
		if err != nil {
			return "", fmt.Errorf("build jira issue request: %w", err)
		}
		token := base64.StdEncoding.EncodeToString([]byte(cfg.Email + ":" + cfg.APIToken))
		req.Header.Set("Authorization", "Basic "+token)
		req.Header.Set("Accept", "application/json")
	case InstanceServer:
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, baseURL+serverIssuePath+issueKey, nil)
		if err != nil {
			return "", fmt.Errorf("build jira issue request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+cfg.PAT)
		req.Header.Set("Accept", "application/json")
	default:
		return "", errors.New("type d'instance Jira invalide")
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrConnectionFailed, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var issue struct {
			Key string `json:"key"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
			return "", fmt.Errorf("%w: invalid response", ErrConnectionFailed)
		}
		if issue.Key == "" {
			return issueKey, nil
		}
		return strings.ToUpper(issue.Key), nil
	case http.StatusNotFound:
		return "", ErrIssueNotFound
	default:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("%w: status %d %s", ErrConnectionFailed, resp.StatusCode, strings.TrimSpace(string(body)))
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

	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	baseURL := NormalizeBaseURL(cfg.BaseURL)
	var body []byte
	var err error

	switch cfg.InstanceType {
	case InstanceCloud:
		body, err = json.Marshal(map[string]any{
			"fields": map[string]any{
				"project":     map[string]string{"key": projectKey},
				"summary":     summary,
				"description": cloudDescriptionADF(input.Description),
				"issuetype":   map[string]string{"name": issueType},
			},
		})
	case InstanceServer:
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

	var issuePath string
	switch cfg.InstanceType {
	case InstanceCloud:
		issuePath = cloudIssuePath
	case InstanceServer:
		issuePath = serverIssuePath
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+issuePath, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build jira create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	switch cfg.InstanceType {
	case InstanceCloud:
		token := base64.StdEncoding.EncodeToString([]byte(cfg.Email + ":" + cfg.APIToken))
		req.Header.Set("Authorization", "Basic "+token)
	case InstanceServer:
		req.Header.Set("Authorization", "Bearer "+cfg.PAT)
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
