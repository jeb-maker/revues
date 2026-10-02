package apiv1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
	appweb "github.com/jeb-maker/revues/internal/web"
)

func TestAuthAPI_BootstrapGuestCSRF(t *testing.T) {
	t.Parallel()

	handler := newTestRouter(t, config.Config{
		SessionSecret:         "test-secret-at-least-thirty-two-bytes",
		LoginRequireWhitelist: false,
		BootstrapAdminEmail:   "admin@example.com",
		Env:                   "development",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bootstrap", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	if body["authenticated"] != false {
		t.Fatalf("authenticated = %v", body["authenticated"])
	}
	csrf, _ := body["csrf_token"].(string)
	if csrf == "" {
		t.Fatal("expected csrf_token")
	}
	foundGuest := false
	for _, c := range resultCookies(t, rec) {
		if c.Name == "revues_guest" && c.Value != "" {
			foundGuest = true
		}
	}
	if !foundGuest {
		t.Fatal("expected revues_guest cookie")
	}
}

func TestAuthAPI_LoginRequiresCSRF(t *testing.T) {
	t.Parallel()

	handler := newTestRouter(t, config.Config{
		SessionSecret:       "test-secret-at-least-thirty-two-bytes",
		BootstrapAdminEmail: "admin@example.com",
		Env:                 "development",
	})

	payload := `{"email":"user@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 CSRF reject; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "csrf") {
		t.Fatalf("body = %q, want csrf error", rec.Body.String())
	}
}

func TestAuthAPI_MeRequiresSession(t *testing.T) {
	t.Parallel()

	handler := newTestRouter(t, config.Config{
		SessionSecret: "test-secret-at-least-thirty-two-bytes",
		Env:           "development",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401; body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthAPI_RegisterWhitelistReject(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		SessionSecret:         "test-secret-at-least-thirty-two-bytes",
		LoginRequireWhitelist: true,
		BootstrapAdminEmail:   "admin@example.com",
		Env:                   "development",
	}
	handler, st := newTestRouterWithStore(t, cfg)

	// Seed an org + whitelist unrelated email so strict mode is meaningful.
	ctx := context.Background()
	orgCtx := testutil.DefaultOrgContext(ctx, st)
	if err := st.InsertAllowedEmail(orgCtx, "allowed@example.com", auth.RoleEditor); err != nil {
		t.Fatalf("InsertAllowedEmail: %v", err)
	}
	_ = orgCtx

	guest, csrf := bootstrapGuest(t, handler)

	payload := map[string]string{
		"email":            "stranger@example.com",
		"display_name":     "Stranger",
		"password":         "password123",
		"password_confirm": "password123",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", csrf)
	req.AddCookie(guest)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "registration_failed") &&
		!strings.Contains(rec.Body.String(), "Impossible de créer") {
		t.Fatalf("unexpected body %q", rec.Body.String())
	}
}

func TestAuthAPI_RegisterAndMe(t *testing.T) {
	t.Parallel()

	handler := newTestRouter(t, config.Config{
		SessionSecret:         "test-secret-at-least-thirty-two-bytes",
		LoginRequireWhitelist: false,
		BootstrapAdminEmail:   "admin@example.com",
		Env:                   "development",
	})

	guest, csrf := bootstrapGuest(t, handler)
	payload := map[string]string{
		"email":            "newbie@example.com",
		"display_name":     "Newbie",
		"password":         "password123",
		"password_confirm": "password123",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", csrf)
	req.AddCookie(guest)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("register status = %d body=%s", rec.Code, rec.Body.String())
	}

	var sessionCookie *http.Cookie
	for _, c := range resultCookies(t, rec) {
		if c.Name == "revues_session" && c.Value != "" {
			sessionCookie = c
		}
	}
	if sessionCookie == nil {
		t.Fatal("expected session cookie after register")
	}

	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	meReq.AddCookie(sessionCookie)
	meRec := httptest.NewRecorder()
	handler.ServeHTTP(meRec, meReq)
	if meRec.Code != http.StatusOK {
		t.Fatalf("me status = %d body=%s", meRec.Code, meRec.Body.String())
	}
	if !strings.Contains(meRec.Body.String(), "newbie@example.com") {
		t.Fatalf("me body = %s", meRec.Body.String())
	}
}

func TestAuthAPI_LogoutRequiresCSRF(t *testing.T) {
	t.Parallel()

	handler := newTestRouter(t, config.Config{
		SessionSecret:         "test-secret-at-least-thirty-two-bytes",
		LoginRequireWhitelist: false,
		BootstrapAdminEmail:   "admin@example.com",
		Env:                   "development",
	})

	guest, csrf := bootstrapGuest(t, handler)
	payload := map[string]string{
		"email":            "logout@example.com",
		"display_name":     "Logout",
		"password":         "password123",
		"password_confirm": "password123",
	}
	body, _ := json.Marshal(payload)
	reg := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	reg.Header.Set("Content-Type", "application/json")
	reg.Header.Set("X-CSRF-Token", csrf)
	reg.AddCookie(guest)
	regRec := httptest.NewRecorder()
	handler.ServeHTTP(regRec, reg)
	if regRec.Code != http.StatusCreated {
		t.Fatalf("register: %d %s", regRec.Code, regRec.Body.String())
	}

	var sessionCookie *http.Cookie
	var sessionCSRF string
	var authOK map[string]any
	_ = json.Unmarshal(regRec.Body.Bytes(), &authOK)
	sessionCSRF, _ = authOK["csrf_token"].(string)
	for _, c := range resultCookies(t, regRec) {
		if c.Name == "revues_session" {
			sessionCookie = c
		}
	}

	// Missing CSRF → 403
	bad := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	bad.AddCookie(sessionCookie)
	badRec := httptest.NewRecorder()
	handler.ServeHTTP(badRec, bad)
	if badRec.Code != http.StatusForbidden {
		t.Fatalf("logout without csrf status = %d", badRec.Code)
	}

	okReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	okReq.AddCookie(sessionCookie)
	okReq.Header.Set("X-CSRF-Token", sessionCSRF)
	okRec := httptest.NewRecorder()
	handler.ServeHTTP(okRec, okReq)
	if okRec.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d body=%s", okRec.Code, okRec.Body.String())
	}
}

func newTestRouter(t *testing.T, cfg config.Config) http.Handler {
	t.Helper()
	h, _ := newTestRouterWithStore(t, cfg)
	return h
}

func newTestRouterWithStore(t *testing.T, cfg config.Config) (http.Handler, *store.Store) {
	t.Helper()

	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/test.db", 0)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err = store.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	st := store.New(db)
	_ = testutil.DefaultOrgContext(ctx, st)

	handler, _, _, err := appweb.NewRouter(appweb.Deps{Config: cfg, DB: db})
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}
	return handler, st
}

func bootstrapGuest(t *testing.T, handler http.Handler) (*http.Cookie, string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/bootstrap", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("bootstrap status = %d", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	csrf, _ := body["csrf_token"].(string)
	var guest *http.Cookie
	for _, c := range resultCookies(t, rec) {
		if c.Name == "revues_guest" {
			guest = c
		}
	}
	if guest == nil || csrf == "" {
		t.Fatal("missing guest cookie or csrf")
	}
	return guest, csrf
}

// resultCookies returns recorder cookies and closes the synthetic response body (bodyclose).
func resultCookies(t *testing.T, rec *httptest.ResponseRecorder) []*http.Cookie {
	t.Helper()
	res := rec.Result()
	t.Cleanup(func() { _ = res.Body.Close() })
	return res.Cookies()
}
