package confluence

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
	currentUserPath = "/wiki/rest/api/user/current"
	contentPath     = "/wiki/rest/api/content"
)

// ErrConnectionFailed is returned when Confluence rejects credentials or is unreachable.
var ErrConnectionFailed = errors.New("confluence connection failed")

// ErrCreateFailed is returned when Confluence rejects page creation.
var ErrCreateFailed = errors.New("confluence page creation failed")

// CreatePageInput holds fields for a new Confluence page.
type CreatePageInput struct {
	Title        string
	SpaceKey     string
	ParentPageID string
	BodyHTML     string
}

// CreatedPage is the result of a successful CreatePage call.
type CreatedPage struct {
	ID  string
	URL string
}

// Client talks to the Confluence Cloud REST API.
type Client struct {
	HTTPClient *http.Client
}

// TestConnection verifies credentials against GET /wiki/rest/api/user/current.
func (c *Client) TestConnection(ctx context.Context, cfg Config) error {
	if !cfg.Configured() {
		return errors.New("configuration Confluence incomplète")
	}

	client := c.httpClient(cfg.BaseURL, 10*time.Second)
	baseURL := NormalizeBaseURL(cfg.BaseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+currentUserPath, nil)
	if err != nil {
		return fmt.Errorf("build confluence request: %w", err)
	}
	setAuth(req, cfg)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrConnectionFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("%w: status %d %s", ErrConnectionFailed, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var me struct {
		AccountID string `json:"accountId"`
		Type      string `json:"type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&me); err != nil {
		return fmt.Errorf("%w: invalid response", ErrConnectionFailed)
	}
	return nil
}

// CreatePage creates a Confluence page with storage HTML body and returns its URL.
func (c *Client) CreatePage(ctx context.Context, cfg Config, input CreatePageInput) (CreatedPage, error) {
	if !cfg.Configured() {
		return CreatedPage{}, errors.New("configuration Confluence incomplète")
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		return CreatedPage{}, errors.New("titre Confluence requis")
	}
	spaceKey := strings.ToUpper(strings.TrimSpace(input.SpaceKey))
	if spaceKey == "" {
		spaceKey = strings.ToUpper(strings.TrimSpace(cfg.SpaceKey))
	}
	if spaceKey == "" {
		return CreatedPage{}, errors.New("clé d'espace Confluence requise")
	}

	payload := map[string]any{
		"type":  "page",
		"title": title,
		"space": map[string]string{"key": spaceKey},
		"body": map[string]any{
			"storage": map[string]string{
				"value":          input.BodyHTML,
				"representation": "storage",
			},
		},
	}
	parentID := strings.TrimSpace(input.ParentPageID)
	if parentID == "" {
		parentID = strings.TrimSpace(cfg.ParentPageID)
	}
	if parentID != "" {
		payload["ancestors"] = []map[string]string{{"id": parentID}}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return CreatedPage{}, fmt.Errorf("marshal confluence create payload: %w", err)
	}

	client := c.httpClient(cfg.BaseURL, 20*time.Second)
	baseURL := NormalizeBaseURL(cfg.BaseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+contentPath, bytes.NewReader(body))
	if err != nil {
		return CreatedPage{}, fmt.Errorf("build confluence create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	setAuth(req, cfg)

	resp, err := client.Do(req)
	if err != nil {
		return CreatedPage{}, fmt.Errorf("%w: %w", ErrConnectionFailed, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return CreatedPage{}, fmt.Errorf("%w: status %d %s", ErrCreateFailed, resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var created struct {
		ID    string `json:"id"`
		Links struct {
			Base  string `json:"base"`
			WebUI string `json:"webui"`
		} `json:"_links"`
	}
	if err := json.Unmarshal(respBody, &created); err != nil {
		return CreatedPage{}, fmt.Errorf("%w: invalid response", ErrCreateFailed)
	}
	if created.ID == "" {
		return CreatedPage{}, fmt.Errorf("%w: missing page id", ErrCreateFailed)
	}

	pageURL := pageURLFromLinks(baseURL, created.Links.Base, created.Links.WebUI, created.ID)
	return CreatedPage{ID: created.ID, URL: pageURL}, nil
}

func pageURLFromLinks(cfgBase, linkBase, webUI, pageID string) string {
	webUI = strings.TrimSpace(webUI)
	if webUI != "" {
		base := strings.TrimSpace(linkBase)
		if base == "" {
			base = cfgBase
		}
		if strings.HasPrefix(webUI, "http://") || strings.HasPrefix(webUI, "https://") {
			return webUI
		}
		return strings.TrimRight(base, "/") + webUI
	}
	return strings.TrimRight(cfgBase, "/") + "/wiki/pages/viewpage.action?pageId=" + url.QueryEscape(pageID)
}

func setAuth(req *http.Request, cfg Config) {
	token := base64.StdEncoding.EncodeToString([]byte(cfg.Email + ":" + cfg.APIToken))
	req.Header.Set("Authorization", "Basic "+token)
	req.Header.Set("Accept", "application/json")
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
