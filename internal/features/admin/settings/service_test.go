package settings_test

import (
	"context"
	"database/sql"
	"encoding/base64"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/jeb-maker/revues/internal/crypto"
	"github.com/jeb-maker/revues/internal/features/admin/settings"
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

func TestSettingsServiceSaveLoadSMTP(t *testing.T) {
	ctx := context.Background()
	svc, _ := testSettingsService(t)

	cfg := settings.SMTPConfig{
		Host:     "smtp.example.com",
		Port:     587,
		TLS:      true,
		Username: "user",
		Password: "secret",
		From:     "revues@example.com",
	}
	if err := svc.SaveSMTP(ctx, cfg); err != nil {
		t.Fatalf("SaveSMTP(): %v", err)
	}

	got, ok, err := svc.LoadSMTP(ctx)
	if err != nil {
		t.Fatalf("LoadSMTP(): %v", err)
	}
	if !ok {
		t.Fatal("expected configured smtp")
	}
	if got.Host != cfg.Host || got.Port != cfg.Port || got.Username != cfg.Username || got.Password != cfg.Password || got.From != cfg.From || got.TLS != cfg.TLS {
		t.Fatalf("LoadSMTP() = %+v, want %+v", got, cfg)
	}
}

func TestSettingsServiceSaveWithoutKey(t *testing.T) {
	ctx := context.Background()
	svc, _ := testSettingsService(t)
	svc.EncryptionKey = nil

	err := svc.SaveSMTP(ctx, settings.SMTPConfig{Host: "smtp.example.com", Port: 587, From: "a@example.com"})
	if err == nil {
		t.Fatal("expected error without encryption key")
	}
}

func TestValidateSMTP(t *testing.T) {
	if err := settings.ValidateSMTP(settings.SMTPConfig{}); err == nil {
		t.Fatal("expected validation error")
	}
	if err := settings.ValidateSMTP(settings.SMTPConfig{Host: "smtp.example.com", Port: 587, From: "revues@example.com"}); err != nil {
		t.Fatalf("ValidateSMTP(): %v", err)
	}
}

func TestMergePassword(t *testing.T) {
	current := settings.SMTPConfig{Password: "stored"}
	if got := settings.MergePassword(current, ""); got != "stored" {
		t.Fatalf("MergePassword(empty) = %q, want stored", got)
	}
	if got := settings.MergePassword(current, "new"); got != "new" {
		t.Fatalf("MergePassword(new) = %q, want new", got)
	}
}

func TestDecodeKeyUsedByService(t *testing.T) {
	key := make([]byte, crypto.KeySize)
	encoded := base64.StdEncoding.EncodeToString(key)
	decoded, err := crypto.DecodeKey(encoded)
	if err != nil {
		t.Fatalf("DecodeKey(): %v", err)
	}
	if len(decoded) != crypto.KeySize {
		t.Fatalf("key length = %d", len(decoded))
	}
}

func TestValidateWebhooks(t *testing.T) {
	valid := settings.WebhookConfig{URLs: []string{"https://hooks.example.com/revues"}, Secret: "secret", ReviewCompleted: true}
	if err := settings.ValidateWebhooks(valid); err != nil {
		t.Fatalf("ValidateWebhooks(valid): %v", err)
	}
	for name, cfg := range map[string]settings.WebhookConfig{
		"no urls": {Secret: "s", ReviewCompleted: true}, "no secret": {URLs: []string{"https://x.test"}, ReviewCompleted: true},
		"no events": {URLs: []string{"https://x.test"}, Secret: "s"}, "bad url": {URLs: []string{"ftp://x.test"}, Secret: "s", ReviewCompleted: true},
	} {
		t.Run(name, func(t *testing.T) {
			if err := settings.ValidateWebhooks(cfg); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestParseWebhookURLs(t *testing.T) {
	if len(settings.ParseWebhookURLs("https://a.test\nhttps://b.test, https://a.test\n")) != 2 {
		t.Fatal("expected 2 urls")
	}
}
