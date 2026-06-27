package web_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
	appmiddleware "github.com/jeb-maker/revues/internal/web/middleware"
)

func TestRBAC_Matrix(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	st := store.New(db)

	users := map[string]*store.User{}
	for role, githubID := range map[string]int64{
		auth.RoleAdmin:  1,
		auth.RoleEditor: 2,
		auth.RoleReader: 3,
	} {
		user, err := st.UpsertGitHubUser(ctx, githubID, role+"-user", role+"@example.com", role, "", role)
		if err != nil {
			t.Fatalf("UpsertGitHubUser(%s): %v", role, err)
		}
		users[role] = user
	}

	sessions := &auth.SessionManager{
		Store:         st,
		SessionSecret: "test-secret-at-least-thirty-two-bytes",
	}

	tokens := map[string]string{}
	for role, user := range users {
		token, _, err := sessions.CreateLoginSession(ctx, user.ID)
		if err != nil {
			t.Fatalf("CreateLoginSession(%s): %v", role, err)
		}
		tokens[role] = token
	}

	mux := chi.NewRouter()
	mux.Use(appmiddleware.LoadUser(st))
	mux.With(appmiddleware.RequireAuth).Get("/protected", okHandler)
	mux.With(appmiddleware.RequireAuth, appmiddleware.RequireRole(auth.RoleEditor)).Get("/editor", okHandler)
	mux.With(appmiddleware.RequireAuth, appmiddleware.RequireRole(auth.RoleAdmin)).Get("/admin/users", okHandler)

	tests := []struct {
		name       string
		path       string
		token      string
		wantStatus int
	}{
		{"protected anonymous", "/protected", "", http.StatusFound},
		{"protected reader", "/protected", tokens[auth.RoleReader], http.StatusOK},
		{"editor zone reader denied", "/editor", tokens[auth.RoleReader], http.StatusForbidden},
		{"editor zone editor ok", "/editor", tokens[auth.RoleEditor], http.StatusOK},
		{"editor zone admin ok", "/editor", tokens[auth.RoleAdmin], http.StatusOK},
		{"admin zone reader denied", "/admin/users", tokens[auth.RoleReader], http.StatusForbidden},
		{"admin zone editor denied", "/admin/users", tokens[auth.RoleEditor], http.StatusForbidden},
		{"admin zone admin ok", "/admin/users", tokens[auth.RoleAdmin], http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.token != "" {
				req.AddCookie(&http.Cookie{Name: "revues_session", Value: tt.token})
			}
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/rbac.db")
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	t.Cleanup(func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Errorf("Close(): %v", closeErr)
		}
	})
	if migrateErr := store.Migrate(ctx, db); migrateErr != nil {
		t.Fatalf("Migrate(): %v", migrateErr)
	}
	return db
}
