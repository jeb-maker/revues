package auth

import (
	"context"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"unicode"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/features/organizations"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/middleware"
	"github.com/jeb-maker/revues/internal/web/templates"
)

type Auth struct {
	Templates *template.Template
	Store     AuthStore
	Sessions  *auth.SessionManager
	GitHub    *auth.GitHubOAuth
	Config    config.Config
}

func (h *Auth) Login(w http.ResponseWriter, r *http.Request) {
	if user, ok := middleware.UserFromContext(r.Context()); ok {
		slog.Debug("already authenticated", "user_id", user.ID)
		if dest := h.postLoginRedirect(r.Context(), user.ID); dest != "" {
			http.Redirect(w, r, dest, http.StatusFound)
			return
		}
		http.Redirect(w, r, "/revues", http.StatusFound)
		return
	}

	devAuth := h.Config.DevAuthEnabled() && middleware.IsLocalDevRequest(r)
	data := templates.ApplyPageMeta(templates.PageData{
		LoginError: auth.LoginErrorMessage(r.URL.Query().Get("error")),
		DevAuth:    devAuth,
	}, templates.BCLogin())
	if err := h.attachPublicCSRF(w, r, &data); err != nil {
		slog.Error("login csrf", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if devAuth {
		if users, err := h.Store.ListUsers(r.Context()); err == nil {
			data.DevAuthUsers = users
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "login", data); err != nil {
		slog.Error("render login page", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// PasswordLogin authenticates email + password accounts.
func (h *Auth) PasswordLogin(w http.ResponseWriter, r *http.Request) {
	if _, ok := middleware.UserFromContext(r.Context()); ok {
		http.Redirect(w, r, "/revues", http.StatusSeeOther)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	if email == "" || password == "" {
		h.renderLoginError(w, r, "identifiants invalides", email)
		return
	}

	user, passwordHash, err := h.Store.UserCredentialsByEmail(r.Context(), email)
	if err != nil || passwordHash == "" || !auth.VerifyPassword(passwordHash, password) {
		// Constant-ish work on missing accounts to limit timing oracle.
		if errors.Is(err, store.ErrUserNotFound) || passwordHash == "" {
			_, _ = auth.HashPassword("timing-padding-password")
		}
		h.renderLoginError(w, r, "identifiants invalides", email)
		return
	}

	if err = h.finishLocalLogin(w, r, user); err != nil {
		slog.Error("password login", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// RegisterForm shows the manual registration page.
func (h *Auth) RegisterForm(w http.ResponseWriter, r *http.Request) {
	if _, ok := middleware.UserFromContext(r.Context()); ok {
		http.Redirect(w, r, "/revues", http.StatusFound)
		return
	}

	data := templates.ApplyPageMeta(templates.PageData{}, templates.BCRegister())
	if err := h.attachPublicCSRF(w, r, &data); err != nil {
		slog.Error("register csrf", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "register", data); err != nil {
		slog.Error("render register page", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Register creates a local email/password account.
func (h *Auth) Register(w http.ResponseWriter, r *http.Request) {
	if _, ok := middleware.UserFromContext(r.Context()); ok {
		http.Redirect(w, r, "/revues", http.StatusSeeOther)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	displayName := strings.TrimSpace(r.FormValue("display_name"))
	password := r.FormValue("password")
	confirm := r.FormValue("password_confirm")

	form := templates.PageData{
		FormEmail:       email,
		FormDisplayName: displayName,
	}

	if _, err := mail.ParseAddress(email); err != nil || !strings.Contains(email, "@") {
		h.renderRegisterError(w, r, form, "Adresse email invalide.")
		return
	}
	email = strings.ToLower(email)

	if displayName == "" {
		h.renderRegisterError(w, r, form, "Indiquez un nom affiché.")
		return
	}
	if password != confirm {
		h.renderRegisterError(w, r, form, "Les mots de passe ne correspondent pas.")
		return
	}
	if err := auth.ValidatePasswordStrength(password); err != nil {
		h.renderRegisterError(w, r, form, "Le mot de passe doit contenir entre 8 et 72 caractères.")
		return
	}

	role, err := h.Store.ResolveLoginRoleStrict(r.Context(), email, h.Config.BootstrapAdminEmail, h.Config.LoginRequireWhitelist)
	if err != nil {
		if errors.Is(err, store.ErrEmailNotAllowed) {
			// Generic copy — do not reveal whitelist membership.
			h.renderRegisterError(w, r, form, "Impossible de créer ce compte. Contactez un administrateur si besoin.")
			return
		}
		slog.Error("register resolve role", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		slog.Error("register hash password", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	login := localLoginFromEmail(email)
	user, err := h.Store.CreateLocalUser(r.Context(), email, login, displayName, role, passwordHash)
	if err != nil {
		if errors.Is(err, store.ErrEmailTaken) {
			h.renderRegisterError(w, r, form, "Impossible de créer ce compte. Contactez un administrateur si besoin.")
			return
		}
		slog.Error("register create user", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err = h.Store.EnsureBootstrapOrgOwner(r.Context(), user.ID, email, h.Config.BootstrapAdminEmail); err != nil {
		slog.Error("register bootstrap org", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err = h.finishLocalLogin(w, r, user); err != nil {
		slog.Error("register session", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// DevLogin switches the local session to another seeded user (REVUES_DEV_AUTH + loopback only).
func (h *Auth) DevLogin(w http.ResponseWriter, r *http.Request) {
	if !h.Config.DevAuthEnabled() || !middleware.IsLocalDevRequest(r) {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(strings.TrimSpace(r.FormValue("user_id")), 10, 64)
	if err != nil || userID <= 0 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	user, err := h.Store.UserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			http.NotFound(w, r)
			return
		}
		slog.Error("dev login user", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if token, tokenErr := auth.SessionTokenFromRequest(r); tokenErr == nil && token != "" {
		if clearErr := h.Sessions.ClearSession(r.Context(), token); clearErr != nil {
			slog.Debug("dev login clear previous session", "err", clearErr)
		}
	}

	sessionOrgID, redirect, err := organizations.PostLoginRoute(r.Context(), h.Store, user.ID)
	if err != nil {
		slog.Error("dev login org route", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	sessionToken, _, err := h.Sessions.CreateLoginSession(r.Context(), user.ID, sessionOrgID)
	if err != nil {
		slog.Error("dev login session", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.Sessions.SetSessionCookie(w, sessionToken)
	if next := strings.TrimSpace(r.FormValue("next")); next == "/revues" || next == "/mes-taches" || next == "/subjects" {
		redirect = next
	}
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

func (h *Auth) StartGitHub(w http.ResponseWriter, r *http.Request) {
	if h.GitHub.ClientID == "" || h.GitHub.ClientSecret == "" {
		http.Redirect(w, r, "/login?error=oauth+non+configur%C3%A9", http.StatusFound)
		return
	}

	state, _, err := auth.RandomToken(16)
	if err != nil {
		slog.Error("oauth state", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	verifier, _, err := auth.RandomToken(32)
	if err != nil {
		slog.Error("oauth verifier", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	payload, signature := h.Sessions.BuildOAuthCookiePayload(state, verifier)
	h.Sessions.SetOAuthCookie(w, payload, signature)

	url := h.GitHub.AuthURL(state, auth.PKCEChallenge(verifier))
	http.Redirect(w, r, url, http.StatusFound)
}

func (h *Auth) Callback(w http.ResponseWriter, r *http.Request) {
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		http.Redirect(w, r, "/login?error="+errParam, http.StatusFound)
		return
	}

	state, verifier, err := h.Sessions.ParseOAuthCookie(r)
	if err != nil {
		slog.Error("oauth cookie", "err", err)
		http.Redirect(w, r, "/login?error=session+oauth+invalide", http.StatusFound)
		return
	}

	if !auth.ConstantTimeEqual(state, r.URL.Query().Get("state")) {
		http.Redirect(w, r, "/login?error=state+invalide", http.StatusFound)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Redirect(w, r, "/login?error=code+manquant", http.StatusFound)
		return
	}

	accessToken, err := h.GitHub.ExchangeCode(r.Context(), code, verifier)
	if err != nil {
		slog.Error("oauth exchange", "err", err)
		http.Redirect(w, r, "/login?error=%C3%A9chec+oauth", http.StatusFound)
		return
	}

	profile, err := h.GitHub.FetchProfile(r.Context(), accessToken)
	if err != nil {
		slog.Error("github profile", "err", err)
		http.Redirect(w, r, "/login?error=profil+github", http.StatusFound)
		return
	}

	role, err := h.Store.ResolveLoginRoleStrict(r.Context(), profile.Email, h.Config.BootstrapAdminEmail, h.Config.LoginRequireWhitelist)
	if err != nil {
		if errors.Is(err, store.ErrEmailNotAllowed) {
			http.Redirect(w, r, "/login?error=email+non+autoris%C3%A9", http.StatusFound)
			return
		}
		slog.Error("resolve login role", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	displayName := profile.DisplayName
	if displayName == "" {
		displayName = profile.Login
	}

	user, err := h.Store.UpsertGitHubUser(
		r.Context(),
		profile.ID,
		profile.Login,
		profile.Email,
		displayName,
		profile.AvatarURL,
		role,
	)
	if err != nil {
		slog.Error("upsert user", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err = h.Store.EnsureBootstrapOrgOwner(r.Context(), user.ID, profile.Email, h.Config.BootstrapAdminEmail); err != nil {
		slog.Error("bootstrap org owner", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	sessionOrgID, redirect, err := organizations.PostLoginRoute(r.Context(), h.Store, user.ID)
	if err != nil {
		slog.Error("post-login organization route", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	sessionToken, _, err := h.Sessions.CreateLoginSession(r.Context(), user.ID, sessionOrgID)
	if err != nil {
		slog.Error("create session", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.Sessions.ClearOAuthCookie(w)
	h.Sessions.ClearGuestCookie(w)
	h.Sessions.SetSessionCookie(w, sessionToken)
	http.Redirect(w, r, redirect, http.StatusFound)
}

func (h *Auth) postLoginRedirect(ctx context.Context, userID int64) string {
	_, redirect, err := organizations.PostLoginRoute(ctx, h.Store, userID)
	if err != nil {
		return ""
	}
	if redirect == "/revues" {
		return ""
	}
	return redirect
}

func (h *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	token, err := auth.SessionTokenFromRequest(r)
	if err == nil {
		if err := h.Sessions.ClearSession(r.Context(), token); err != nil {
			slog.Error("clear session", "err", err)
		}
	}

	h.Sessions.ClearSessionCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Auth) finishLocalLogin(w http.ResponseWriter, r *http.Request, user *store.User) error {
	if err := h.Store.TouchLastLogin(r.Context(), user.ID); err != nil {
		return err
	}

	sessionOrgID, redirect, err := organizations.PostLoginRoute(r.Context(), h.Store, user.ID)
	if err != nil {
		return err
	}

	sessionToken, _, err := h.Sessions.CreateLoginSession(r.Context(), user.ID, sessionOrgID)
	if err != nil {
		return err
	}

	h.Sessions.ClearGuestCookie(w)
	h.Sessions.SetSessionCookie(w, sessionToken)
	http.Redirect(w, r, redirect, http.StatusSeeOther)
	return nil
}

func (h *Auth) attachPublicCSRF(w http.ResponseWriter, r *http.Request, data *templates.PageData) error {
	if token := middleware.SessionTokenFromContext(r); token != "" {
		data.CSRFToken = auth.CSRFToken(token, h.Config.SessionSecret)
		return nil
	}
	_, csrf, err := h.Sessions.EnsureGuestToken(w, r)
	if err != nil {
		return err
	}
	data.CSRFToken = csrf
	return nil
}

func (h *Auth) renderLoginError(w http.ResponseWriter, r *http.Request, code, email string) {
	devAuth := h.Config.DevAuthEnabled() && middleware.IsLocalDevRequest(r)
	data := templates.ApplyPageMeta(templates.PageData{
		LoginError: auth.LoginErrorMessage(code),
		FormEmail:  email,
		DevAuth:    devAuth,
	}, templates.BCLogin())
	if err := h.attachPublicCSRF(w, r, &data); err != nil {
		slog.Error("login error csrf", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if devAuth {
		if users, err := h.Store.ListUsers(r.Context()); err == nil {
			data.DevAuthUsers = users
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	if err := h.Templates.ExecuteTemplate(w, "login", data); err != nil {
		slog.Error("render login error", "err", err)
	}
}

func (h *Auth) renderRegisterError(w http.ResponseWriter, r *http.Request, data templates.PageData, message string) {
	data.LoginError = message
	data = templates.ApplyPageMeta(data, templates.BCRegister())
	if err := h.attachPublicCSRF(w, r, &data); err != nil {
		slog.Error("register error csrf", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	if err := h.Templates.ExecuteTemplate(w, "register", data); err != nil {
		slog.Error("render register error", "err", err)
	}
}

func localLoginFromEmail(email string) string {
	local := email
	if at := strings.IndexByte(email, '@'); at > 0 {
		local = email[:at]
	}
	var b strings.Builder
	for _, r := range strings.ToLower(local) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if out == "" {
		return "user"
	}
	if len(out) > 64 {
		return out[:64]
	}
	return out
}
