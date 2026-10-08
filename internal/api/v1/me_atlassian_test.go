package apiv1_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/crypto"
	"github.com/jeb-maker/revues/internal/store"
)

func testAtlassianEncryptionKeyB64(t *testing.T) string {
	t.Helper()
	key := make([]byte, crypto.KeySize)
	for i := range key {
		key[i] = byte(i + 3)
	}
	return base64.StdEncoding.EncodeToString(key)
}

func TestMeAtlassian_StatusAndDisconnectCSRF(t *testing.T) {
	cfg := config.Config{
		SessionSecret:         "test-secret-at-least-thirty-two-bytes",
		AtlassianClientID:     "atlassian-client",
		AtlassianClientSecret: "atlassian-secret",
		EncryptionKey:         testAtlassianEncryptionKeyB64(t),
	}
	handler, st := newTestRouterWithStore(t, cfg)
	user, session, csrf := seedSessionUser(t, st, cfg, "atlassian-me@example.com", "Atlassian Me", auth.RoleEditor, true)

	// Disabled flag path when OAuth not configured is covered elsewhere; here OAuth is on.
	get := doJSON(t, handler, http.MethodGet, "/api/v1/me/atlassian", nil, session, "")
	if get.Code != http.StatusOK {
		t.Fatalf("GET me/atlassian status = %d body=%s", get.Code, get.Body.String())
	}
	var status map[string]any
	if err := json.Unmarshal(get.Body.Bytes(), &status); err != nil {
		t.Fatalf("json: %v", err)
	}
	if enabled, _ := status["enabled"].(bool); !enabled {
		t.Fatalf("enabled = %#v, want true", status["enabled"])
	}
	if connected, _ := status["connected"].(bool); connected {
		t.Fatalf("connected = %#v, want false before upsert", status["connected"])
	}

	encKey, err := cfg.EncryptionKeyBytes()
	if err != nil {
		t.Fatalf("EncryptionKeyBytes: %v", err)
	}
	accessEnc, err := crypto.Encrypt(encKey, []byte("access-token"))
	if err != nil {
		t.Fatalf("Encrypt access: %v", err)
	}
	refreshEnc, err := crypto.Encrypt(encKey, []byte("refresh-token"))
	if err != nil {
		t.Fatalf("Encrypt refresh: %v", err)
	}
	if upsertErr := st.UpsertAtlassianOAuthTokens(context.Background(), store.AtlassianOAuthTokens{
		UserID:                user.ID,
		CloudID:               "cloud-1",
		SiteURL:               "https://example.atlassian.net",
		AccountEmail:          "alice@example.atlassian.net",
		AccessTokenEncrypted:  accessEnc,
		RefreshTokenEncrypted: refreshEnc,
		ExpiresAt:             time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
		Scopes:                "read:jira-work write:jira-work offline_access",
	}); upsertErr != nil {
		t.Fatalf("UpsertAtlassianOAuthTokens: %v", upsertErr)
	}

	getConn := doJSON(t, handler, http.MethodGet, "/api/v1/me/atlassian", nil, session, "")
	if getConn.Code != http.StatusOK {
		t.Fatalf("GET connected status = %d body=%s", getConn.Code, getConn.Body.String())
	}
	var connectedBody map[string]any
	if err := json.Unmarshal(getConn.Body.Bytes(), &connectedBody); err != nil {
		t.Fatalf("json connected: %v", err)
	}
	if connected, _ := connectedBody["connected"].(bool); !connected {
		t.Fatalf("connected = %#v, want true", connectedBody["connected"])
	}
	if site, _ := connectedBody["site_url"].(string); site != "https://example.atlassian.net" {
		t.Fatalf("site_url = %q", site)
	}

	noCSRF := doJSON(t, handler, http.MethodDelete, "/api/v1/me/atlassian", nil, session, "")
	if noCSRF.Code != http.StatusForbidden {
		t.Fatalf("DELETE without CSRF status = %d, want 403", noCSRF.Code)
	}

	del := doJSON(t, handler, http.MethodDelete, "/api/v1/me/atlassian", nil, session, csrf)
	if del.Code != http.StatusNoContent {
		t.Fatalf("DELETE me/atlassian status = %d body=%s", del.Code, del.Body.String())
	}
	if _, getErr := st.GetAtlassianOAuthTokensByUserID(context.Background(), user.ID); !errors.Is(getErr, store.ErrAtlassianOAuthNotFound) {
		t.Fatalf("tokens after DELETE = %v, want ErrAtlassianOAuthNotFound", getErr)
	}
}

func TestAtlassianDisconnectForm_RequiresCSRF(t *testing.T) {
	cfg := config.Config{
		SessionSecret:         "test-secret-at-least-thirty-two-bytes",
		AtlassianClientID:     "atlassian-client",
		AtlassianClientSecret: "atlassian-secret",
		EncryptionKey:         testAtlassianEncryptionKeyB64(t),
	}
	handler, st := newTestRouterWithStore(t, cfg)
	user, session, csrf := seedSessionUser(t, st, cfg, "atlassian-disc@example.com", "Atlassian Disc", auth.RoleEditor, true)

	encKey, err := cfg.EncryptionKeyBytes()
	if err != nil {
		t.Fatalf("EncryptionKeyBytes: %v", err)
	}
	accessEnc, err := crypto.Encrypt(encKey, []byte("access-token"))
	if err != nil {
		t.Fatalf("Encrypt access: %v", err)
	}
	refreshEnc, err := crypto.Encrypt(encKey, []byte("refresh-token"))
	if err != nil {
		t.Fatalf("Encrypt refresh: %v", err)
	}
	if upsertErr := st.UpsertAtlassianOAuthTokens(context.Background(), store.AtlassianOAuthTokens{
		UserID:                user.ID,
		CloudID:               "cloud-2",
		SiteURL:               "https://example.atlassian.net",
		AccountEmail:          "bob@example.atlassian.net",
		AccessTokenEncrypted:  accessEnc,
		RefreshTokenEncrypted: refreshEnc,
		ExpiresAt:             time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
		Scopes:                "read:jira-work",
	}); upsertErr != nil {
		t.Fatalf("UpsertAtlassianOAuthTokens: %v", upsertErr)
	}

	noCSRF := httptest.NewRequest(http.MethodPost, "/auth/atlassian/disconnect", strings.NewReader(""))
	noCSRF.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	noCSRF.AddCookie(session)
	noCSRFRec := httptest.NewRecorder()
	handler.ServeHTTP(noCSRFRec, noCSRF)
	if noCSRFRec.Code != http.StatusForbidden {
		t.Fatalf("POST disconnect without CSRF status = %d, want 403", noCSRFRec.Code)
	}

	form := url.Values{"csrf_token": {csrf}}
	okReq := httptest.NewRequest(http.MethodPost, "/auth/atlassian/disconnect", strings.NewReader(form.Encode()))
	okReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	okReq.AddCookie(session)
	okRec := httptest.NewRecorder()
	handler.ServeHTTP(okRec, okReq)
	if okRec.Code != http.StatusSeeOther {
		t.Fatalf("POST disconnect status = %d body=%s", okRec.Code, okRec.Body.String())
	}
	if loc := okRec.Header().Get("Location"); !strings.Contains(loc, "atlassian=disconnected") {
		t.Fatalf("Location = %q, want disconnected", loc)
	}
	if _, getErr := st.GetAtlassianOAuthTokensByUserID(context.Background(), user.ID); !errors.Is(getErr, store.ErrAtlassianOAuthNotFound) {
		t.Fatalf("tokens after form disconnect = %v", getErr)
	}
}

func TestAtlassianStart_RequiresSession(t *testing.T) {
	cfg := config.Config{
		SessionSecret:         "test-secret-at-least-thirty-two-bytes",
		AtlassianClientID:     "atlassian-client",
		AtlassianClientSecret: "atlassian-secret",
		BaseURL:               "https://revues.example",
	}
	handler, st := newTestRouterWithStore(t, cfg)

	anon := httptest.NewRequest(http.MethodGet, "/auth/atlassian/start", nil)
	anonRec := httptest.NewRecorder()
	handler.ServeHTTP(anonRec, anon)
	if anonRec.Code != http.StatusFound && anonRec.Code != http.StatusSeeOther {
		t.Fatalf("anon start status = %d", anonRec.Code)
	}
	if loc := anonRec.Header().Get("Location"); !strings.Contains(loc, "/login") {
		t.Fatalf("anon Location = %q, want login", loc)
	}

	_, session, _ := seedSessionUser(t, st, cfg, "atlassian-start@example.com", "Start", auth.RoleEditor, true)
	start := httptest.NewRequest(http.MethodGet, "/auth/atlassian/start", nil)
	start.AddCookie(session)
	startRec := httptest.NewRecorder()
	handler.ServeHTTP(startRec, start)
	if startRec.Code != http.StatusFound {
		t.Fatalf("start status = %d body=%s", startRec.Code, startRec.Body.String())
	}
	loc := startRec.Header().Get("Location")
	if !strings.Contains(loc, "https://auth.atlassian.com/authorize") {
		t.Fatalf("start Location = %q, want Atlassian authorize", loc)
	}
	if !strings.Contains(loc, "offline_access") {
		t.Fatalf("start Location missing offline_access: %s", loc)
	}
}
