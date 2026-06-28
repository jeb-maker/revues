package notifications_test

import (
	"context"
	"testing"

	"github.com/jeb-maker/revues/internal/admin"
	"github.com/jeb-maker/revues/internal/notifications"
)

func TestMailerNotConfigured(t *testing.T) {
	m := notifications.Mailer{}
	if m.Enabled() {
		t.Fatal("expected disabled mailer")
	}
	if err := m.Send(context.Background(), "a@example.com", "test", "body"); err == nil {
		t.Fatal("expected error when smtp not configured")
	}
}

func TestMailerSend(t *testing.T) {
	host, port := startTestSMTPServer(t)

	m := notifications.Mailer{
		Config: admin.SMTPConfig{
			Host: host,
			Port: port,
			From: "revues@example.com",
		},
	}
	if err := m.Send(context.Background(), "admin@example.com", "Test Revues", "Hello"); err != nil {
		t.Fatalf("Send(): %v", err)
	}
}
