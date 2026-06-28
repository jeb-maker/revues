package admin_test

import (
	"testing"

	"github.com/jeb-maker/revues/internal/admin"
)

func TestValidateWebhooks(t *testing.T) {
	valid := admin.WebhookConfig{URLs: []string{"https://hooks.example.com/revues"}, Secret: "secret", ReviewCompleted: true}
	if err := admin.ValidateWebhooks(valid); err != nil {
		t.Fatalf("ValidateWebhooks(valid): %v", err)
	}
	for name, cfg := range map[string]admin.WebhookConfig{
		"no urls": {Secret: "s", ReviewCompleted: true}, "no secret": {URLs: []string{"https://x.test"}, ReviewCompleted: true},
		"no events": {URLs: []string{"https://x.test"}, Secret: "s"}, "bad url": {URLs: []string{"ftp://x.test"}, Secret: "s", ReviewCompleted: true},
	} {
		t.Run(name, func(t *testing.T) {
			if err := admin.ValidateWebhooks(cfg); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestParseWebhookURLs(t *testing.T) {
	if len(admin.ParseWebhookURLs("https://a.test\nhttps://b.test, https://a.test\n")) != 2 {
		t.Fatal("expected 2 urls")
	}
}
