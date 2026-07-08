package auth

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/jeb-maker/revues/internal/web/middleware"
	"github.com/jeb-maker/revues/internal/web/templates"
)

func (h *Auth) DevLogin(w http.ResponseWriter, r *http.Request) {
	if h.Config.Env != "development" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if user, ok := middleware.UserFromContext(r.Context()); ok {
		slog.Debug("already authenticated", "user_id", user.ID)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		email = h.Config.BootstrapAdminEmail
	}
	if email == "" {
		email = "dev@local"
	}

	login := strings.Split(email, "@")[0]
	displayName := login
	role := "admin"

	user, err := h.Store.UpsertGitHubUser(r.Context(), -1, login, email, displayName, "", role)
	if err != nil {
		slog.Error("dev login upsert user", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	sessionToken, _, err := h.Sessions.CreateLoginSession(r.Context(), user.ID)
	if err != nil {
		slog.Error("dev login create session", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.Sessions.SetSessionCookie(w, sessionToken)
	http.Redirect(w, r, "/projects", http.StatusFound)
}

func (h *Auth) DevLoginPage(w http.ResponseWriter, r *http.Request) {
	if h.Config.Env != "development" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if user, ok := middleware.UserFromContext(r.Context()); ok {
		slog.Debug("already authenticated", "user_id", user.ID)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	data := templates.PageData{Title: "Connexion (dev)"}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "devlogin", data); err != nil {
		slog.Error("render dev login page", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
