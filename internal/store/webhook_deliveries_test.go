package store_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jeb-maker/revues/internal/store"
)

func TestWebhookDeliveryQueue_EnqueueListUpdate(t *testing.T) {
	ctx := context.Background()
	st, _ := testStore(t)
	ctx = defaultOrgCtx(ctx, st)

	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	id, err := st.EnqueueWebhookDelivery(ctx, "evt-1", "webhook.test", "https://example.com/hook", []byte(`{"ok":true}`), now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatalf("EnqueueWebhookDelivery: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero id")
	}

	due, err := st.ListDueWebhookDeliveries(ctx, now, 10)
	if err != nil {
		t.Fatalf("ListDueWebhookDeliveries: %v", err)
	}
	if len(due) != 1 || due[0].ID != id || due[0].State != store.WebhookDeliveryPending {
		t.Fatalf("due = %+v", due)
	}
	if due[0].OrganizationID <= 0 {
		t.Fatalf("OrganizationID = %d, want > 0", due[0].OrganizationID)
	}

	future := now.Add(time.Minute)
	if err = st.UpdateWebhookDeliveryAttempt(ctx, id, 503, false, 1, &future, store.WebhookDeliveryPending, "unexpected status 503"); err != nil {
		t.Fatalf("UpdateWebhookDeliveryAttempt: %v", err)
	}

	due, err = st.ListDueWebhookDeliveries(ctx, now, 10)
	if err != nil {
		t.Fatalf("ListDueWebhookDeliveries(after): %v", err)
	}
	if len(due) != 0 {
		t.Fatalf("expected not due yet, got %+v", due)
	}

	due, err = st.ListDueWebhookDeliveries(ctx, future, 10)
	if err != nil {
		t.Fatalf("ListDueWebhookDeliveries(future): %v", err)
	}
	if len(due) != 1 || due[0].Attempts != 1 {
		t.Fatalf("due future = %+v", due)
	}

	if err = st.UpdateWebhookDeliveryAttempt(ctx, id, 200, true, 2, nil, store.WebhookDeliveryDone, ""); err != nil {
		t.Fatalf("mark done: %v", err)
	}
	due, err = st.ListDueWebhookDeliveries(ctx, future.Add(time.Hour), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(due) != 0 {
		t.Fatalf("done row still listed: %+v", due)
	}
}

func TestWebhookDeliveryQueue_OrgListAndRetry(t *testing.T) {
	ctx := context.Background()
	st, _ := testStore(t)
	ctx = defaultOrgCtx(ctx, st)

	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	id, err := st.EnqueueWebhookDelivery(ctx, "evt-list", "webhook.test", "https://example.com/hook", []byte(`{}`), now, now.Add(time.Hour))
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if err := st.UpdateWebhookDeliveryAttempt(ctx, id, 500, false, 5, nil, store.WebhookDeliveryPoison, "max"); err != nil {
		t.Fatalf("poison: %v", err)
	}

	list, err := st.ListOrgWebhookDeliveries(ctx, 10)
	if err != nil {
		t.Fatalf("ListOrgWebhookDeliveries: %v", err)
	}
	if len(list) != 1 || list[0].ID != id || list[0].State != store.WebhookDeliveryPoison {
		t.Fatalf("list = %+v", list)
	}
	if len(list[0].Payload) != 0 {
		t.Fatalf("list must omit payload, got %q", list[0].Payload)
	}

	got, err := st.WebhookDeliveryByID(ctx, id)
	if err != nil {
		t.Fatalf("WebhookDeliveryByID: %v", err)
	}
	if got.State != store.WebhookDeliveryPoison {
		t.Fatalf("got = %+v", got)
	}

	next := now.Add(time.Minute)
	exp := now.Add(24 * time.Hour)
	if err := st.ResetWebhookDeliveryForRetry(ctx, id, next, exp); err != nil {
		t.Fatalf("ResetWebhookDeliveryForRetry: %v", err)
	}
	got, err = st.WebhookDeliveryByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if got.State != store.WebhookDeliveryPending || got.Attempts != 0 || got.NextAttemptAt != next.UTC().Format(time.RFC3339) {
		t.Fatalf("after reset = %+v", got)
	}

	_, err = st.WebhookDeliveryByID(ctx, 99999)
	if !errors.Is(err, store.ErrWebhookDeliveryNotFound) {
		t.Fatalf("missing id err = %v", err)
	}
}

func TestWebhookDeliveryQueue_RequiresOrg(t *testing.T) {
	st, _ := testStore(t)
	now := time.Now().UTC()
	_, err := st.EnqueueWebhookDelivery(context.Background(), "e", "webhook.test", "https://example.com", nil, now, now.Add(time.Hour))
	if err == nil {
		t.Fatal("expected organization context required")
	}
}

func TestWebhookDeliveryQueue_MigrationColumns(t *testing.T) {
	ctx := context.Background()
	_, db := testStore(t)
	rows, err := db.QueryContext(ctx, `PRAGMA table_info(webhook_deliveries)`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	cols := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt any
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatal(err)
		}
		cols[name] = true
	}
	for _, want := range []string{"payload", "attempts", "next_attempt_at", "expires_at", "state", "last_error", "organization_id"} {
		if !cols[want] {
			t.Fatalf("missing column %q", want)
		}
	}
}
