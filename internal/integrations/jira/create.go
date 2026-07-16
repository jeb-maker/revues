package jira

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jeb-maker/revues/internal/store"
)

// ErrProjectKeyMissing is returned when Jira config has no default project key.
var ErrProjectKeyMissing = errors.New("jira project key not configured")

// ErrNotNOK is returned when a run item is not marked nok.
var ErrNotNOK = errors.New("run item is not nok")

// ErrAlreadyLinked is returned when a Jira issue is already linked.
var ErrAlreadyLinked = errors.New("jira issue already linked")

// CreateService creates Jira issues from run items.
type CreateService struct {
	Store         *store.Store
	EncryptionKey []byte
	Client        *Client
}

// CreateInput holds user-editable issue fields.
type CreateInput struct {
	Title       string
	Description string
}

// RunItemContext provides metadata for pre-filled issue content.
type RunItemContext struct {
	SubjectName string
	RunTitle    string
	ItemURL     string
}

// DefaultIssueContent builds pre-filled title and description for a nok run item.
func DefaultIssueContent(item *store.RunItem, ctx RunItemContext) (title, description string) {
	title = strings.TrimSpace(item.Label)
	if title == "" {
		title = "Point non conforme"
	}

	var lines []string
	if ctx.SubjectName != "" {
		lines = append(lines, "Sujet : "+ctx.SubjectName)
	}
	if ctx.RunTitle != "" {
		lines = append(lines, "Revue : "+ctx.RunTitle)
	}
	if item.Section != "" {
		lines = append(lines, "Section : "+item.Section)
	}
	if item.Comment != "" {
		lines = append(lines, "Commentaire : "+item.Comment)
	}
	if ctx.ItemURL != "" {
		lines = append(lines, "Lien Revues : "+ctx.ItemURL)
	}
	description = strings.Join(lines, "\n")
	return title, description
}

// CreateRunItem creates a Jira issue for a nok run item and stores the link.
func (s *CreateService) CreateRunItem(ctx context.Context, runID, runItemID int64, input CreateInput, itemCtx RunItemContext) (*store.IntegrationLink, error) {
	cfg, ok, err := s.config(ctx)
	if err != nil {
		return nil, err
	}
	if !ok || !cfg.Configured() {
		return nil, ErrNotConfigured
	}
	if strings.TrimSpace(cfg.ProjectKey) == "" {
		return nil, ErrProjectKeyMissing
	}

	item, err := s.Store.RunItemByID(ctx, runID, runItemID)
	if err != nil {
		return nil, err
	}
	if item.Status != store.RunItemStatusNOK {
		return nil, ErrNotNOK
	}

	existing, err := s.Store.IntegrationLinkByRunItemAndType(ctx, runItemID, store.IntegrationTypeJira)
	if err == nil && existing != nil {
		return nil, ErrAlreadyLinked
	}
	if err != nil && !errors.Is(err, store.ErrIntegrationLinkNotFound) {
		return nil, fmt.Errorf("check jira link: %w", err)
	}

	title := strings.TrimSpace(input.Title)
	description := strings.TrimSpace(input.Description)
	if title == "" || description == "" {
		defaultTitle, defaultDesc := DefaultIssueContent(item, itemCtx)
		if title == "" {
			title = defaultTitle
		}
		if description == "" {
			description = defaultDesc
		}
	}

	key, err := s.client().CreateIssue(ctx, cfg, CreateIssueInput{
		ProjectKey:  cfg.ProjectKey,
		IssueType:   cfg.IssueType,
		Summary:     title,
		Description: description,
	})
	if err != nil {
		return nil, err
	}

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

func (s *CreateService) config(ctx context.Context) (Config, bool, error) {
	svc := &Service{Store: s.Store, EncryptionKey: s.EncryptionKey}
	return svc.Load(ctx)
}

func (s *CreateService) client() *Client {
	if s.Client != nil {
		return s.Client
	}
	return &Client{}
}
