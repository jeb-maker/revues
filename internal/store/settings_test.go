package store_test

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/jeb-maker/revues/internal/store"
)

func testStore(t *testing.T) (*store.Store, *sql.DB) {
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

	return store.New(db), db
}

func TestSettingsUpsertGetDelete(t *testing.T) {
	ctx := context.Background()
	st, _ := testStore(t)

	value := []byte("encrypted-payload")
	if err := st.UpsertSetting(ctx, "smtp", value); err != nil {
		t.Fatalf("UpsertSetting(): %v", err)
	}

	got, err := st.GetSetting(ctx, "smtp")
	if err != nil {
		t.Fatalf("GetSetting(): %v", err)
	}
	if string(got) != string(value) {
		t.Fatalf("GetSetting() = %q, want %q", got, value)
	}

	updated := []byte("updated-payload")
	if err := st.UpsertSetting(ctx, "smtp", updated); err != nil {
		t.Fatalf("UpsertSetting(update): %v", err)
	}

	got, err = st.GetSetting(ctx, "smtp")
	if err != nil {
		t.Fatalf("GetSetting(update): %v", err)
	}
	if string(got) != string(updated) {
		t.Fatalf("GetSetting(update) = %q, want %q", got, updated)
	}

	if err := st.DeleteSetting(ctx, "smtp"); err != nil {
		t.Fatalf("DeleteSetting(): %v", err)
	}
	if _, err := st.GetSetting(ctx, "smtp"); err == nil {
		t.Fatal("expected missing setting after delete")
	}
}
