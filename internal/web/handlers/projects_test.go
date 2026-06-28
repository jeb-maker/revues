package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

func TestIDOR_CrossProject(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)

	userA, err := st.UpsertGitHubUser(ctx, 10, "alice", "alice@example.com", "Alice", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(alice): %v", err)
	}
	bob, err := st.UpsertGitHubUser(ctx, 11, "bob", "bob@example.com", "Bob", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(bob): %v", err)
	}

	project, err := st.CreateProject(ctx, "Secret", "hidden", userA.ID)
	if err != nil {
		t.Fatalf("CreateProject(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	bobToken, _, err := sessions.CreateLoginSession(ctx, bob.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(bob): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/projects/"+strconv.FormatInt(project.ID, 10), nil)
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: bobToken})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d (IDOR must return 404)", rec.Code, http.StatusNotFound)
	}
}

func TestProjects_CreateAndList(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)

	editor, err := st.UpsertGitHubUser(ctx, 20, "carol", "carol@example.com", "Carol", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, editor.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}
	csrf := auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes")

	form := url.Values{}
	form.Set("csrf_token", csrf)
	form.Set("name", "Projet test")
	form.Set("description", "desc")
	req := httptest.NewRequest(http.MethodPost, "/projects", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("create status = %d, want %d", rec.Code, http.StatusSeeOther)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/projects", nil)
	listReq.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	listRec := httptest.NewRecorder()
	handler.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", listRec.Code, http.StatusOK)
	}
	if !strings.Contains(listRec.Body.String(), "Projet test") {
		t.Fatal("expected project name in list")
	}
}

func TestProjects_ReaderCannotCreate(t *testing.T) {
	handler, db := testRouter(t)
	ctx := context.Background()
	st := store.New(db)

	reader, err := st.UpsertGitHubUser(ctx, 30, "dave", "dave@example.com", "Dave", "", auth.RoleReader)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}

	sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
	token, _, err := sessions.CreateLoginSession(ctx, reader.ID)
	if err != nil {
		t.Fatalf("CreateLoginSession(): %v", err)
	}

	form := url.Values{}
	form.Set("csrf_token", auth.CSRFToken(token, "test-secret-at-least-thirty-two-bytes"))
	form.Set("name", "Nope")
	req := httptest.NewRequest(http.MethodPost, "/projects", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestDashboardEmptyState_ByRole(t *testing.T) {
	roles := []struct {
		role    string
		want    string
		notWant string
	}{
		{auth.RoleAdmin, "Gérer les utilisateurs autorisés", "ne vous est encore assigné"},
		{auth.RoleEditor, "Créer un projet", "Gérer les utilisateurs autorisés"},
		{auth.RoleReader, "ne vous est encore assigné", "Créer un projet"},
	}

	for _, tt := range roles {
		t.Run(tt.role, func(t *testing.T) {
			handler, db := testRouter(t)
			ctx := context.Background()
			st := store.New(db)

			user, err := st.UpsertGitHubUser(ctx, 40, "user-"+tt.role, tt.role+"@example.com", tt.role, "", tt.role)
			if err != nil {
				t.Fatalf("UpsertGitHubUser(): %v", err)
			}

			sessions := &auth.SessionManager{Store: st, SessionSecret: "test-secret-at-least-thirty-two-bytes"}
			token, _, err := sessions.CreateLoginSession(ctx, user.ID)
			if err != nil {
				t.Fatalf("CreateLoginSession(): %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, "/projects", nil)
			req.AddCookie(&http.Cookie{Name: "revues_session", Value: token})
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}
			body := rec.Body.String()
			if !strings.Contains(body, "empty-state") {
				t.Fatal("expected empty dashboard state")
			}
			if !strings.Contains(body, tt.want) {
				t.Fatalf("expected CTA %q in body", tt.want)
			}
			if tt.notWant != "" && strings.Contains(body, tt.notWant) {
				t.Fatalf("unexpected CTA %q in body", tt.notWant)
			}
		})
	}
}
