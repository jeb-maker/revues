package web

import (
	"database/sql"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/handlers"
	appmiddleware "github.com/jeb-maker/revues/internal/web/middleware"
	"github.com/jeb-maker/revues/internal/web/templates"
	webassets "github.com/jeb-maker/revues/web"
)

// Deps holds runtime dependencies for HTTP routing.
type Deps struct {
	Config config.Config
	DB     *sql.DB
}

// NewRouter builds the HTTP handler tree for the application.
func NewRouter(deps Deps) (http.Handler, error) {
	tpl, err := templates.Parse()
	if err != nil {
		return nil, fmt.Errorf("load templates: %w", err)
	}

	staticFS, err := fs.Sub(webassets.Static, "static")
	if err != nil {
		return nil, fmt.Errorf("static assets: %w", err)
	}

	st := store.New(deps.DB)
	sessions := &auth.SessionManager{
		Store:         st,
		SessionSecret: deps.Config.SessionSecret,
		SecureCookies: deps.Config.SecureCookies(),
	}
	github := &auth.GitHubOAuth{
		ClientID:     deps.Config.GitHubClientID,
		ClientSecret: deps.Config.GitHubClientSecret,
		BaseURL:      deps.Config.BaseURL,
	}

	authHandler := &handlers.Auth{
		Templates: tpl,
		Store:     st,
		Sessions:  sessions,
		GitHub:    github,
		Config:    deps.Config,
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(appmiddleware.LoadUser(st))
	r.Use(appmiddleware.CSRF(deps.Config.SessionSecret))

	r.Get("/healthz", handlers.Health)
	r.Get("/", (&handlers.Home{Templates: tpl, SessionSecret: deps.Config.SessionSecret}).ServeHTTP)
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	r.Get("/login", authHandler.Login)
	r.Get("/auth/github/start", authHandler.StartGitHub)
	r.Get("/auth/github/callback", authHandler.Callback)
	r.Post("/logout", authHandler.Logout)

	r.Group(func(r chi.Router) {
		r.Use(appmiddleware.RequireAuth)
		r.With(appmiddleware.RequireRole(auth.RoleAdmin)).Get("/admin", handlers.AdminStub)
	})

	return r, nil
}
