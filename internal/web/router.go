package web

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/jeb-maker/revues/internal/web/handlers"
	"github.com/jeb-maker/revues/internal/web/templates"
	webassets "github.com/jeb-maker/revues/web"
)

// NewRouter builds the HTTP handler tree for the application.
func NewRouter() (http.Handler, error) {
	tpl, err := templates.Parse()
	if err != nil {
		return nil, fmt.Errorf("load templates: %w", err)
	}

	staticFS, err := fs.Sub(webassets.Static, "static")
	if err != nil {
		return nil, fmt.Errorf("static assets: %w", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", handlers.Health)
	r.Get("/", (&handlers.Home{Templates: tpl}).ServeHTTP)
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	return r, nil
}
