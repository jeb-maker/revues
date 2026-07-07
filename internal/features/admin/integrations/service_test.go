package integrations_test

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/jeb-maker/revues/internal/crypto"
	"github.com/jeb-maker/revues/internal/features/admin/integrations"
	"github.com/jeb-maker/revues/internal/features/admin/settings"
	"github.com/jeb-maker/revues/internal/integrations/jira"
	"github.com/jeb-maker/revues/internal/integrations/notion"
	"github.com/jeb-maker/revues/internal/store"
)

func testSettingsService(t *testing.T) (*settings.SettingsService, *store.Store) {
	t.Helper()

	ctx := context.Background()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("Close(): %v", err)
		}
	})
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("foreign_keys: %v", err)
	}
	if err := store.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate(): %v", err)
	}

	key := make([]byte, crypto.KeySize)
	st := store.New(db)
	return &settings.SettingsService{Store: st, EncryptionKey: key}, st
}

func TestIntegrationsServiceOverview(t *testing.T) {
	ctx := context.Background()
	settingsSvc, st := testSettingsService(t)
	jiraSvc := &jira.Service{Store: st, EncryptionKey: settingsSvc.EncryptionKey}
	notionSvc := &notion.Service{Store: st, EncryptionKey: settingsSvc.EncryptionKey}
	svc := &integrations.IntegrationsService{Settings: settingsSvc, Jira: jiraSvc, Notion: notionSvc}

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

	_ = settingsSvc.SaveSMTP(ctx, settings.SMTPConfig{Host: "smtp.example.com", Port: 587, From: "revues@example.com"})
	_ = jiraSvc.Save(ctx, jira.Config{InstanceType: jira.InstanceCloud, BaseURL: "https://example.atlassian.net", Email: "user@example.com", APIToken: "token"})
	_ = notionSvc.Save(ctx, notion.Config{APIToken: "notion-token"})
	_ = settingsSvc.SaveWebhooks(ctx, settings.WebhookConfig{URLs: []string{"https://hooks.example.com/revues"}, Secret: "secret", ReviewCompleted: true})

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
