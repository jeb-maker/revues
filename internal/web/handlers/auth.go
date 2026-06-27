package handlers

import (
	"errors"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/middleware"
	"github.com/jeb-maker/revues/internal/web/templates"
)

// Auth handles login, OAuth callback and logout.
type Auth struct {
	Templates *template.Template
	Store     *store.Store
	Sessions  *auth.SessionManager
	GitHub    *auth.GitHubOAuth
	Config    config.Config
}

// Login shows the GitHub sign-in page.
func (h *Auth) Login(w http.ResponseWriter, r *http.Request) {
	if user, ok := middleware.UserFromContext(r.Context()); ok {
		slog.Debug("already authenticated", "user_id", user.ID)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	data := templates.PageData{
		Title:      "Connexion",
		LoginError: r.URL.Query().Get("error"),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "login", data); err != nil {
		slog.Error("render login page", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// StartGitHub begins the OAuth Authorization Code + PKCE flow.
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

// Callback completes GitHub OAuth and creates a server session.
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

	role, err := h.Store.ResolveLoginRole(r.Context(), profile.Email, h.Config.BootstrapAdminEmail)
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

	sessionToken, _, err := h.Sessions.CreateLoginSession(r.Context(), user.ID)
	if err != nil {
		slog.Error("create session", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.Sessions.ClearOAuthCookie(w)
	h.Sessions.SetSessionCookie(w, sessionToken)
	http.Redirect(w, r, "/", http.StatusFound)
}

// Logout destroys the current session.
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
