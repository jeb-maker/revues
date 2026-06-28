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
)

// ErrConnectionFailed is returned when Jira rejects credentials or is unreachable.
var ErrConnectionFailed = errors.New("jira connection failed")

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
