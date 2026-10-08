package atlassian_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/crypto"
	"github.com/jeb-maker/revues/internal/integrations/atlassian"
	"github.com/jeb-maker/revues/internal/store"
)

func testKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, crypto.KeySize)
	for i := range key {
		key[i] = byte(i + 1)
	}
	return key
}

func testTokenStore(t *testing.T) (*store.Store, []byte) {
	t.Helper()
	ctx := context.Background()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("foreign_keys: %v", err)
	}
	if err := store.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate(): %v", err)
	}
	return store.New(db), testKey(t)
}

func TestEnsureAccessToken_RefreshAndMissing(t *testing.T) {
	ctx := context.Background()
	st, key := testTokenStore(t)

	user, err := st.UpsertGitHubUser(ctx, 7, "bob", "bob@example.com", "Bob", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	svc := &atlassian.TokenService{Store: st, EncryptionKey: key}
	_, err = svc.EnsureAccessToken(ctx, user.ID)
	if !errors.Is(err, atlassian.ErrUserOAuthRequired) {
		t.Fatalf("missing = %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["grant_type"] != "refresh_token" || body["refresh_token"] != "refresh-old" {
			http.Error(w, "bad", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "access-new",
			"refresh_token": "refresh-new",
			"expires_in":    3600,
			"scope":         "read:jira-work write:jira-work",
		})
	}))
	t.Cleanup(srv.Close)

	accessEnc, _ := crypto.Encrypt(key, []byte("access-old"))
	refreshEnc, _ := crypto.Encrypt(key, []byte("refresh-old"))
	if upsertErr := st.UpsertAtlassianOAuthTokens(ctx, store.AtlassianOAuthTokens{
		UserID:                user.ID,
		CloudID:               "cloud-x",
		SiteURL:               "https://x.atlassian.net",
		AccountEmail:          "bob@x.example",
		AccessTokenEncrypted:  accessEnc,
		RefreshTokenEncrypted: refreshEnc,
		ExpiresAt:             time.Now().UTC().Add(-time.Minute).Format(time.RFC3339),
		Scopes:                "read:jira-work",
	}); upsertErr != nil {
		t.Fatalf("Upsert: %v", upsertErr)
	}

	svc.OAuth = &auth.AtlassianOAuth{
		ClientID:     "cid",
		ClientSecret: "sec",
		BaseURL:      "http://localhost",
		TokenURL:     srv.URL,
	}

	creds, err := svc.EnsureAccessToken(ctx, user.ID)
	if err != nil {
		t.Fatalf("EnsureAccessToken(): %v", err)
	}
	if creds.AccessToken != "access-new" || creds.CloudID != "cloud-x" {
		t.Fatalf("creds = %+v", creds)
	}

	row, err := st.GetAtlassianOAuthTokensByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	gotAccess, _ := crypto.Decrypt(key, row.AccessTokenEncrypted)
	if string(gotAccess) != "access-new" {
		t.Fatalf("stored access = %q", gotAccess)
	}
}
