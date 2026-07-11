package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/store"
)

func TestAdminWebhooks_ReaderForbidden(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)
	reader, _ := st.UpsertGitHubUser(ctx, 1, "reader", "reader@example.com", "Reader", "", auth.RoleReader)
	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, _ := sessions.CreateLoginSession(ctx, reader.ID, 0)
	req := httptest.NewRequest(http.MethodGet, "/admin/settings/webhooks", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestAdminWebhooks_SaveAndTest(t *testing.T) {
	receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	t.Cleanup(receiver.Close)
	handler, db := testRouterWithEncryptionKey(t, config.TestEncryptionKey())
	ctx := context.Background()
	st := store.New(db)
	adminUser, _ := st.UpsertGitHubUser(ctx, 99, "admin", "admin@example.com", "Admin", "", auth.RoleAdmin)
	secret := "test-secret-at-least-thirty-two-bytes"
	sessions := &auth.SessionManager{Store: st, SessionSecret: secret}
	token, _, _ := sessions.CreateLoginSession(ctx, adminUser.ID, 0)
	csrf := auth.CSRFToken(token, secret)
	saveForm := url.Values{"csrf_token": {csrf}, "action": {"save"}, "urls": {receiver.URL}, "secret": {"webhook-hmac-secret"}, "review_completed": {"on"}, "review_item_nok": {"on"}}
	saveReq := httptest.NewRequest(http.MethodPost, "/admin/settings/webhooks", strings.NewReader(saveForm.Encode()))
	saveReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	saveReq.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	saveRec := httptest.NewRecorder()
	handler.ServeHTTP(saveRec, saveReq)
	if saveRec.Code != http.StatusSeeOther {
		t.Fatalf("save status = %d", saveRec.Code)
	}
	testForm := url.Values{"csrf_token": {csrf}, "action": {"test"}}
	testReq := httptest.NewRequest(http.MethodPost, "/admin/settings/webhooks", strings.NewReader(testForm.Encode()))
	testReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	testReq.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	testRec := httptest.NewRecorder()
	handler.ServeHTTP(testRec, testReq)
	if testRec.Code != http.StatusSeeOther {
		t.Fatalf("test status = %d body=%q", testRec.Code, testRec.Body.String())
	}
}
