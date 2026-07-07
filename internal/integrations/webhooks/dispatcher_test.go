package webhooks_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/features/admin/settings"
	"github.com/jeb-maker/revues/internal/integrations/webhooks"
)

func TestWebhook_HMAC(t *testing.T) {
	secret := "super-secret-key"
	body := []byte(`{"event_id":"abc","event_type":"webhook.test"}`)
	sig := webhooks.SignBody(secret, body)
	if !strings.HasPrefix(sig, "sha256=") {
		t.Fatalf("signature = %q", sig)
	}
	if !webhooks.VerifySignature(secret, body, sig) {
		t.Fatal("valid signature rejected")
	}
	if webhooks.VerifySignature(secret, body, "sha256=deadbeef") {
		t.Fatal("invalid signature accepted")
	}
}

func TestWebhook_SSRF_Block(t *testing.T) {
	tests := []struct {
		url, name string
		dev       bool
		wantErr   bool
	}{
		{"https://example.com/hook", "https ok", false, false},
		{"http://example.com/hook", "http prod", false, true},
		{"http://127.0.0.1:8080/hook", "localhost dev", true, false},
		{"file:///etc/passwd", "file", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := webhooks.ValidateTargetURL(tt.url, tt.dev)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestWebhook_SSRF_BlockPrivateDial(t *testing.T) {
	client := webhooks.NewSafeClient(false)
	ctx := context.Background()
	for _, target := range []string{"https://192.168.0.1/hook", "https://169.254.169.254/latest/meta-data"} {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, strings.NewReader("{}"))
		if err != nil {
			t.Fatal(err)
		}
		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			t.Fatalf("expected block for %s", target)
		}
	}
}

func TestDispatcher_SendTest(t *testing.T) {
	var gotSig string
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSig = r.Header.Get("X-Revues-Signature")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	cfg := settings.WebhookConfig{URLs: []string{srv.URL}, Secret: "test-secret", ReviewCompleted: true}
	d := &webhooks.Dispatcher{Settings: stubSettings{cfg, true}, Store: stubStore{}, DevMode: true, Client: srv.Client()}
	if err := d.SendTest(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !webhooks.VerifySignature(cfg.Secret, gotBody, gotSig) {
		t.Fatal("bad signature")
	}
	var env webhooks.Envelope
	if err := json.Unmarshal(gotBody, &env); err != nil || env.EventType != webhooks.EventTest {
		t.Fatal("bad payload")
	}
}

func TestValidateTargetURL_ResolvesLoopback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	client := webhooks.NewSafeClient(true)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/hook", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
}

type stubSettings struct {
	cfg settings.WebhookConfig
	ok  bool
}

func (s stubSettings) LoadWebhooks(context.Context) (settings.WebhookConfig, bool, error) {
	return s.cfg, s.ok, nil
}

type stubStore struct{}

func (stubStore) InsertWebhookDelivery(context.Context, string, string, string, int, bool) error {
	return nil
}
