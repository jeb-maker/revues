package notion

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const pagesPath = "/pages"

var ErrExportFailed = errors.New("notion export failed")

type PageItem struct {
	Section, Label, Status, Comment string
}

type CreatePageInput struct {
	DatabaseID, Title, Project, Date, RevuesURL, ClosingNote string
	Items                                                    []PageItem
}

type CreatePageResult struct {
	PageID, URL string
}

func (c *Client) CreateReviewPage(ctx context.Context, cfg Config, in CreatePageInput) (CreatePageResult, error) {
	if !cfg.Configured() {
		return CreatePageResult{}, errors.New("configuration Notion incomplète")
	}
	databaseID := NormalizeDatabaseID(in.DatabaseID)
	if databaseID == "" {
		return CreatePageResult{}, errors.New("identifiant base Notion requis")
	}
	body, err := json.Marshal(buildCreatePagePayload(databaseID, in))
	if err != nil {
		return CreatePageResult{}, fmt.Errorf("marshal notion page: %w", err)
	}
	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiBaseURL()+pagesPath, bytes.NewReader(body))
	if err != nil {
		return CreatePageResult{}, fmt.Errorf("build notion page request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIToken)
	req.Header.Set("Notion-Version", notionVersion)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return CreatePageResult{}, fmt.Errorf("%w: %w", ErrExportFailed, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return CreatePageResult{}, fmt.Errorf("%w: status %d %s", ErrExportFailed, resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	var created struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return CreatePageResult{}, fmt.Errorf("%w: invalid response", ErrExportFailed)
	}
	if strings.TrimSpace(created.URL) == "" {
		return CreatePageResult{}, fmt.Errorf("%w: missing page url", ErrExportFailed)
	}
	return CreatePageResult{PageID: created.ID, URL: created.URL}, nil
}

func buildCreatePagePayload(databaseID string, in CreatePageInput) map[string]any {
	props := map[string]any{
		"Name":        map[string]any{"title": []map[string]any{textSegment(in.Title)}},
		"Projet":      map[string]any{"rich_text": []map[string]any{textSegment(in.Project)}},
		"Lien Revues": map[string]any{"url": in.RevuesURL},
	}
	if date := strings.TrimSpace(in.Date); date != "" {
		props["Date"] = map[string]any{"date": map[string]string{"start": date}}
	}
	return map[string]any{"parent": map[string]string{"database_id": databaseID}, "properties": props, "children": buildPageBlocks(in)}
}

func buildPageBlocks(in CreatePageInput) []map[string]any {
	var blocks []map[string]any
	if note := strings.TrimSpace(in.ClosingNote); note != "" {
		blocks = append(blocks, heading("Note de clôture", 2), paragraph(note))
	}
	blocks = append(blocks, heading("Points de contrôle", 2))
	for _, item := range in.Items {
		blocks = append(blocks, bullet(formatPageItem(item)))
	}
	return blocks
}

func formatPageItem(item PageItem) string {
	var parts []string
	if s := strings.TrimSpace(item.Section); s != "" {
		parts = append(parts, "["+s+"]")
	}
	parts = append(parts, strings.TrimSpace(item.Label), "·", statusLabel(item.Status))
	line := strings.Join(parts, " ")
	if c := strings.TrimSpace(item.Comment); c != "" {
		line += " — " + c
	}
	return line
}

func statusLabel(status string) string {
	switch status {
	case "ok":
		return "OK"
	case "nok":
		return "NOK"
	case "na":
		return "N/A"
	case "pending":
		return "En attente"
	default:
		return status
	}
}

func textSegment(text string) map[string]any {
	return map[string]any{"type": "text", "text": map[string]string{"content": text}}
}

func heading(text string, level int) map[string]any {
	key := fmt.Sprintf("heading_%d", level)
	return map[string]any{"object": "block", "type": key, key: map[string]any{"rich_text": []map[string]any{textSegment(text)}}}
}

func paragraph(text string) map[string]any {
	return map[string]any{"object": "block", "type": "paragraph", "paragraph": map[string]any{"rich_text": []map[string]any{textSegment(text)}}}
}

func bullet(text string) map[string]any {
	return map[string]any{"object": "block", "type": "bulleted_list_item", "bulleted_list_item": map[string]any{"rich_text": []map[string]any{textSegment(text)}}}
}
