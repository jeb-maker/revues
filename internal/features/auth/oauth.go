package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/jeb-maker/revues/internal/auth"
	appmiddleware "github.com/jeb-maker/revues/internal/web/middleware"
)

// OAuthHandlers exposes browser redirect endpoints for GitHub OAuth (+ optional DevAuth switch).
type OAuthHandlers struct {
	Service *Service
}

// StartGitHub redirects to GitHub authorize (PKCE).
func (h *OAuthHandlers) StartGitHub(w http.ResponseWriter, r *http.Request) {
	if !h.Service.GitHubOAuthConfigured() {
		http.Redirect(w, r, "/login?error="+url.QueryEscape("oauth non configuré"), http.StatusFound)
		return
	}

	// Drop any existing session (e.g. DevAuth demo) so a cancelled OAuth cannot leave you logged in.
	if token, err := auth.SessionTokenFromRequest(r); err == nil && token != "" {
		if clearErr := h.Service.Sessions.ClearSession(r.Context(), token); clearErr != nil {
			slog.Debug("oauth start clear previous session", "err", clearErr)
		}
		h.Service.Sessions.ClearSessionCookie(w)
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

	payload, signature := h.Service.Sessions.BuildOAuthCookiePayload(state, verifier)
	h.Service.Sessions.SetOAuthCookie(w, payload, signature)

	redir := h.Service.GitHub.AuthURL(state, auth.PKCEChallenge(verifier))
	http.Redirect(w, r, redir, http.StatusFound)
}

// Callback completes the GitHub OAuth code exchange and redirects to the SPA.
func (h *OAuthHandlers) Callback(w http.ResponseWriter, r *http.Request) {
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		http.Redirect(w, r, "/login?error="+url.QueryEscape(errParam), http.StatusFound)
		return
	}

	state, verifier, err := h.Service.Sessions.ParseOAuthCookie(r)
	if err != nil {
		slog.Error("oauth cookie", "err", err)
		http.Redirect(w, r, "/login?error="+url.QueryEscape("session oauth invalide"), http.StatusFound)
		return
	}

	if !auth.ConstantTimeEqual(state, r.URL.Query().Get("state")) {
		http.Redirect(w, r, "/login?error="+url.QueryEscape("state invalide"), http.StatusFound)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Redirect(w, r, "/login?error="+url.QueryEscape("code manquant"), http.StatusFound)
		return
	}

	accessToken, err := h.Service.GitHub.ExchangeCode(r.Context(), code, verifier)
	if err != nil {
		slog.Error("oauth exchange", "err", err)
		http.Redirect(w, r, "/login?error="+url.QueryEscape("échec oauth"), http.StatusFound)
		return
	}

	profile, err := h.Service.GitHub.FetchProfile(r.Context(), accessToken)
	if err != nil {
		slog.Error("github profile", "err", err)
		http.Redirect(w, r, "/login?error="+url.QueryEscape("profil github"), http.StatusFound)
		return
	}

	result, err := h.Service.CompleteGitHubLogin(r.Context(), profile)
	if err != nil {
		if errors.Is(err, ErrEmailNotAllowed) {
			http.Redirect(w, r, "/login?error="+url.QueryEscape("email non autorisé"), http.StatusFound)
			return
		}
		slog.Error("github login", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.Service.Sessions.ClearOAuthCookie(w)
	h.Service.Sessions.ClearGuestCookie(w)
	h.Service.Sessions.SetSessionCookie(w, result.SessionRaw)
	http.Redirect(w, r, result.Redirect, http.StatusFound)
}

// DevLogin switches the local session to another seeded user (REVUES_DEV_AUTH + loopback only).
func (h *OAuthHandlers) DevLogin(w http.ResponseWriter, r *http.Request) {
	if !h.Service.Config.DevAuthEnabled() || !appmiddleware.IsLocalDevRequest(r) {
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

	user, err := h.Service.Store.UserByID(r.Context(), userID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if token, tokenErr := auth.SessionTokenFromRequest(r); tokenErr == nil && token != "" {
		if clearErr := h.Service.Sessions.ClearSession(r.Context(), token); clearErr != nil {
			slog.Debug("dev login clear previous session", "err", clearErr)
		}
	}

	sessionOrgID, redirect, err := PostLoginRoute(r.Context(), h.Service.Store, user.ID)
	if err != nil {
		slog.Error("dev login org route", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	sessionToken, _, err := h.Service.Sessions.CreateLoginSession(r.Context(), user.ID, sessionOrgID)
	if err != nil {
		slog.Error("dev login session", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.Service.Sessions.SetSessionCookie(w, sessionToken)
	if next := strings.TrimSpace(r.FormValue("next")); next == "/" || next == "/login" {
		redirect = next
	}
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}
