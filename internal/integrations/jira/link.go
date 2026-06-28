package jira

import (
	"context"
	"errors"
	"fmt"

	"github.com/jeb-maker/revues/internal/store"
)

// ErrNotConfigured is returned when Jira integration is missing or incomplete.
var ErrNotConfigured = errors.New("jira not configured")

// LinkService links run items to Jira issues.
type LinkService struct {
	Store         *store.Store
	EncryptionKey []byte
	Client        *Client
}

// LinkRunItem validates and stores a Jira issue link on a run item.
func (s *LinkService) LinkRunItem(ctx context.Context, runItemID int64, input string) (*store.IntegrationLink, error) {
	cfg, ok, err := s.config(ctx)
	if err != nil {
		return nil, err
	}
	if !ok || !cfg.Configured() {
		return nil, ErrNotConfigured
	}

	key, err := ParseIssueReference(input)
	if err != nil {
		return nil, err
	}

	client := s.client()
	resolvedKey, getErr := client.GetIssue(ctx, cfg, key)
	if getErr != nil {
		return nil, getErr
	}
	_ = resolvedKey

	browseURL := BrowseURL(cfg.BaseURL, key)
	if validateErr := ValidateBrowseURL(cfg, browseURL); validateErr != nil {
		return nil, validateErr
	}

	integration, err := s.Store.GetIntegrationByType(ctx, store.IntegrationTypeJira)
	if err != nil {
		return nil, fmt.Errorf("load jira integration: %w", err)
	}

	link, err := s.Store.UpsertIntegrationLink(ctx, runItemID, integration.ID, key, browseURL)
	if err != nil {
		return nil, fmt.Errorf("store jira link: %w", err)
	}

	return link, nil
}

// Configured reports whether Jira integration is stored and complete.
func (s *LinkService) Configured(ctx context.Context) (bool, error) {
	cfg, ok, err := s.config(ctx)
	if err != nil {
		return false, err
	}
	return ok && cfg.Configured(), nil
}

func (s *LinkService) config(ctx context.Context) (Config, bool, error) {
	svc := &Service{Store: s.Store, EncryptionKey: s.EncryptionKey}
	return svc.Load(ctx)
}

func (s *LinkService) client() *Client {
	if s.Client != nil {
		return s.Client
	}
	return &Client{}
}
