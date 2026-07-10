package store_test

import (
	"context"
	"testing"

	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
)

func TestIntegrationUpsertGet(t *testing.T) {
	ctx := context.Background()
	st, _ := testStore(t)
	ctx = testutil.DefaultOrgContext(ctx, st)

	payload := []byte("encrypted-jira-config")
	if err := st.UpsertIntegrationByType(ctx, store.IntegrationTypeJira, true, payload); err != nil {
		t.Fatalf("UpsertIntegrationByType(): %v", err)
	}

	got, err := st.GetIntegrationByType(ctx, store.IntegrationTypeJira)
	if err != nil {
		t.Fatalf("GetIntegrationByType(): %v", err)
	}
	if got.Type != store.IntegrationTypeJira || !got.Enabled {
		t.Fatalf("GetIntegrationByType() = %+v", got)
	}
	if string(got.ConfigEncrypted) != string(payload) {
		t.Fatalf("ConfigEncrypted = %q, want %q", got.ConfigEncrypted, payload)
	}

	updated := []byte("updated-config")
	if upsertErr := st.UpsertIntegrationByType(ctx, store.IntegrationTypeJira, false, updated); upsertErr != nil {
		t.Fatalf("UpsertIntegrationByType(update): %v", upsertErr)
	}

	got, err = st.GetIntegrationByType(ctx, store.IntegrationTypeJira)
	if err != nil {
		t.Fatalf("GetIntegrationByType(update): %v", err)
	}
	if got.Enabled {
		t.Fatal("expected disabled after update")
	}
	if string(got.ConfigEncrypted) != string(updated) {
		t.Fatalf("ConfigEncrypted = %q, want %q", got.ConfigEncrypted, updated)
	}
}
