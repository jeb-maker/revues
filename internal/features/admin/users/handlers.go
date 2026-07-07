package users

import (
	"errors"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/handlers"
	"github.com/jeb-maker/revues/internal/web/middleware"
	"github.com/jeb-maker/revues/internal/web/templates"
)

// AdminUsers manages the login email whitelist.
type AdminUsers struct {
	handlers.Deps
}

// List shows whitelisted emails.
func (h *AdminUsers) List(w http.ResponseWriter, r *http.Request) {
	emails, err := h.Store.ListAllowedEmails(r.Context())
	if err != nil {
		slog.Error("list allowed emails", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := h.adminData(r, "Utilisateurs autorisés")
	data.Emails = emails
	data.Message = r.URL.Query().Get("msg")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "admin_users", data); err != nil {
		slog.Error("render admin users", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Add inserts or updates a whitelisted email.
func (h *AdminUsers) Add(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	email, err := normalizeEmail(r.FormValue("email"))
	if err != nil {
		h.renderError(w, r, "Adresse email invalide.")
		return
	}

	role := strings.TrimSpace(r.FormValue("role"))
	if !auth.ValidRole(role) {
		h.renderError(w, r, "Rôle invalide.")
		return
	}

	if err := h.Store.InsertAllowedEmail(r.Context(), email, role); err != nil {
		slog.Error("insert allowed email", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/users?msg=Email+ajout%C3%A9", http.StatusSeeOther)
}

// Remove deletes a whitelisted email.
func (h *AdminUsers) Remove(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	email, err := normalizeEmail(r.FormValue("email"))
	if err != nil {
		h.renderError(w, r, "Adresse email invalide.")
		return
	}

	user, ok := middleware.UserFromContext(r.Context())
	if ok && strings.EqualFold(user.Email, email) {
		h.renderError(w, r, "Vous ne pouvez pas retirer votre propre email.")
		return
	}

	if err := h.Store.DeleteAllowedEmail(r.Context(), email); err != nil {
		if errors.Is(err, store.ErrAllowedEmailNotFound) {
			h.renderError(w, r, "Email introuvable.")
			return
		}
		slog.Error("delete allowed email", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/users?msg=Email+retir%C3%A9", http.StatusSeeOther)
}

func (h *AdminUsers) adminData(r *http.Request, title string) templates.AdminUsersData {
	data := templates.AdminUsersData{
		PageData: h.PageData(r, title),
	}
	data.AdminSection = "users"
	return data
}

func (h *AdminUsers) renderError(w http.ResponseWriter, r *http.Request, message string) {
	emails, err := h.Store.ListAllowedEmails(r.Context())
	if err != nil {
		slog.Error("list allowed emails", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := h.adminData(r, "Utilisateurs autorisés")
	data.Emails = emails
	data.Error = message

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	if err := h.Templates.ExecuteTemplate(w, "admin_users", data); err != nil {
		slog.Error("render admin users error", "err", err)
	}
}

func normalizeEmail(raw string) (string, error) {
	email := strings.TrimSpace(strings.ToLower(raw))
	if email == "" {
		return "", errors.New("empty email")
	}
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return "", err
	}
	return strings.ToLower(addr.Address), nil
}
