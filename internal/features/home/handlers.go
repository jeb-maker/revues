package home

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/middleware"
	"github.com/jeb-maker/revues/internal/web/templates"
)

// Deps holds shared dependencies for the home HTTP handlers.
//
// This mirrors internal/web/handlerdeps.HandlerDeps but is local to the home
// feature package to avoid an import cycle.
type Deps struct {
	Templates     *template.Template
	Store         *store.Store
	SessionSecret string
}

// Home renders the landing page.
type Home struct {
	Deps
}

// ServeHTTP implements http.Handler.
func (h *Home) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := templates.PageData{Title: "Accueil"}

	if _, ok := middleware.UserFromContext(r.Context()); ok {
		http.Redirect(w, r, "/projects", http.StatusFound)
		return
	}

	if err := h.Templates.ExecuteTemplate(w, "home", data); err != nil {
		slog.Error("render home page", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
