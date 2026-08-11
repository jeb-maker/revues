package webhooks_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jeb-maker/revues/internal/features/admin/settings"
	"github.com/jeb-maker/revues/internal/integrations/webhooks"
	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
)

func testOrgCtx() context.Context {
	return orgctx.WithOrganizationID(context.Background(), 1)
}

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
	mem := newMemStore()
	d := &webhooks.Dispatcher{Settings: stubSettings{cfg, true}, Store: mem, DevMode: true, Client: srv.Client()}
	if err := d.SendTest(testOrgCtx()); err != nil {
		t.Fatal(err)
	}
	if !webhooks.VerifySignature(cfg.Secret, gotBody, gotSig) {
		t.Fatal("bad signature")
	}
	var env webhooks.Envelope
	if err := json.Unmarshal(gotBody, &env); err != nil || env.EventType != webhooks.EventTest {
		t.Fatal("bad payload")
	}
	if len(mem.byID) != 1 || mem.byID[1].State != store.WebhookDeliveryDone {
		t.Fatalf("delivery state = %+v", mem.byID)
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

func TestBackoffAfterAttempt(t *testing.T) {
	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{1, time.Minute},
		{2, 2 * time.Minute},
		{3, 4 * time.Minute},
		{4, 8 * time.Minute},
		{5, 16 * time.Minute},
		{6, 30 * time.Minute},
		{10, 30 * time.Minute},
	}
	for _, tt := range tests {
		if got := webhooks.BackoffAfterAttempt(tt.attempt); got != tt.want {
			t.Fatalf("attempt %d: got %v want %v", tt.attempt, got, tt.want)
		}
	}
}

func TestDispatcher_Drain_RetriesThenSucceeds(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := hits.Add(1)
		if n == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	mem := newMemStore()
	cfg := settings.WebhookConfig{URLs: []string{srv.URL}, Secret: "s", ReviewCompleted: true}
	d := &webhooks.Dispatcher{
		Settings: stubSettings{cfg, true},
		Store:    mem,
		DevMode:  true,
		Client:   srv.Client(),
		Now:      func() time.Time { return now },
	}
	if err := d.SendTest(testOrgCtx()); err == nil {
		t.Fatal("expected first attempt failure")
	}
	del := mem.byID[1]
	if del.State != store.WebhookDeliveryPending || del.Attempts != 1 {
		t.Fatalf("after fail: %+v", del)
	}
	next, err := time.Parse(time.RFC3339, del.NextAttemptAt)
	if err != nil {
		t.Fatal(err)
	}
	now = next
	if err := d.Drain(context.Background()); err != nil {
		t.Fatal(err)
	}
	del = mem.byID[1]
	if del.State != store.WebhookDeliveryDone || del.Attempts != 2 || hits.Load() != 2 {
		t.Fatalf("after drain: %+v hits=%d", del, hits.Load())
	}
}

func TestDispatcher_Drain_PoisonAfterMaxAttempts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	mem := newMemStore()
	cfg := settings.WebhookConfig{URLs: []string{srv.URL}, Secret: "s", ReviewCompleted: true}
	d := &webhooks.Dispatcher{
		Settings: stubSettings{cfg, true},
		Store:    mem,
		DevMode:  true,
		Client:   srv.Client(),
		Now:      func() time.Time { return now },
	}
	_ = d.SendTest(testOrgCtx())
	for i := 0; i < webhooks.MaxAttempts; i++ {
		del := mem.byID[1]
		if del.State != store.WebhookDeliveryPending {
			break
		}
		next, err := time.Parse(time.RFC3339, del.NextAttemptAt)
		if err != nil {
			t.Fatal(err)
		}
		now = next
		if err := d.Drain(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	del := mem.byID[1]
	if del.State != store.WebhookDeliveryPoison || del.Attempts != webhooks.MaxAttempts {
		t.Fatalf("want poison after %d attempts, got %+v", webhooks.MaxAttempts, del)
	}
}

func TestDispatcher_Drain_PoisonAfterTTL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	t.Cleanup(srv.Close)

	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	mem := newMemStore()
	cfg := settings.WebhookConfig{URLs: []string{srv.URL}, Secret: "s", ReviewCompleted: true}
	d := &webhooks.Dispatcher{
		Settings: stubSettings{cfg, true},
		Store:    mem,
		DevMode:  true,
		Client:   srv.Client(),
		Now:      func() time.Time { return now },
	}
	_ = d.SendTest(testOrgCtx())
	del := mem.byID[1]
	if del.State != store.WebhookDeliveryPending {
		t.Fatalf("expected pending, got %+v", del)
	}
	now = now.Add(webhooks.DeliveryTTL + time.Second)
	if err := d.Drain(context.Background()); err != nil {
		t.Fatal(err)
	}
	del = mem.byID[1]
	if del.State != store.WebhookDeliveryPoison || !strings.Contains(del.LastError, "ttl") {
		t.Fatalf("want ttl poison, got %+v", del)
	}
}

func TestDispatcher_Drain_ReChecksSSRF(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	mem := newMemStore()
	id, err := mem.EnqueueWebhookDelivery(testOrgCtx(), "e1", webhooks.EventTest, "https://192.168.1.10/hook", []byte(`{}`), now, now.Add(webhooks.DeliveryTTL))
	if err != nil {
		t.Fatal(err)
	}
	cfg := settings.WebhookConfig{URLs: []string{"https://example.com"}, Secret: "s", ReviewCompleted: true}
	d := &webhooks.Dispatcher{
		Settings: stubSettings{cfg, true},
		Store:    mem,
		DevMode:  false,
		Client:   webhooks.NewSafeClient(false),
		Now:      func() time.Time { return now },
	}
	if err := d.Drain(context.Background()); err != nil {
		t.Fatal(err)
	}
	del := mem.byID[id]
	if del.State != store.WebhookDeliveryPoison || !strings.Contains(del.LastError, "blocked ip") {
		t.Fatalf("want SSRF poison, got %+v", del)
	}
}

func TestDispatcher_Drain_UsesOrgScopedSecret(t *testing.T) {
	var gotSig string
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSig = r.Header.Get("X-Revues-Signature")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	mem := newMemStore()
	orgA := int64(7)
	_, err := mem.EnqueueWebhookDelivery(
		orgctx.WithOrganizationID(context.Background(), orgA),
		"e-org", webhooks.EventTest, srv.URL, []byte(`{"x":1}`), now, now.Add(webhooks.DeliveryTTL),
	)
	if err != nil {
		t.Fatal(err)
	}

	secretA := "secret-for-org-a-only"
	loader := &orgScopedSettings{byOrg: map[int64]settings.WebhookConfig{
		orgA: {URLs: []string{srv.URL}, Secret: secretA, ReviewCompleted: true},
		99:   {URLs: []string{srv.URL}, Secret: "wrong-org-secret!!!!!!!!!!!", ReviewCompleted: true},
	}}
	d := &webhooks.Dispatcher{
		Settings: loader,
		Store:    mem,
		DevMode:  true,
		Client:   srv.Client(),
		Now:      func() time.Time { return now },
	}
	if err := d.Drain(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !webhooks.VerifySignature(secretA, gotBody, gotSig) {
		t.Fatalf("drain signed with wrong org secret: sig=%q body=%s", gotSig, gotBody)
	}
	if mem.byID[1].State != store.WebhookDeliveryDone {
		t.Fatalf("delivery = %+v", mem.byID[1])
	}
}

type orgScopedSettings struct {
	byOrg map[int64]settings.WebhookConfig
}

func (s *orgScopedSettings) LoadWebhooks(ctx context.Context) (settings.WebhookConfig, bool, error) {
	orgID, ok := orgctx.OrganizationID(ctx)
	if !ok {
		return settings.WebhookConfig{}, false, store.ErrOrganizationRequired
	}
	cfg, ok := s.byOrg[orgID]
	return cfg, ok, nil
}

type stubSettings struct {
	cfg settings.WebhookConfig
	ok  bool
}

func (s stubSettings) LoadWebhooks(context.Context) (settings.WebhookConfig, bool, error) {
	return s.cfg, s.ok, nil
}

type memStore struct {
	mu   sync.Mutex
	seq  int64
	byID map[int64]store.WebhookDelivery
}

func newMemStore() *memStore {
	return &memStore{byID: make(map[int64]store.WebhookDelivery)}
}

func (m *memStore) EnqueueWebhookDelivery(ctx context.Context, eventID, eventType, url string, payload []byte, nextAttemptAt, expiresAt time.Time) (int64, error) {
	orgID, ok := orgctx.OrganizationID(ctx)
	if !ok {
		return 0, store.ErrOrganizationRequired
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seq++
	id := m.seq
	m.byID[id] = store.WebhookDelivery{
		ID:             id,
		OrganizationID: orgID,
		EventID:        eventID,
		EventType:      eventType,
		URL:            url,
		Payload:        append([]byte(nil), payload...),
		Attempts:       0,
		NextAttemptAt:  nextAttemptAt.UTC().Format(time.RFC3339),
		ExpiresAt:      expiresAt.UTC().Format(time.RFC3339),
		State:          store.WebhookDeliveryPending,
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
	}
	return id, nil
}

func (m *memStore) ListDueWebhookDeliveries(_ context.Context, now time.Time, limit int) ([]store.WebhookDelivery, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	nowStr := now.UTC().Format(time.RFC3339)
	var out []store.WebhookDelivery
	for _, d := range m.byID {
		if d.State != store.WebhookDeliveryPending || d.NextAttemptAt == "" || d.NextAttemptAt > nowStr {
			continue
		}
		cp := d
		cp.Payload = append([]byte(nil), d.Payload...)
		out = append(out, cp)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (m *memStore) UpdateWebhookDeliveryAttempt(_ context.Context, id int64, statusCode int, success bool, attempts int, nextAttemptAt *time.Time, state, lastError string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, ok := m.byID[id]
	if !ok {
		return nil
	}
	d.Attempts = attempts
	d.Success = success
	d.State = state
	d.LastError = lastError
	if statusCode > 0 {
		d.StatusCode.Valid = true
		d.StatusCode.Int64 = int64(statusCode)
	}
	if nextAttemptAt != nil {
		d.NextAttemptAt = nextAttemptAt.UTC().Format(time.RFC3339)
	} else {
		d.NextAttemptAt = ""
	}
	m.byID[id] = d
	return nil
}
