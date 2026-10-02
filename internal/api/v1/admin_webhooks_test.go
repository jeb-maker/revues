package apiv1_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/integrations/webhooks"
	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
)

func TestAdminWebhooks_ConfigHMACAndSSRF(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		SessionSecret:  "test-secret-at-least-thirty-two-bytes",
		EncryptionKey:  config.TestEncryptionKey(),
		Env:            "development",
		AttachmentsDir: t.TempDir(),
	}
	handler, st := newTestRouterWithStore(t, cfg)

	admin, session, csrf := seedSessionUser(t, st, cfg, "wh-admin@example.com", "WH Admin", auth.RoleAdmin, true)
	_ = admin
	member, memberSession, memberCSRF := seedSessionUser(t, st, cfg, "wh-member@example.com", "Member", auth.RoleEditor, true)
	_ = member

	denied := doJSON(t, handler, http.MethodGet, "/api/v1/admin/webhooks", nil, memberSession, "")
	if denied.Code != http.StatusForbidden {
		t.Fatalf("member get status = %d body=%s", denied.Code, denied.Body.String())
	}

	// Anti-SSRF on save (private IP blocked even in development for https literal IPs)
	ssrf := doJSON(t, handler, http.MethodPut, "/api/v1/admin/webhooks", map[string]any{
		"urls": []string{"https://192.168.1.10/hook"}, "secret": "s3cret",
		"review_completed": true, "review_item_nok": false,
	}, session, csrf)
	if ssrf.Code != http.StatusBadRequest {
		t.Fatalf("ssrf put status = %d body=%s", ssrf.Code, ssrf.Body.String())
	}

	var gotSig string
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSig = r.Header.Get("X-Revues-Signature")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	secret := "super-hmac-secret-key"
	put := doJSON(t, handler, http.MethodPut, "/api/v1/admin/webhooks", map[string]any{
		"urls": []string{srv.URL}, "secret": secret,
		"review_completed": true, "review_item_nok": true,
	}, session, csrf)
	if put.Code != http.StatusOK {
		t.Fatalf("put status = %d body=%s", put.Code, put.Body.String())
	}
	body := put.Body.String()
	if strings.Contains(body, secret) {
		t.Fatalf("secret leaked: %s", body)
	}
	var settingsResp map[string]any
	if err := json.Unmarshal(put.Body.Bytes(), &settingsResp); err != nil {
		t.Fatalf("json: %v", err)
	}
	if settingsResp["has_secret"] != true || settingsResp["configured"] != true || settingsResp["enabled"] != true {
		t.Fatalf("settings = %#v", settingsResp)
	}
	if _, ok := settingsResp["secret"]; ok {
		t.Fatal("secret field must be absent")
	}

	// Blank secret keeps existing
	putKeep := doJSON(t, handler, http.MethodPut, "/api/v1/admin/webhooks", map[string]any{
		"urls": []string{srv.URL}, "secret": "",
		"review_completed": true, "review_item_nok": false,
	}, session, csrf)
	if putKeep.Code != http.StatusOK {
		t.Fatalf("put keep status = %d body=%s", putKeep.Code, putKeep.Body.String())
	}
	var keep map[string]any
	_ = json.Unmarshal(putKeep.Body.Bytes(), &keep)
	if keep["has_secret"] != true {
		t.Fatal("expected secret retained")
	}

	noCSRF := doJSON(t, handler, http.MethodPost, "/api/v1/admin/webhooks/test", nil, session, "")
	if noCSRF.Code != http.StatusForbidden {
		t.Fatalf("no csrf test status = %d", noCSRF.Code)
	}

	test := doJSON(t, handler, http.MethodPost, "/api/v1/admin/webhooks/test", nil, session, csrf)
	if test.Code != http.StatusNoContent {
		t.Fatalf("test status = %d body=%s", test.Code, test.Body.String())
	}
	if !webhooks.VerifySignature(secret, gotBody, gotSig) {
		t.Fatalf("HMAC mismatch sig=%q body=%s", gotSig, gotBody)
	}
	var env webhooks.Envelope
	if err := json.Unmarshal(gotBody, &env); err != nil || env.EventType != webhooks.EventTest {
		t.Fatalf("payload = %s err=%v", gotBody, err)
	}

	list := doJSON(t, handler, http.MethodGet, "/api/v1/admin/webhooks/deliveries", nil, session, "")
	if list.Code != http.StatusOK {
		t.Fatalf("list status = %d body=%s", list.Code, list.Body.String())
	}
	var deliveries map[string]any
	_ = json.Unmarshal(list.Body.Bytes(), &deliveries)
	items, _ := deliveries["deliveries"].([]any)
	if len(items) < 1 {
		t.Fatalf("expected deliveries, got %#v", deliveries)
	}
	first := items[0].(map[string]any)
	if first["state"] != "done" {
		t.Fatalf("delivery state = %#v", first)
	}
	if _, ok := first["payload"]; ok {
		t.Fatal("payload must not be exposed")
	}

	deniedTest := doJSON(t, handler, http.MethodPost, "/api/v1/admin/webhooks/test", nil, memberSession, memberCSRF)
	if deniedTest.Code != http.StatusForbidden {
		t.Fatalf("member test status = %d", deniedTest.Code)
	}

	hub := doJSON(t, handler, http.MethodGet, "/api/v1/admin/integrations", nil, session, "")
	if hub.Code != http.StatusOK {
		t.Fatalf("hub status = %d", hub.Code)
	}
	var overview map[string]any
	_ = json.Unmarshal(hub.Body.Bytes(), &overview)
	found := false
	for _, raw := range overview["items"].([]any) {
		item := raw.(map[string]any)
		if item["key"] == "webhooks" {
			found = true
			if item["enabled"] != true {
				t.Fatalf("webhooks enabled = %v", item["enabled"])
			}
			if item["config_path"] != "/admin/settings/webhooks" {
				t.Fatalf("config_path = %v", item["config_path"])
			}
		}
	}
	if !found {
		t.Fatal("webhooks missing from hub")
	}

	del := doJSON(t, handler, http.MethodDelete, "/api/v1/admin/webhooks", nil, session, csrf)
	if del.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d", del.Code)
	}
	after := doJSON(t, handler, http.MethodGet, "/api/v1/admin/webhooks", nil, session, "")
	var cleared map[string]any
	_ = json.Unmarshal(after.Body.Bytes(), &cleared)
	if cleared["configured"] != false || cleared["enabled"] != false {
		t.Fatalf("after clear = %#v", cleared)
	}
}

func TestAdminWebhooks_DrainAndRetry(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		SessionSecret:  "test-secret-at-least-thirty-two-bytes",
		EncryptionKey:  config.TestEncryptionKey(),
		Env:            "development",
		AttachmentsDir: t.TempDir(),
	}
	handler, st := newTestRouterWithStore(t, cfg)
	_, session, csrf := seedSessionUser(t, st, cfg, "wh-drain@example.com", "Drain Admin", auth.RoleAdmin, true)

	failOnce := true
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if failOnce {
			failOnce = false
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	put := doJSON(t, handler, http.MethodPut, "/api/v1/admin/webhooks", map[string]any{
		"urls": []string{srv.URL}, "secret": "drain-secret",
		"review_completed": true, "review_item_nok": false,
	}, session, csrf)
	if put.Code != http.StatusOK {
		t.Fatalf("put status = %d body=%s", put.Code, put.Body.String())
	}

	test := doJSON(t, handler, http.MethodPost, "/api/v1/admin/webhooks/test", nil, session, csrf)
	if test.Code != http.StatusBadRequest {
		t.Fatalf("expected first test failure, got %d body=%s", test.Code, test.Body.String())
	}

	list := doJSON(t, handler, http.MethodGet, "/api/v1/admin/webhooks/deliveries", nil, session, "")
	var payload map[string]any
	_ = json.Unmarshal(list.Body.Bytes(), &payload)
	items := payload["deliveries"].([]any)
	del := items[0].(map[string]any)
	if del["state"] != "pending" {
		t.Fatalf("want pending, got %#v", del)
	}
	id := int64(del["id"].(float64))

	// Force next attempt due
	orgID := mustDefaultOrgID(t, st)
	ctx := orgctx.WithOrganizationID(context.Background(), orgID)
	now := time.Now().UTC().Add(-time.Minute)
	if err := st.UpdateWebhookDeliveryAttempt(ctx, id, 503, false, 1, &now, store.WebhookDeliveryPending, "forced"); err != nil {
		t.Fatalf("force due: %v", err)
	}

	drain := doJSON(t, handler, http.MethodPost, "/api/v1/admin/webhooks/deliveries/drain", nil, session, csrf)
	if drain.Code != http.StatusNoContent {
		t.Fatalf("drain status = %d body=%s", drain.Code, drain.Body.String())
	}

	list2 := doJSON(t, handler, http.MethodGet, "/api/v1/admin/webhooks/deliveries", nil, session, "")
	_ = json.Unmarshal(list2.Body.Bytes(), &payload)
	del = payload["deliveries"].([]any)[0].(map[string]any)
	if del["state"] != "done" {
		t.Fatalf("after drain want done, got %#v", del)
	}

	// Seed a poison row and retry
	poisonID, err := st.EnqueueWebhookDelivery(ctx, "poison-evt", webhooks.EventTest, srv.URL, []byte(`{"x":1}`), now, now.Add(webhooks.DeliveryTTL))
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	err = st.UpdateWebhookDeliveryAttempt(ctx, poisonID, 500, false, webhooks.MaxAttempts, nil, store.WebhookDeliveryPoison, "max")
	if err != nil {
		t.Fatalf("poison: %v", err)
	}

	retry := doJSON(t, handler, http.MethodPost,
		"/api/v1/admin/webhooks/deliveries/"+strconv.FormatInt(poisonID, 10)+"/retry", nil, session, csrf)
	if retry.Code != http.StatusNoContent {
		t.Fatalf("retry status = %d body=%s", retry.Code, retry.Body.String())
	}
	got, err := st.WebhookDeliveryByID(ctx, poisonID)
	if err != nil {
		t.Fatal(err)
	}
	if got.State != store.WebhookDeliveryDone {
		t.Fatalf("after retry state = %+v", got)
	}
}
