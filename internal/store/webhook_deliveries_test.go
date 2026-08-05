package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/jeb-maker/revues/internal/store"
)

func TestWebhookDeliveryQueue_EnqueueListUpdate(t *testing.T) {
	ctx := context.Background()
	st, _ := testStore(t)

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

	future := now.Add(time.Minute)
	if err := st.UpdateWebhookDeliveryAttempt(ctx, id, 503, false, 1, &future, store.WebhookDeliveryPending, "unexpected status 503"); err != nil {
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

	if err := st.UpdateWebhookDeliveryAttempt(ctx, id, 200, true, 2, nil, store.WebhookDeliveryDone, ""); err != nil {
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
	for _, want := range []string{"payload", "attempts", "next_attempt_at", "expires_at", "state", "last_error"} {
		if !cols[want] {
			t.Fatalf("missing column %q", want)
		}
	}
}
