package web

import (
	"database/sql"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/jeb-maker/revues/internal/admin"
	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/features/projects"
	"github.com/jeb-maker/revues/internal/integrations/webhooks"
	"github.com/jeb-maker/revues/internal/notifications"
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
func NewRouter(deps Deps) (http.Handler, *notifications.Service, error) {
	tpl, err := templates.Parse()
	if err != nil {
		return nil, nil, fmt.Errorf("load templates: %w", err)
	}

	staticFS, err := fs.Sub(webassets.Static, "static")
	if err != nil {
		return nil, nil, fmt.Errorf("static assets: %w", err)
	}

	st := store.New(deps.DB)
	handlerDeps := handlers.Deps{
		Templates:     tpl,
		Store:         st,
		SessionSecret: deps.Config.SessionSecret,
	}
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
	adminUsers := &handlers.AdminUsers{Deps: handlerDeps}
	adminSMTPKey, err := deps.Config.EncryptionKeyBytes()
	if err != nil {
		return nil, nil, fmt.Errorf("encryption key: %w", err)
	}
	settingsSvc := &admin.SettingsService{
		Store:         st,
		EncryptionKey: adminSMTPKey,
	}
	notificationsSvc := &notifications.Service{
		Store:    st,
		Settings: settingsSvc,
		BaseURL:  deps.Config.BaseURL,
	}
	webhookDispatcher := &webhooks.Dispatcher{Settings: settingsSvc, Store: st, Runs: st, DevMode: deps.Config.Env == "development"}
	adminWebhooks := &handlers.AdminWebhooks{Deps: handlerDeps, EncryptionKey: adminSMTPKey, Webhooks: webhookDispatcher}
	adminSMTP := &handlers.AdminSMTP{Deps: handlerDeps, EncryptionKey: adminSMTPKey}
	adminJira := &handlers.AdminJira{Deps: handlerDeps, EncryptionKey: adminSMTPKey}
	adminNotion := &handlers.AdminNotion{Deps: handlerDeps, EncryptionKey: adminSMTPKey}
	adminIntegrations := &handlers.AdminIntegrations{Deps: handlerDeps, EncryptionKey: adminSMTPKey}
	projectsHandler := &projects.Projects{Deps: projects.Deps{
		Templates:     tpl,
		Store:         st,
		SessionSecret: deps.Config.SessionSecret,
	}}
	checklistTemplates := &handlers.ChecklistTemplates{Deps: handlerDeps, EncryptionKey: adminSMTPKey}
	runsHandler := &handlers.Runs{
		Deps:           handlerDeps,
		EncryptionKey:  adminSMTPKey,
		AttachmentsDir: deps.Config.AttachmentsDir,
		BaseURL:        deps.Config.BaseURL,
		Webhooks:       webhookDispatcher,
		Notifications:  notificationsSvc,
	}
	myTasks := &handlers.MyTasks{Deps: handlerDeps}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(appmiddleware.LoadUser(st))
	r.Use(appmiddleware.CSRF(deps.Config.SessionSecret))

	r.Get("/healthz", handlers.Health)
	r.Get("/", (&handlers.Home{Deps: handlerDeps}).ServeHTTP)
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	r.Get("/login", authHandler.Login)
	r.Get("/auth/github/start", authHandler.StartGitHub)
	r.Get("/auth/github/callback", authHandler.Callback)
	r.Post("/logout", authHandler.Logout)

	r.Group(func(r chi.Router) {
		r.Use(appmiddleware.RequireAuth)
		r.Get("/projects", projectsHandler.List)
		r.Get("/projects/new", projectsHandler.NewForm)
		r.Post("/projects", projectsHandler.Create)
		r.Get("/projects/{id}", projectsHandler.Show)
		r.Get("/projects/{id}/edit", projectsHandler.EditForm)
		r.Post("/projects/{id}", projectsHandler.Update)
		r.Post("/projects/{id}/archive", projectsHandler.Archive)
		r.Post("/projects/{id}/members", projectsHandler.AddMember)
		r.Post("/projects/{id}/members/remove", projectsHandler.RemoveMember)
		r.Get("/projects/{id}/templates", checklistTemplates.List)
		r.Get("/projects/{id}/templates/new", checklistTemplates.NewForm)
		r.Get("/projects/{id}/templates/notion-import", checklistTemplates.NotionImportForm)
		r.Post("/projects/{id}/templates/notion-import", checklistTemplates.NotionImport)
		r.Post("/projects/{id}/templates", checklistTemplates.Create)
		r.Get("/projects/{id}/templates/{tid}", checklistTemplates.Show)
		r.Get("/projects/{id}/templates/{tid}/edit", checklistTemplates.EditForm)
		r.Post("/projects/{id}/templates/{tid}", checklistTemplates.Save)
		r.Post("/projects/{id}/templates/{tid}/archive", checklistTemplates.Archive)
		r.Post("/projects/{id}/runs", runsHandler.Create)
		r.Get("/runs/new", runsHandler.WizardProjects)
		r.Get("/runs/new/projects/{id}", runsHandler.WizardTemplates)
		r.Get("/runs/new/projects/{id}/templates/{tid}", runsHandler.WizardLaunch)
		r.Get("/runs/{id}", runsHandler.Show)
		r.Get("/runs/{id}/export.csv", runsHandler.ExportCSV)
		r.Post("/runs/{id}/export/notion", runsHandler.ExportNotion)
		r.Get("/runs/{id}/items/{itemId}", runsHandler.ShowItem)
		r.Post("/runs/{id}/items/{itemId}", runsHandler.UpdateItem)
		r.Post("/runs/{id}/items/{itemId}/assign", runsHandler.AssignItem)
		r.Post("/runs/{id}/items/{itemId}/jira-link", runsHandler.LinkJiraItem)
		r.Post("/runs/{id}/items/{itemId}/jira-create", runsHandler.CreateJiraItem)
		r.Post("/runs/{id}/items/{itemId}/attachment", runsHandler.UploadAttachment)
		r.Get("/attachments/{id}", runsHandler.DownloadAttachment)
		r.Post("/runs/{id}/start", runsHandler.Start)
		r.Post("/runs/{id}/complete", runsHandler.Complete)
		r.Get("/mes-taches", myTasks.List)
		r.Get("/modeles", checklistTemplates.IndexAll)
	})

	r.Group(func(r chi.Router) {
		r.Use(appmiddleware.RequireAuth)
		r.Use(appmiddleware.RequireRole(auth.RoleAdmin))
		r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/admin/users", http.StatusFound)
		})
		r.Get("/admin/users", adminUsers.List)
		r.Post("/admin/users", adminUsers.Add)
		r.Post("/admin/users/remove", adminUsers.Remove)
		r.Get("/admin/integrations", adminIntegrations.Show)
		r.Get("/admin/settings/smtp", adminSMTP.Show)
		r.Post("/admin/settings/smtp", adminSMTP.Save)
		r.Get("/admin/settings/webhooks", adminWebhooks.Show)
		r.Post("/admin/settings/webhooks", adminWebhooks.Save)
		r.Get("/admin/integrations/jira", adminJira.Show)
		r.Post("/admin/integrations/jira", adminJira.Save)
		r.Get("/admin/integrations/notion", adminNotion.Show)
		r.Post("/admin/integrations/notion", adminNotion.Save)
	})

	return r, notificationsSvc, nil
}
