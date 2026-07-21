package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

// mutatingRoutes are POST routes that must reject missing CSRF tokens with 403.
// Keep in sync when adding new mutating endpoints (HTMX included).
var mutatingRoutes = []string{
	"/logout",
	"/org/new",
	"/org/select",
	"/org/switch",
	"/subjects",
	"/admin/users",
	"/admin/settings/webhooks",
	"/signaler",
	"/modeles",
}

func TestCSRF_MissingToken_MutatingRoutes(t *testing.T) {
	handler, db := testRouter(t)
	st := store.New(db)
	ctx := context.Background()

	user, err := st.UpsertGitHubUser(ctx, 88001, "csrf-user", "csrf@example.com", "CSRF", "", "editor")
	if err != nil {
		t.Fatalf("UpsertGitHubUser: %v", err)
	}
	sessions := &auth.SessionManager{
		Store:         st,
		SessionSecret: "test-secret-at-least-thirty-two-bytes",
	}
	token, _, err := sessions.CreateLoginSession(ctx, user.ID, 0)
	if err != nil {
		t.Fatalf("CreateLoginSession: %v", err)
	}

	for _, path := range mutatingRoutes {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, path, nil)
			req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want 403 without CSRF token", rec.Code)
			}
		})
	}
}
