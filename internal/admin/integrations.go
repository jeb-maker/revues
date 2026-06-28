package admin

import (
	"context"
	"fmt"

	"github.com/jeb-maker/revues/internal/integrations/jira"
)

const (
	integrationPathSMTP     = "/admin/settings/smtp"
	integrationPathWebhooks = "/admin/settings/webhooks"
	integrationPathJira     = "/admin/integrations/jira"
)

// IntegrationSummary is one row on the admin integrations overview.
type IntegrationSummary struct {
	Name        string
	Description string
	Enabled     bool
	ConfigPath  string
}

// IntegrationsOverview lists configured integrations and their status.
type IntegrationsOverview struct {
	Items []IntegrationSummary
}

// IntegrationsService loads integration status for the admin overview.
type IntegrationsService struct {
	Settings *SettingsService
	Jira     *jira.Service
}

// Overview returns enabled/disabled status for SMTP, Jira, and webhooks.
func (s *IntegrationsService) Overview(ctx context.Context) (IntegrationsOverview, error) {
	if s.Settings == nil {
		return IntegrationsOverview{}, fmt.Errorf("settings service required")
	}
	if s.Jira == nil {
		return IntegrationsOverview{}, fmt.Errorf("jira service required")
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

	return IntegrationsOverview{
		Items: []IntegrationSummary{
			{
				Name:        "SMTP",
				Description: "Relais email pour les notifications.",
				Enabled:     smtpEnabled,
				ConfigPath:  integrationPathSMTP,
			},
			{
				Name:        "Jira",
				Description: "Lier et créer des tickets depuis les revues.",
				Enabled:     jiraEnabled,
				ConfigPath:  integrationPathJira,
			},
			{
				Name:        "Webhooks",
				Description: "Notifications JSON signées vers des URLs externes.",
				Enabled:     webhooksEnabled,
				ConfigPath:  integrationPathWebhooks,
			},
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
