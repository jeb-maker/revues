package apiv1_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/jeb-maker/revues/internal/config"
)

func TestTemplatesAPI_CRUDAndVersionImmutability(t *testing.T) {
	t.Parallel()

	const email = "tpl-admin@example.com"
	handler := newTestRouter(t, config.Config{
		SessionSecret:         "test-secret-at-least-thirty-two-bytes",
		LoginRequireWhitelist: false,
		BootstrapAdminEmail:   email,
		Env:                   "development",
	})

	session, csrf := registerAndSession(t, handler, email, "Tpl Admin")

	createBody := map[string]any{
		"name":    "Contrôle véhicule",
		"domains": []string{"auto"},
		"items": []map[string]any{
			{"section": "Général", "label": "Freins", "required": true},
			{"section": "Général", "label": "Pneus", "help_text": "Usure"},
		},
	}
	rec := doJSON(t, handler, http.MethodPost, "/api/v1/templates", createBody, session, csrf)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", rec.Code, rec.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("json: %v", err)
	}
	id := int64(created["id"].(float64))
	version := created["version"].(map[string]any)
	if int(version["version"].(float64)) != 1 {
		t.Fatalf("version = %v, want 1", version["version"])
	}
	items := created["items"].([]any)
	if len(items) != 2 {
		t.Fatalf("items len = %d", len(items))
	}

	// List
	listRec := doJSON(t, handler, http.MethodGet, "/api/v1/templates", nil, session, "")
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d", listRec.Code)
	}

	// Save → new version (immutability)
	saveBody := map[string]any{
		"name":    "Contrôle véhicule",
		"domains": []string{"auto"},
		"items": []map[string]any{
			{"section": "Sécurité", "label": "Ceinture", "required": true},
		},
	}
	saveRec := doJSON(t, handler, http.MethodPut, "/api/v1/templates/"+strconv.FormatInt(id, 10), saveBody, session, csrf)
	if saveRec.Code != http.StatusOK {
		t.Fatalf("save status = %d body=%s", saveRec.Code, saveRec.Body.String())
	}
	var saved map[string]any
	_ = json.Unmarshal(saveRec.Body.Bytes(), &saved)
	savedVer := saved["version"].(map[string]any)
	if int(savedVer["version"].(float64)) != 2 {
		t.Fatalf("saved version = %v, want 2", savedVer["version"])
	}

	// Old version snapshot intact
	v1Rec := doJSON(t, handler, http.MethodGet, "/api/v1/templates/"+strconv.FormatInt(id, 10)+"/versions/1", nil, session, "")
	if v1Rec.Code != http.StatusOK {
		t.Fatalf("get v1 status = %d body=%s", v1Rec.Code, v1Rec.Body.String())
	}
	var v1 map[string]any
	_ = json.Unmarshal(v1Rec.Body.Bytes(), &v1)
	v1Items := v1["items"].([]any)
	if len(v1Items) != 2 {
		t.Fatalf("v1 items mutated: %v", v1Items)
	}
	first := v1Items[0].(map[string]any)
	if first["label"] != "Freins" {
		t.Fatalf("v1 first label = %v", first["label"])
	}

	// Versions list
	versRec := doJSON(t, handler, http.MethodGet, "/api/v1/templates/"+strconv.FormatInt(id, 10)+"/versions", nil, session, "")
	if versRec.Code != http.StatusOK {
		t.Fatalf("versions status = %d", versRec.Code)
	}
	var vers map[string]any
	_ = json.Unmarshal(versRec.Body.Bytes(), &vers)
	if len(vers["versions"].([]any)) != 2 {
		t.Fatalf("versions = %v", vers["versions"])
	}

	// Archive
	archRec := doJSON(t, handler, http.MethodDelete, "/api/v1/templates/"+strconv.FormatInt(id, 10), nil, session, csrf)
	if archRec.Code != http.StatusNoContent {
		t.Fatalf("archive status = %d body=%s", archRec.Code, archRec.Body.String())
	}
	getRec := doJSON(t, handler, http.MethodGet, "/api/v1/templates/"+strconv.FormatInt(id, 10), nil, session, "")
	if getRec.Code != http.StatusNotFound {
		t.Fatalf("get archived status = %d, want 404", getRec.Code)
	}
}

func TestTemplatesAPI_RequiresAuthAndCSRF(t *testing.T) {
	t.Parallel()

	handler := newTestRouter(t, config.Config{
		SessionSecret:       "test-secret-at-least-thirty-two-bytes",
		BootstrapAdminEmail: "admin@example.com",
		Env:                 "development",
	})

	rec := doJSON(t, handler, http.MethodGet, "/api/v1/templates", nil, nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauth list status = %d", rec.Code)
	}

	session, _ := registerAndSession(t, handler, "csrf-tpl@example.com", "CSRF")
	body := map[string]any{
		"name":  "X",
		"items": []map[string]any{{"label": "A"}},
	}
	noCSRF := doJSON(t, handler, http.MethodPost, "/api/v1/templates", body, session, "")
	if noCSRF.Code != http.StatusForbidden {
		t.Fatalf("create without csrf status = %d", noCSRF.Code)
	}
}

func registerAndSession(t *testing.T, handler http.Handler, email, name string) (*http.Cookie, string) {
	t.Helper()
	guest, guestCSRF := bootstrapGuest(t, handler)
	payload := map[string]string{
		"email":            email,
		"display_name":     name,
		"password":         "password123",
		"password_confirm": "password123",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", guestCSRF)
	req.AddCookie(guest)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("register: %d %s", rec.Code, rec.Body.String())
	}
	var authOK map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &authOK)
	csrf, _ := authOK["csrf_token"].(string)
	var session *http.Cookie
	for _, c := range resultCookies(t, rec) {
		if c.Name == "revues_session" {
			session = c
		}
	}
	if session == nil || csrf == "" {
		t.Fatal("missing session or csrf after register")
	}
	return session, csrf
}

func doJSON(t *testing.T, handler http.Handler, method, path string, body any, session *http.Cookie, csrf string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if csrf != "" {
		req.Header.Set("X-CSRF-Token", csrf)
	}
	if session != nil {
		req.AddCookie(session)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}
