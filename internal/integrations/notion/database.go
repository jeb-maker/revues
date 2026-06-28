package notion

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	databasePath      = "/databases/"
	databaseQueryPath = "/databases/%s/query"
	maxQueryPages     = 500
)

var (
	ErrDatabaseNotFound = errors.New("notion database not found")
	databaseURLIDRe     = regexp.MustCompile(`([0-9a-fA-F]{32})(?:\?[^#]*)?(?:#.*)?$`)
	databaseUUIDRe      = regexp.MustCompile(`([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})`)
)

type PropertyInfo struct {
	Name string
	Type string
}

type DatabaseInfo struct {
	ID         string
	Title      string
	Properties []PropertyInfo
}

type DatabasePage struct {
	Properties map[string]json.RawMessage
}

func ParseDatabaseRef(input string) (string, error) {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return "", errors.New("URL ou identifiant de base Notion requis")
	}
	id := NormalizeDatabaseID(raw)
	if len(id) == 32 && isHex32(id) {
		return strings.ToLower(id), nil
	}
	trimmed := strings.TrimRight(raw, "/")
	if m := databaseURLIDRe.FindStringSubmatch(trimmed); len(m) > 1 {
		return strings.ToLower(m[1]), nil
	}
	if m := databaseUUIDRe.FindStringSubmatch(raw); len(m) > 1 {
		return NormalizeDatabaseID(m[1]), nil
	}
	return "", errors.New("identifiant base Notion invalide")
}

func isHex32(s string) bool {
	for _, c := range s {
		switch {
		case c >= '0' && c <= '9', c >= 'a' && c <= 'f', c >= 'A' && c <= 'F':
		default:
			return false
		}
	}
	return true
}

func (c *Client) GetDatabase(ctx context.Context, cfg Config, databaseID string) (DatabaseInfo, error) {
	if !cfg.Configured() {
		return DatabaseInfo{}, errors.New("configuration Notion incomplète")
	}
	id, err := ParseDatabaseRef(databaseID)
	if err != nil {
		return DatabaseInfo{}, err
	}
	resp, err := c.doJSON(ctx, cfg, http.MethodGet, databasePath+id, nil)
	if err != nil {
		return DatabaseInfo{}, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return DatabaseInfo{}, ErrDatabaseNotFound
	default:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return DatabaseInfo{}, fmt.Errorf("%w: status %d %s", ErrConnectionFailed, resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var payload struct {
		ID         string                       `json:"id"`
		Title      []plainTextFragment          `json:"title"`
		Properties map[string]propertyTypeField `json:"properties"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return DatabaseInfo{}, fmt.Errorf("%w: invalid database response", ErrConnectionFailed)
	}
	info := DatabaseInfo{ID: NormalizeDatabaseID(payload.ID), Title: joinPlainText(payload.Title)}
	for name, prop := range payload.Properties {
		info.Properties = append(info.Properties, PropertyInfo{Name: name, Type: prop.Type})
	}
	return info, nil
}

func (c *Client) QueryDatabase(ctx context.Context, cfg Config, databaseID string) ([]DatabasePage, error) {
	if !cfg.Configured() {
		return nil, errors.New("configuration Notion incomplète")
	}
	id, err := ParseDatabaseRef(databaseID)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf(databaseQueryPath, id)
	var pages []DatabasePage
	var startCursor string
	for len(pages) < maxQueryPages {
		body := map[string]any{}
		if startCursor != "" {
			body["start_cursor"] = startCursor
		}
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal query body: %w", err)
		}
		resp, err := c.doJSON(ctx, cfg, http.MethodPost, path, bytes.NewReader(raw))
		if err != nil {
			return nil, err
		}
		var payload struct {
			Results []struct {
				Properties map[string]json.RawMessage `json:"properties"`
			} `json:"results"`
			HasMore    bool   `json:"has_more"`
			NextCursor string `json:"next_cursor"`
		}
		decodeErr := json.NewDecoder(resp.Body).Decode(&payload)
		resp.Body.Close()
		if decodeErr != nil {
			return nil, fmt.Errorf("%w: invalid query response", ErrConnectionFailed)
		}
		for _, row := range payload.Results {
			pages = append(pages, DatabasePage{Properties: row.Properties})
		}
		if !payload.HasMore || payload.NextCursor == "" {
			break
		}
		startCursor = payload.NextCursor
	}
	return pages, nil
}

type plainTextFragment struct {
	PlainText string `json:"plain_text"`
}

type propertyTypeField struct {
	Type string `json:"type"`
}

func (c *Client) doJSON(ctx context.Context, cfg Config, method, path string, body io.Reader) (*http.Response, error) {
	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, method, c.apiBaseURL()+path, body)
	if err != nil {
		return nil, fmt.Errorf("build notion request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIToken)
	req.Header.Set("Notion-Version", notionVersion)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrConnectionFailed, err)
	}
	return resp, nil
}

func joinPlainText(parts []plainTextFragment) string {
	var b strings.Builder
	for _, p := range parts {
		b.WriteString(p.PlainText)
	}
	return strings.TrimSpace(b.String())
}
