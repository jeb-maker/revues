package admin_test

import (
	"context"
	"database/sql"
	"encoding/base64"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/jeb-maker/revues/internal/admin"
	"github.com/jeb-maker/revues/internal/crypto"
	"github.com/jeb-maker/revues/internal/store"
)

func testSettingsService(t *testing.T) (*admin.SettingsService, *store.Store) {
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
	return &admin.SettingsService{Store: st, EncryptionKey: key}, st
}

func TestSettingsServiceSaveLoadSMTP(t *testing.T) {
	ctx := context.Background()
	svc, _ := testSettingsService(t)

	cfg := admin.SMTPConfig{
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

	err := svc.SaveSMTP(ctx, admin.SMTPConfig{Host: "smtp.example.com", Port: 587, From: "a@example.com"})
	if err == nil {
		t.Fatal("expected error without encryption key")
	}
}

func TestValidateSMTP(t *testing.T) {
	if err := admin.ValidateSMTP(admin.SMTPConfig{}); err == nil {
		t.Fatal("expected validation error")
	}
	if err := admin.ValidateSMTP(admin.SMTPConfig{Host: "smtp.example.com", Port: 587, From: "revues@example.com"}); err != nil {
		t.Fatalf("ValidateSMTP(): %v", err)
	}
}

func TestMergePassword(t *testing.T) {
	current := admin.SMTPConfig{Password: "stored"}
	if got := admin.MergePassword(current, ""); got != "stored" {
		t.Fatalf("MergePassword(empty) = %q, want stored", got)
	}
	if got := admin.MergePassword(current, "new"); got != "new" {
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
