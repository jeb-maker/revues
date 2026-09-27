package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/jeb-maker/revues/internal/store"
)

func TestEmailDeliveryEnqueueAndDrainList(t *testing.T) {
	db := openMemoryDB(t)
	st := store.New(db)
	ctx := context.Background()

	now := time.Now().UTC()
	id, err := st.EnqueueEmailDelivery(ctx, 0, "a@example.com", "Sujet", "Corps", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatalf("EnqueueEmailDelivery: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero id")
	}

	due, err := st.ListDueEmailDeliveries(ctx, now.Add(time.Second), 10)
	if err != nil {
		t.Fatalf("ListDueEmailDeliveries: %v", err)
	}
	if len(due) != 1 || due[0].ToAddress != "a@example.com" {
		t.Fatalf("due = %#v", due)
	}

	if updErr := st.UpdateEmailDeliveryAttempt(ctx, id, 1, nil, store.EmailDeliveryDone, ""); updErr != nil {
		t.Fatalf("UpdateEmailDeliveryAttempt: %v", updErr)
	}
	due, err = st.ListDueEmailDeliveries(ctx, now.Add(time.Second), 10)
	if err != nil {
		t.Fatalf("ListDueEmailDeliveries after done: %v", err)
	}
	if len(due) != 0 {
		t.Fatalf("expected no due after done, got %#v", due)
	}
}
