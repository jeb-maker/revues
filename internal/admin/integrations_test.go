package admin_test

import (
	"context"
	"testing"

	"github.com/jeb-maker/revues/internal/admin"
	"github.com/jeb-maker/revues/internal/integrations/jira"
	"github.com/jeb-maker/revues/internal/integrations/notion"
)

func TestIntegrationsServiceOverview(t *testing.T) {
	ctx := context.Background()
	settings, st := testSettingsService(t)
	jiraSvc := &jira.Service{Store: st, EncryptionKey: settings.EncryptionKey}
	notionSvc := &notion.Service{Store: st, EncryptionKey: settings.EncryptionKey}
	svc := &admin.IntegrationsService{Settings: settings, Jira: jiraSvc, Notion: notionSvc}

	overview, err := svc.Overview(ctx)
	if err != nil {
		t.Fatalf("Overview(): %v", err)
	}
	if len(overview.Items) != 4 {
		t.Fatalf("len(Items) = %d, want 4", len(overview.Items))
	}
	for _, item := range overview.Items {
		if item.Enabled {
			t.Fatalf("%s should be disabled when unset", item.Name)
		}
	}

	_ = settings.SaveSMTP(ctx, admin.SMTPConfig{Host: "smtp.example.com", Port: 587, From: "revues@example.com"})
	_ = jiraSvc.Save(ctx, jira.Config{InstanceType: jira.InstanceCloud, BaseURL: "https://example.atlassian.net", Email: "user@example.com", APIToken: "token"})
	_ = notionSvc.Save(ctx, notion.Config{APIToken: "notion-token"})
	_ = settings.SaveWebhooks(ctx, admin.WebhookConfig{URLs: []string{"https://hooks.example.com/revues"}, Secret: "secret", ReviewCompleted: true})

	overview, err = svc.Overview(ctx)
	if err != nil {
		t.Fatalf("Overview() configured: %v", err)
	}
	for _, item := range overview.Items {
		if !item.Enabled {
			t.Fatalf("%s should be enabled", item.Name)
		}
	}
}
