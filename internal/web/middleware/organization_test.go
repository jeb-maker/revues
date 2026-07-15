package middleware_test

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

func TestLoadActiveOrganization_InjectsOrganization(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	st := store.New(db)

	user, err := st.UpsertGitHubUser(ctx, 1, "alice", "alice@example.com", "Alice", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	org, err := st.CreateOrganization(ctx, "Team", "team", user.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, org.ID, user.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, user.ID, org.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	var gotOrg *store.Organization
	handler := chi.NewRouter()
	handler.Use(appmiddleware.LoadUser(st))
	handler.Use(appmiddleware.LoadActiveOrganization(st))
	handler.Get("/subjects", func(w http.ResponseWriter, r *http.Request) {
		ctxOrg, ok := appmiddleware.OrganizationFromContext(r.Context())
		if !ok {
			http.Error(w, "missing organization", http.StatusInternalServerError)
			return
		}
		gotOrg = ctxOrg
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/subjects", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if gotOrg == nil || gotOrg.ID != org.ID {
		t.Fatalf("organization = %+v, want id %d", gotOrg, org.ID)
	}
}

func TestLoadActiveOrganization_RedirectsWithoutMembership(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	st := store.New(db)

	user, err := st.UpsertGitHubUser(ctx, 2, "bob", "bob@example.com", "Bob", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	memberOrg, err := st.CreateOrganization(ctx, "Member", "member-org", user.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(member): %v", err)
	}
	otherOrg, err := st.CreateOrganization(ctx, "Other", "other", user.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(other): %v", err)
	}
	if err = st.AddOrganizationMember(ctx, memberOrg.ID, user.ID, store.OrgRoleMember); err != nil {
		t.Fatalf("AddOrganizationMember(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, user.ID, otherOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	handler := chi.NewRouter()
	handler.Use(appmiddleware.LoadUser(st))
	handler.Use(appmiddleware.LoadActiveOrganization(st))
	handler.Get("/subjects", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/subjects", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if loc := rec.Header().Get("Location"); loc != "/org/select" {
		t.Fatalf("Location = %q, want /org/select", loc)
	}
}

func TestLoadActiveOrganization_ExemptOrgSelect(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	st := store.New(db)

	user, err := st.UpsertGitHubUser(ctx, 3, "carol", "carol@example.com", "Carol", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	otherOrg, err := st.CreateOrganization(ctx, "Foreign", "foreign", user.ID)
	if err != nil {
		t.Fatalf("CreateOrganization(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, user.ID, otherOrg.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	handler := chi.NewRouter()
	handler.Use(appmiddleware.LoadUser(st))
	handler.Use(appmiddleware.LoadActiveOrganization(st))
	handler.Get("/org/select", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/org/select", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestLoadActiveOrganization_RedirectsPendingWithoutOrganizations(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	st := store.New(db)

	user, err := st.UpsertGitHubUser(ctx, 4, "dana", "dana@example.com", "Dana", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, user.ID, auth.SessionOrgPending)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	handler := chi.NewRouter()
	handler.Use(appmiddleware.LoadUser(st))
	handler.Use(appmiddleware.LoadActiveOrganization(st))
	handler.Get("/subjects", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/subjects", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if loc := rec.Header().Get("Location"); loc != "/org/new" {
		t.Fatalf("Location = %q, want /org/new", loc)
	}
}

func TestLoadActiveOrganization_ExemptOrgNew(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	st := store.New(db)

	user, err := st.UpsertGitHubUser(ctx, 5, "erin", "erin@example.com", "Erin", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, user.ID, auth.SessionOrgPending)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	handler := chi.NewRouter()
	handler.Use(appmiddleware.LoadUser(st))
	handler.Use(appmiddleware.LoadActiveOrganization(st))
	handler.Get("/org/new", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/org/new", nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/middleware.db", 0)
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
