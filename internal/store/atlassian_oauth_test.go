package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

func TestAtlassianOAuthTokensUpsertGetDelete(t *testing.T) {
	ctx := context.Background()
	st, _ := testStore(t)

	user, err := st.UpsertGitHubUser(ctx, 42, "alice", "alice@example.com", "Alice", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	row := store.AtlassianOAuthTokens{
		UserID:                user.ID,
		CloudID:               "cloud-1",
		SiteURL:               "https://example.atlassian.net",
		AccountEmail:          "alice@atlassian.example",
		AccessTokenEncrypted:  []byte("access-enc"),
		RefreshTokenEncrypted: []byte("refresh-enc"),
		ExpiresAt:             "2099-01-01T00:00:00Z",
		Scopes:                "read:jira-work write:jira-work",
	}
	if err := st.UpsertAtlassianOAuthTokens(ctx, row); err != nil {
		t.Fatalf("UpsertAtlassianOAuthTokens(): %v", err)
	}

	got, err := st.GetAtlassianOAuthTokensByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetAtlassianOAuthTokensByUserID(): %v", err)
	}
	if got.CloudID != row.CloudID || got.SiteURL != row.SiteURL || got.AccountEmail != row.AccountEmail {
		t.Fatalf("got = %+v", got)
	}
	if string(got.AccessTokenEncrypted) != "access-enc" || string(got.RefreshTokenEncrypted) != "refresh-enc" {
		t.Fatalf("tokens = %q / %q", got.AccessTokenEncrypted, got.RefreshTokenEncrypted)
	}
	if got.CreatedAt == "" || got.UpdatedAt == "" {
		t.Fatal("expected timestamps")
	}

	row.CloudID = "cloud-2"
	row.AccessTokenEncrypted = []byte("access-2")
	if err := st.UpsertAtlassianOAuthTokens(ctx, row); err != nil {
		t.Fatalf("UpsertAtlassianOAuthTokens(update): %v", err)
	}
	got, err = st.GetAtlassianOAuthTokensByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Get after update: %v", err)
	}
	if got.CloudID != "cloud-2" || string(got.AccessTokenEncrypted) != "access-2" {
		t.Fatalf("update = %+v", got)
	}

	if err := st.DeleteAtlassianOAuthTokensByUserID(ctx, user.ID); err != nil {
		t.Fatalf("DeleteAtlassianOAuthTokensByUserID(): %v", err)
	}
	if _, err := st.GetAtlassianOAuthTokensByUserID(ctx, user.ID); !errors.Is(err, store.ErrAtlassianOAuthNotFound) {
		t.Fatalf("Get after delete = %v", err)
	}
	if err := st.DeleteAtlassianOAuthTokensByUserID(ctx, user.ID); !errors.Is(err, store.ErrAtlassianOAuthNotFound) {
		t.Fatalf("Delete again = %v", err)
	}
}
