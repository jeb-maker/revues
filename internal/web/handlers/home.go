package handlers

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/web/middleware"
	"github.com/jeb-maker/revues/internal/web/templates"
)

// Home renders the landing page.
type Home struct {
	Templates     *template.Template
	SessionSecret string
}

// ServeHTTP implements http.Handler.
func (h *Home) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := templates.PageData{Title: "Accueil"}

	if user, ok := middleware.UserFromContext(r.Context()); ok {
		data.User = user
		if token := middleware.SessionTokenFromContext(r); token != "" {
			data.CSRFToken = auth.CSRFToken(token, h.SessionSecret)
		}
	}

	if err := h.Templates.ExecuteTemplate(w, "base", data); err != nil {
		slog.Error("render home page", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
