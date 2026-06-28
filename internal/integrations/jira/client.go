package jira

import (
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
