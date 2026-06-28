package admin_test

import (
	"context"
	"testing"

	"github.com/jeb-maker/revues/internal/admin"
	"github.com/jeb-maker/revues/internal/integrations/jira"
)

func TestIntegrationsServiceOverview(t *testing.T) {
	ctx := context.Background()
	settings, st := testSettingsService(t)
	jiraSvc := &jira.Service{Store: st, EncryptionKey: settings.EncryptionKey}
	svc := &admin.IntegrationsService{Settings: settings, Jira: jiraSvc}

	overview, err := svc.Overview(ctx)
	if err != nil {
		t.Fatalf("Overview(): %v", err)
	}
	if len(overview.Items) != 3 {
		t.Fatalf("len(Items) = %d, want 3", len(overview.Items))
	}
	for _, item := range overview.Items {
		if item.Enabled {
			t.Fatalf("%s should be disabled when unset", item.Name)
		}
	}

	if saveErr := settings.SaveSMTP(ctx, admin.SMTPConfig{
		Host: "smtp.example.com",
		Port: 587,
		From: "revues@example.com",
	}); saveErr != nil {
		t.Fatalf("SaveSMTP(): %v", saveErr)
	}
	if saveErr := jiraSvc.Save(ctx, jira.Config{
		InstanceType: jira.InstanceCloud,
		BaseURL:      "https://example.atlassian.net",
		Email:        "user@example.com",
		APIToken:     "token",
	}); saveErr != nil {
		t.Fatalf("Save(): %v", saveErr)
	}
	if saveErr := settings.SaveWebhooks(ctx, admin.WebhookConfig{
		URLs:            []string{"https://hooks.example.com/revues"},
		Secret:          "secret",
		ReviewCompleted: true,
	}); saveErr != nil {
		t.Fatalf("SaveWebhooks(): %v", saveErr)
	}

	overview, err = svc.Overview(ctx)
	if err != nil {
		t.Fatalf("Overview() configured: %v", err)
	}
	for _, item := range overview.Items {
		if !item.Enabled {
			t.Fatalf("%s should be enabled", item.Name)
		}
		if item.ConfigPath == "" {
			t.Fatalf("%s missing config path", item.Name)
		}
	}
}
