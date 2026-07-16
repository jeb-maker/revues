package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/store"
	appmiddleware "github.com/jeb-maker/revues/internal/web/middleware"
)

func localDevRequest(method, target string) *http.Request {
	req := httptest.NewRequest(method, target, nil)
	req.Host = "127.0.0.1:8080"
	req.RemoteAddr = "127.0.0.1:54321"
	return req
}

func TestEnsureDevAuth_InjectsUserAndCookie(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/test.db", 0)
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err = store.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate(): %v", err)
	}

	st := store.New(db)
	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}

	var gotUser bool
	var gotToken string
	handler := appmiddleware.CapturePeerAddr(
		appmiddleware.LoadUser(st)(
			appmiddleware.EnsureDevAuth(st, sessions, true, "admin@example.com")(
				appmiddleware.LoadActiveOrganization(st)(
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if _, ok := appmiddleware.UserFromContext(r.Context()); ok {
							gotUser = true
						}
						gotToken = appmiddleware.SessionTokenFromContext(r)
						if _, ok := appmiddleware.OrganizationFromContext(r.Context()); !ok {
							t.Error("expected organization in context")
						}
						w.WriteHeader(http.StatusOK)
					}),
				),
			),
		),
	)

	req := localDevRequest(http.MethodGet, "/revues")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !gotUser {
		t.Fatal("expected user in context")
	}
	if gotToken == "" {
		t.Fatal("expected session token in context")
	}
	foundCookie := false
	for _, c := range rec.Result().Cookies() {
		if c.Name == "revues_session" && c.Value != "" {
			foundCookie = true
		}
	}
	if !foundCookie {
		t.Fatal("expected revues_session cookie")
	}
}

func TestEnsureDevAuth_RejectsNonLoopback(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/test.db", 0)
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err = store.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate(): %v", err)
	}

	st := store.New(db)
	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}

	var gotUser bool
	handler := appmiddleware.CapturePeerAddr(
		appmiddleware.EnsureDevAuth(st, sessions, true, "admin@example.com")(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, gotUser = appmiddleware.UserFromContext(r.Context())
				w.WriteHeader(http.StatusOK)
			}),
		),
	)

	req := httptest.NewRequest(http.MethodGet, "/revues", nil)
	req.Host = "192.168.1.10:8080"
	req.RemoteAddr = "192.168.1.20:9999"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if gotUser {
		t.Fatal("dev auth must not inject user for non-loopback requests")
	}
}

func TestEnsureDevAuth_RejectsSpoofedForwardedFor(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/test.db", 0)
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err = store.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate(): %v", err)
	}

	st := store.New(db)
	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}

	var gotUser bool
	// CapturePeerAddr then pretend RealIP rewrote RemoteAddr to 127.0.0.1
	handler := appmiddleware.CapturePeerAddr(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.RemoteAddr = "127.0.0.1:1" // spoof after capture
		appmiddleware.EnsureDevAuth(st, sessions, true, "admin@example.com")(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, gotUser = appmiddleware.UserFromContext(r.Context())
				w.WriteHeader(http.StatusOK)
			}),
		).ServeHTTP(w, r)
	}))

	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8080/revues", nil)
	req.Host = "127.0.0.1:8080"
	req.RemoteAddr = "203.0.113.10:9999"
	req.Header.Set("X-Forwarded-For", "127.0.0.1")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if gotUser {
		t.Fatal("dev auth must ignore spoofed X-Forwarded-For / rewritten RemoteAddr")
	}
}

func TestEnsureDevAuth_DisabledIsNoop(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/test.db", 0)
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err = store.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate(): %v", err)
	}

	st := store.New(db)
	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}

	var gotUser bool
	handler := appmiddleware.CapturePeerAddr(
		appmiddleware.EnsureDevAuth(st, sessions, false, "admin@example.com")(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, gotUser = appmiddleware.UserFromContext(r.Context())
				w.WriteHeader(http.StatusOK)
			}),
		),
	)

	req := localDevRequest(http.MethodGet, "/revues")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if gotUser {
		t.Fatal("disabled dev auth must not inject user")
	}
}

func TestDevAuthEnabled_NeverInProduction(t *testing.T) {
	t.Setenv("REVUES_ENV", "production")
	t.Setenv("REVUES_DEV_AUTH", "1")
	cfg := config.Load()
	if cfg.DevAuthEnabled() {
		t.Fatal("DevAuthEnabled must be false in production")
	}
}

func TestDevAuthEnabled_InDevelopment(t *testing.T) {
	t.Setenv("REVUES_ENV", "development")
	t.Setenv("REVUES_DEV_AUTH", "1")
	cfg := config.Load()
	if !cfg.DevAuthEnabled() {
		t.Fatal("DevAuthEnabled must be true in development when flag set")
	}
}
