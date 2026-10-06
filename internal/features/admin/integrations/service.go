package integrations

import (
	"context"
	"fmt"

	"github.com/jeb-maker/revues/internal/features/admin/settings"
	"github.com/jeb-maker/revues/internal/integrations/confluence"
	"github.com/jeb-maker/revues/internal/integrations/jira"
)

const (
	IntegrationKeySMTP       = "smtp"
	IntegrationKeyJira       = "jira"
	IntegrationKeyWebhooks   = "webhooks"
	IntegrationKeyConfluence = "confluence"

	integrationPathSMTP       = "/admin/settings/smtp"
	integrationPathWebhooks   = "/admin/settings/webhooks"
	integrationPathJira       = "/admin/integrations/jira"
	integrationPathConfluence = "/admin/integrations/confluence"
)

// IntegrationSummary describes one integration row on the admin overview.
type IntegrationSummary struct {
	Key         string
	Name        string
	Description string
	Enabled     bool
	ConfigPath  string
}

// IntegrationsOverview aggregates integration statuses for the admin page.
type IntegrationsOverview struct {
	Items []IntegrationSummary
}

// IntegrationsService builds the integrations overview from configured services.
type IntegrationsService struct {
	Settings   *settings.SettingsService
	Jira       *jira.Service
	Confluence *confluence.Service
}

// Overview returns the configured status of every integration.
func (s *IntegrationsService) Overview(ctx context.Context) (IntegrationsOverview, error) {
	if s.Settings == nil {
		return IntegrationsOverview{}, fmt.Errorf("settings service required")
	}
	if s.Jira == nil {
		return IntegrationsOverview{}, fmt.Errorf("jira service required")
	}
	if s.Confluence == nil {
		return IntegrationsOverview{}, fmt.Errorf("confluence service required")
	}
	smtpEnabled, err := s.smtpEnabled(ctx)
	if err != nil {
		return IntegrationsOverview{}, err
	}
	webhooksEnabled, err := s.webhooksEnabled(ctx)
	if err != nil {
		return IntegrationsOverview{}, err
	}
	jiraEnabled, err := s.jiraEnabled(ctx)
	if err != nil {
		return IntegrationsOverview{}, err
	}
	confluenceEnabled, err := s.confluenceEnabled(ctx)
	if err != nil {
		return IntegrationsOverview{}, err
	}
	return IntegrationsOverview{
		Items: []IntegrationSummary{
			{Key: IntegrationKeySMTP, Name: "SMTP", Description: "Relais email pour les notifications.", Enabled: smtpEnabled, ConfigPath: integrationPathSMTP},
			{Key: IntegrationKeyJira, Name: "Jira", Description: "Lier et créer des tickets depuis les revues.", Enabled: jiraEnabled, ConfigPath: integrationPathJira},
			{Key: IntegrationKeyConfluence, Name: "Confluence", Description: "Publier une revue clôturée vers une page wiki.", Enabled: confluenceEnabled, ConfigPath: integrationPathConfluence},
			{Key: IntegrationKeyWebhooks, Name: "Webhooks", Description: "Notifications JSON signées vers des URLs externes.", Enabled: webhooksEnabled, ConfigPath: integrationPathWebhooks},
		},
	}, nil
}

func (s *IntegrationsService) smtpEnabled(ctx context.Context) (bool, error) {
	cfg, ok, err := s.Settings.LoadSMTP(ctx)
	if err != nil {
		return false, fmt.Errorf("load smtp: %w", err)
	}
	if !ok {
		return false, nil
	}
	return cfg.Enabled(), nil
}

func (s *IntegrationsService) webhooksEnabled(ctx context.Context) (bool, error) {
	cfg, ok, err := s.Settings.LoadWebhooks(ctx)
	if err != nil {
		return false, fmt.Errorf("load webhooks: %w", err)
	}
	if !ok {
		return false, nil
	}
	return cfg.Enabled(), nil
}

func (s *IntegrationsService) jiraEnabled(ctx context.Context) (bool, error) {
	cfg, ok, err := s.Jira.Load(ctx)
	if err != nil {
		return false, fmt.Errorf("load jira: %w", err)
	}
	if !ok {
		return false, nil
	}
	return cfg.Configured(), nil
}

func (s *IntegrationsService) confluenceEnabled(ctx context.Context) (bool, error) {
	cfg, ok, err := s.Confluence.Load(ctx)
	if err != nil {
		return false, fmt.Errorf("load confluence: %w", err)
	}
	if !ok {
		return false, nil
	}
	return cfg.Configured(), nil
}
