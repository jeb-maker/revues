package notion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	apiBaseURL    = "https://api.notion.com/v1"
	usersMePath   = "/users/me"
	notionVersion = "2022-06-28"
)

var ErrConnectionFailed = errors.New("notion connection failed")

type ConnectionInfo struct {
	UserName      string
	WorkspaceName string
}

type Client struct {
	HTTPClient *http.Client
	APIBaseURL string
}

func (c *Client) apiBaseURL() string {
	if strings.TrimSpace(c.APIBaseURL) != "" {
		return strings.TrimRight(strings.TrimSpace(c.APIBaseURL), "/")
	}
	return apiBaseURL
}

func (c *Client) TestConnection(ctx context.Context, cfg Config) (ConnectionInfo, error) {
	if !cfg.Configured() {
		return ConnectionInfo{}, errors.New("configuration Notion incomplète")
	}
	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBaseURL()+usersMePath, nil)
	if err != nil {
		return ConnectionInfo{}, fmt.Errorf("build notion request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIToken)
	req.Header.Set("Notion-Version", notionVersion)
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return ConnectionInfo{}, fmt.Errorf("%w: %w", ErrConnectionFailed, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return ConnectionInfo{}, fmt.Errorf("%w: status %d %s", ErrConnectionFailed, resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var me struct {
		Name string `json:"name"`
		Bot  *struct {
			WorkspaceName string `json:"workspace_name"`
		} `json:"bot"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&me); err != nil {
		return ConnectionInfo{}, fmt.Errorf("%w: invalid response", ErrConnectionFailed)
	}
	info := ConnectionInfo{UserName: me.Name}
	if me.Bot != nil {
		info.WorkspaceName = me.Bot.WorkspaceName
	}
	return info, nil
}
