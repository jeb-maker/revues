package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	authhandler "github.com/jeb-maker/revues/internal/features/auth"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/testutil"
	"github.com/jeb-maker/revues/internal/web/templates"
)

func newLocalAuthHandler(t *testing.T) (*authhandler.Auth, *store.Store, *auth.SessionManager) {
	t.Helper()

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
	_ = testutil.DefaultOrgContext(ctx, st)

	tpl, err := templates.Parse("")
	if err != nil {
		t.Fatalf("Parse(): %v", err)
	}
	sessions := &auth.SessionManager{
		Store:         st,
		SessionSecret: "test-secret-at-least-thirty-two-bytes",
	}
	h := &authhandler.Auth{
		Templates: tpl,
		Store:     st,
		Sessions:  sessions,
		Config: config.Config{
			SessionSecret:         "test-secret-at-least-thirty-two-bytes",
			BootstrapAdminEmail:   "admin@example.com",
			LoginRequireWhitelist: false,
			Env:                   "development",
		},
	}
	return h, st, sessions
}

func guestCSRF(t *testing.T, sessions *auth.SessionManager) (cookie *http.Cookie, csrf string) {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	_, csrf, err := sessions.EnsureGuestToken(rec, req)
	if err != nil {
		t.Fatalf("EnsureGuestToken(): %v", err)
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == "revues_guest" {
			return c, csrf
		}
	}
	t.Fatal("missing guest cookie")
	return nil, ""
}

func TestRegister_CreatesSession(t *testing.T) {
	t.Parallel()

	h, st, sessions := newLocalAuthHandler(t)
	guest, csrf := guestCSRF(t, sessions)

	form := url.Values{
		"csrf_token":       {csrf},
		"email":            {"newbie@example.com"},
		"display_name":     {"Newbie"},
		"password":         {"password123"},
		"password_confirm": {"password123"},
	}
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(guest)
	rec := httptest.NewRecorder()
	h.Register(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	var sessionCookie string
	for _, c := range rec.Result().Cookies() {
		if c.Name == "revues_session" && c.Value != "" {
			sessionCookie = c.Value
		}
	}
	if sessionCookie == "" {
		t.Fatal("expected session cookie")
	}

	user, err := st.UserByEmail(context.Background(), "newbie@example.com")
	if err != nil {
		t.Fatalf("UserByEmail(): %v", err)
	}
	if user.GitHubID != 0 {
		t.Fatalf("GitHubID = %d, want 0 for local user", user.GitHubID)
	}
	if user.DisplayName != "Newbie" {
		t.Fatalf("DisplayName = %q", user.DisplayName)
	}
}

func TestRegister_DuplicateEmail_GenericError(t *testing.T) {
	t.Parallel()

	h, st, sessions := newLocalAuthHandler(t)
	ctx := context.Background()
	hash, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword(): %v", err)
	}
	if _, err = st.CreateLocalUser(ctx, "taken@example.com", "taken", "Taken", auth.RoleEditor, hash); err != nil {
		t.Fatalf("CreateLocalUser(): %v", err)
	}

	guest, csrf := guestCSRF(t, sessions)
	form := url.Values{
		"csrf_token":       {csrf},
		"email":            {"taken@example.com"},
		"display_name":     {"Other"},
		"password":         {"password123"},
		"password_confirm": {"password123"},
	}
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(guest)
	rec := httptest.NewRecorder()
	h.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Impossible de créer ce compte") {
		t.Fatalf("body missing generic error: %s", body)
	}
	if strings.Contains(strings.ToLower(body), "déjà utilisé") || strings.Contains(strings.ToLower(body), "already") {
		t.Fatalf("body leaks email existence: %s", body)
	}
}

func TestPasswordLogin_SuccessAndFailure(t *testing.T) {
	t.Parallel()

	h, st, sessions := newLocalAuthHandler(t)
	ctx := context.Background()
	hash, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword(): %v", err)
	}
	if _, err = st.CreateLocalUser(ctx, "local@example.com", "local", "Local", auth.RoleEditor, hash); err != nil {
		t.Fatalf("CreateLocalUser(): %v", err)
	}

	guest, csrf := guestCSRF(t, sessions)

	bad := url.Values{
		"csrf_token": {csrf},
		"email":      {"local@example.com"},
		"password":   {"wrong-password"},
	}
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(bad.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(guest)
	rec := httptest.NewRecorder()
	h.PasswordLogin(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("bad login status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Email ou mot de passe incorrect") {
		t.Fatalf("unexpected error body: %s", rec.Body.String())
	}

	guest2, csrf2 := guestCSRF(t, sessions)
	okForm := url.Values{
		"csrf_token": {csrf2},
		"email":      {"local@example.com"},
		"password":   {"password123"},
	}
	req = httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(okForm.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(guest2)
	rec = httptest.NewRecorder()
	h.PasswordLogin(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("good login status = %d body=%s", rec.Code, rec.Body.String())
	}
	found := false
	for _, c := range rec.Result().Cookies() {
		if c.Name == "revues_session" && c.Value != "" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected session cookie after password login")
	}
}

func TestRegister_PasswordMismatch(t *testing.T) {
	t.Parallel()

	h, _, sessions := newLocalAuthHandler(t)
	guest, csrf := guestCSRF(t, sessions)
	form := url.Values{
		"csrf_token":       {csrf},
		"email":            {"x@example.com"},
		"display_name":     {"X"},
		"password":         {"password123"},
		"password_confirm": {"nope"},
	}
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(guest)
	rec := httptest.NewRecorder()
	h.Register(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "ne correspondent pas") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestRegisterForm_Renders(t *testing.T) {
	t.Parallel()

	h, _, _ := newLocalAuthHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/register", nil)
	rec := httptest.NewRecorder()
	h.RegisterForm(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `action="/auth/register"`) {
		t.Fatalf("missing register form: %s", body)
	}
	if !strings.Contains(body, `name="csrf_token"`) {
		t.Fatal("missing csrf field")
	}
}
