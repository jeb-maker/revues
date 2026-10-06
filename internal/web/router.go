package web

import (
	"database/sql"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	apiv1 "github.com/jeb-maker/revues/internal/api/v1"
	"github.com/jeb-maker/revues/internal/attachments"
	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	adminintegrations "github.com/jeb-maker/revues/internal/features/admin/integrations"
	adminsettings "github.com/jeb-maker/revues/internal/features/admin/settings"
	authfeature "github.com/jeb-maker/revues/internal/features/auth"
	"github.com/jeb-maker/revues/internal/features/checklisttemplates"
	"github.com/jeb-maker/revues/internal/features/organizations"
	"github.com/jeb-maker/revues/internal/integrations/confluence"
	"github.com/jeb-maker/revues/internal/integrations/jira"
	"github.com/jeb-maker/revues/internal/integrations/webhooks"
	"github.com/jeb-maker/revues/internal/notifications"
	"github.com/jeb-maker/revues/internal/store"
	appmiddleware "github.com/jeb-maker/revues/internal/web/middleware"
	webassets "github.com/jeb-maker/revues/web"
)

// Deps holds runtime dependencies for HTTP routing.
type Deps struct {
	Config   config.Config
	DB       *sql.DB
	Notifier apiv1.RunNotifier // optional override (tests); default = notifications.Service
}

// NewRouter builds the HTTP handler tree for the application.
//
// NewRouter wires /healthz, auth (OAuth + API), OpenAPI handlers under /api/v1,
// optional /static vendor assets, and the SPA (frontend/build or stub).
func NewRouter(deps Deps) (http.Handler, *notifications.Service, *webhooks.Dispatcher, error) {
	staticFS, err := fs.Sub(webassets.Static, "static")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("static assets: %w", err)
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
	authSvc := &authfeature.Service{
		Store:    st,
		Sessions: sessions,
		GitHub:   github,
		Config:   deps.Config,
	}
	oauthHandlers := &authfeature.OAuthHandlers{Service: authSvc}
	templatesSvc := &checklisttemplates.Service{Store: st}
	orgSvc := &organizations.Service{
		Store:         st,
		Sessions:      sessions,
		SecureCookies: deps.Config.SecureCookies(),
	}

	adminSMTPKey, err := deps.Config.EncryptionKeyBytes()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("encryption key: %w", err)
	}
	settingsSvc := &adminsettings.SettingsService{
		Store:         st,
		EncryptionKey: adminSMTPKey,
	}
	jiraSvc := &jira.Service{Store: st, EncryptionKey: adminSMTPKey}
	confluenceSvc := &confluence.Service{Store: st, EncryptionKey: adminSMTPKey}
	integrationsSvc := &adminintegrations.IntegrationsService{
		Settings:   settingsSvc,
		Jira:       jiraSvc,
		Confluence: confluenceSvc,
	}
	attachmentsSvc := &attachments.Service{Store: st, Dir: deps.Config.AttachmentsDir}
	notificationsSvc := &notifications.Service{
		Store:    st,
		Settings: settingsSvc,
		BaseURL:  deps.Config.BaseURL,
	}
	webhookDispatcher := &webhooks.Dispatcher{
		Settings: settingsSvc,
		Store:    st,
		Runs:     st,
		DevMode:  deps.Config.Env == "development",
	}

	// notificationsSvc is always a non-nil pointer here, so the interface is never a typed nil
	// (which would defeat `s.Notifier != nil` in handlers).
	var runNotifier apiv1.RunNotifier = notificationsSvc
	if deps.Notifier != nil {
		runNotifier = deps.Notifier
	}

	apiServer := apiv1.NewServer(
		authSvc, orgSvc, templatesSvc, st, deps.Config, sessions, webhookDispatcher, runNotifier,
		settingsSvc, integrationsSvc, attachmentsSvc,
	)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(appmiddleware.SecurityHeaders)
	r.Use(appmiddleware.CapturePeerAddr)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(DevNoCache(deps.Config.Env))
	r.Use(appmiddleware.LoadUser(st))
	r.Use(appmiddleware.EnsureDevAuth(st, sessions, deps.Config.DevAuthEnabled(), deps.Config.DevAuthEmail))
	r.Use(appmiddleware.LoadActiveOrganization(st))
	r.Use(appmiddleware.CSRF(deps.Config.SessionSecret))

	r.Get("/healthz", Health)
	r.Handle("/static/*", http.StripPrefix("/static/", StaticHandler(http.FileServer(http.FS(staticFS)), deps.Config.Env)))

	authLimit := appmiddleware.RateLimit(appmiddleware.RateLimitConfig{Max: 30, Window: time.Minute})
	r.With(authLimit).Get("/auth/github/start", oauthHandlers.StartGitHub)
	r.With(authLimit).Get("/auth/github/callback", oauthHandlers.Callback)
	r.With(authLimit).Post("/auth/dev/login", oauthHandlers.DevLogin)

	r.Route("/api/v1", func(r chi.Router) {
		// Password login/register share the OAuth bucket size (credential stuffing).
		// chi keeps the full URL.Path inside Route middleware (verified).
		r.Use(appmiddleware.RateLimitPaths(
			appmiddleware.RateLimitConfig{Max: 30, Window: time.Minute},
			"/api/v1/auth/login",
			"/api/v1/auth/register",
		))
		apiv1.HandlerFromMux(apiServer, r)
	})

	spa := spaHandler(spaBuildDir())
	r.Get("/", spa)
	r.NotFound(spa)

	return r, notificationsSvc, webhookDispatcher, nil
}

func spaBuildDir() string {
	// Explicit override: if set, never fall back to the default path.
	if dir := os.Getenv("REVUES_SPA_DIR"); dir != "" {
		if spaIndexExists(dir) {
			return dir
		}
		return ""
	}
	if spaIndexExists("frontend/build") {
		return "frontend/build"
	}
	return ""
}

func spaIndexExists(dir string) bool {
	st, err := os.Stat(filepath.Join(dir, "index.html"))
	return err == nil && !st.IsDir()
}

func spaHandler(root string) http.HandlerFunc {
	if root == "" {
		return spaStub
	}
	return func(w http.ResponseWriter, r *http.Request) {
		rel := strings.TrimPrefix(r.URL.Path, "/")
		if rel != "" {
			full := filepath.Join(root, filepath.Clean(rel))
			if st, err := os.Stat(full); err == nil && !st.IsDir() {
				http.ServeFile(w, r, full)
				return
			}
		}
		http.ServeFile(w, r, filepath.Join(root, "index.html"))
	}
}

// spaStub answers when no SvelteKit build is available. It is a deployment
// error, so it returns 503: a 200 would hide the outage from Docker
// healthchecks, deploy scripts and HTTP monitoring (the API stays up).
func spaStub(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Retry-After", "60")
	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = w.Write([]byte(`<!doctype html>
<html lang="fr">
<head><meta charset="utf-8"><title>Revues</title></head>
<body>
  <h1>Revues</h1>
  <p>SPA SvelteKit non construite. Depuis la racine du dépôt&nbsp;:</p>
  <pre>cd frontend &amp;&amp; npm ci &amp;&amp; npm run build</pre>
  <p>Puis relancer <code>go run ./cmd/revues</code> (sert <code>frontend/build</code>),
  ou définir <code>REVUES_SPA_DIR</code>. Voir <code>docs/FRONTEND.md</code>.</p>
  <p><a href="/healthz">/healthz</a> · <a href="/login">/login</a></p>
</body>
</html>`))
}
