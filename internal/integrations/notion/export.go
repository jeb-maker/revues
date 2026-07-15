package notion

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jeb-maker/revues/internal/store"
)

var (
	ErrNotConfigured   = errors.New("notion not configured")
	ErrDatabaseMissing = errors.New("notion database not configured")
	ErrRunNotDone      = errors.New("run is not done")
	ErrAlreadyExported = errors.New("run already exported to notion")
)

type ExportService struct {
	Store         *store.Store
	EncryptionKey []byte
	Client        *Client
	BaseURL       string
}

func (s *ExportService) ExportRun(ctx context.Context, runID int64) (string, error) {
	cfg, ok, err := s.config(ctx)
	if err != nil {
		return "", err
	}
	if !ok || !cfg.Configured() {
		return "", ErrNotConfigured
	}
	if NormalizeDatabaseID(cfg.DefaultDatabaseID) == "" {
		return "", ErrDatabaseMissing
	}
	run, err := s.Store.RunByID(ctx, runID)
	if err != nil {
		return "", err
	}
	if run.Status != store.RunStatusDone {
		return "", ErrRunNotDone
	}
	if strings.TrimSpace(run.NotionURL) != "" {
		return "", ErrAlreadyExported
	}
	subject, err := s.Store.SubjectByID(ctx, run.SubjectID)
	if err != nil {
		return "", fmt.Errorf("load subject: %w", err)
	}
	items, err := s.Store.ListRunItems(ctx, runID)
	if err != nil {
		return "", fmt.Errorf("list run items: %w", err)
	}
	pageItems := make([]PageItem, len(items))
	for i, item := range items {
		pageItems[i] = PageItem{Section: item.Section, Label: item.Label, Status: item.Status, Comment: item.Comment}
	}
	versionInfo, err := s.Store.TemplateVersionInfo(ctx, run.TemplateVersionID)
	if err != nil {
		return "", fmt.Errorf("template version info: %w", err)
	}
	displayLabel := store.RunDisplayLabel(versionInfo.Name, subject.Name, run.CreatedAt, run.ID)
	result, err := s.client().CreateReviewPage(ctx, cfg, CreatePageInput{
		DatabaseID: cfg.DefaultDatabaseID, Title: displayLabel, Subject: subject.Name,
		Date: exportDate(run.CompletedAt), RevuesURL: runURL(s.BaseURL, runID),
		ClosingNote: run.ClosingNote, Items: pageItems,
	})
	if err != nil {
		return "", err
	}
	if err := s.Store.SetRunNotionURL(ctx, runID, result.URL); err != nil {
		return "", fmt.Errorf("store notion url: %w", err)
	}
	return result.URL, nil
}

func ExportReady(cfg Config) bool {
	return cfg.Configured() && NormalizeDatabaseID(cfg.DefaultDatabaseID) != ""
}

func (s *ExportService) config(ctx context.Context) (Config, bool, error) {
	return (&Service{Store: s.Store, EncryptionKey: s.EncryptionKey}).Load(ctx)
}

func (s *ExportService) client() *Client {
	if s.Client != nil {
		return s.Client
	}
	return &Client{}
}

func exportDate(completedAt sql.NullString) string {
	if !completedAt.Valid || strings.TrimSpace(completedAt.String) == "" {
		return time.Now().UTC().Format("2006-01-02")
	}
	t, err := time.Parse(time.RFC3339, completedAt.String)
	if err != nil {
		if len(completedAt.String) >= 10 {
			return completedAt.String[:10]
		}
		return time.Now().UTC().Format("2006-01-02")
	}
	return t.UTC().Format("2006-01-02")
}

func runURL(baseURL string, runID int64) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		return fmt.Sprintf("/runs/%d", runID)
	}
	return fmt.Sprintf("%s/runs/%d", base, runID)
}
