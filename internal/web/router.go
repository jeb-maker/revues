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
	adminintegrations "github.com/jeb-maker/revues/internal/features/admin/integrations"
	adminsettings "github.com/jeb-maker/revues/internal/features/admin/settings"
	adminsmtp "github.com/jeb-maker/revues/internal/features/admin/smtp"
	adminteams "github.com/jeb-maker/revues/internal/features/admin/teams"
	adminusers "github.com/jeb-maker/revues/internal/features/admin/users"
	adminwebhooks "github.com/jeb-maker/revues/internal/features/admin/webhooks"
	authhandler "github.com/jeb-maker/revues/internal/features/auth"
	"github.com/jeb-maker/revues/internal/features/checklisttemplates"
	home "github.com/jeb-maker/revues/internal/features/home"
	mytasks "github.com/jeb-maker/revues/internal/features/mytasks"
	"github.com/jeb-maker/revues/internal/features/organizations"
	runs "github.com/jeb-maker/revues/internal/features/runs"
	"github.com/jeb-maker/revues/internal/features/subjects"
	"github.com/jeb-maker/revues/internal/integrations/webhooks"
	"github.com/jeb-maker/revues/internal/notifications"
	"github.com/jeb-maker/revues/internal/store"
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
	staticFS, err := fs.Sub(webassets.Static, "static")
	if err != nil {
		return nil, nil, fmt.Errorf("static assets: %w", err)
	}

	assetVersion, err := StaticAssetVersion(staticFS)
	if err != nil {
		return nil, nil, fmt.Errorf("static asset version: %w", err)
	}

	tpl, err := templates.Parse(assetVersion)
	if err != nil {
		return nil, nil, fmt.Errorf("load templates: %w", err)
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

	authHandler := &authhandler.Auth{
		Templates: tpl,
		Store:     st,
		Sessions:  sessions,
		GitHub:    github,
		Config:    deps.Config,
	}
	adminSMTPKey, err := deps.Config.EncryptionKeyBytes()
	if err != nil {
		return nil, nil, fmt.Errorf("encryption key: %w", err)
	}
	settingsSvc := &adminsettings.SettingsService{
		Store:         st,
		EncryptionKey: adminSMTPKey,
	}
	notificationsSvc := &notifications.Service{
		Store:    st,
		Settings: settingsSvc,
		BaseURL:  deps.Config.BaseURL,
	}
	webhookDispatcher := &webhooks.Dispatcher{Settings: settingsSvc, Store: st, Runs: st, DevMode: deps.Config.Env == "development"}
	adminUsers := &adminusers.AdminUsers{Deps: adminusers.Deps{
		Templates:     tpl,
		Store:         st,
		SessionSecret: deps.Config.SessionSecret,
	}}
	adminTeams := &adminteams.AdminTeams{Deps: adminteams.Deps{
		Templates:     tpl,
		Store:         st,
		SessionSecret: deps.Config.SessionSecret,
	}}
	adminWebhooks := &adminwebhooks.AdminWebhooks{
		Deps: adminwebhooks.Deps{
			Templates:     tpl,
			Store:         st,
			SessionSecret: deps.Config.SessionSecret,
		},
		EncryptionKey: adminSMTPKey,
		Webhooks:      webhookDispatcher,
	}
	adminSMTP := &adminsmtp.AdminSMTP{
		Deps: adminsmtp.Deps{
			Templates:     tpl,
			Store:         st,
			SessionSecret: deps.Config.SessionSecret,
		},
		EncryptionKey: adminSMTPKey,
	}
	adminJira := &adminintegrations.AdminJira{
		Deps: adminintegrations.Deps{
			Templates:     tpl,
			Store:         st,
			SessionSecret: deps.Config.SessionSecret,
		},
		EncryptionKey: adminSMTPKey,
	}
	adminNotion := &adminintegrations.AdminNotion{
		Deps: adminintegrations.Deps{
			Templates:     tpl,
			Store:         st,
			SessionSecret: deps.Config.SessionSecret,
		},
		EncryptionKey: adminSMTPKey,
	}
	adminIntegrations := &adminintegrations.AdminIntegrations{
		Deps: adminintegrations.Deps{
			Templates:     tpl,
			Store:         st,
			SessionSecret: deps.Config.SessionSecret,
		},
		EncryptionKey: adminSMTPKey,
	}
	subjectsHandler := &subjects.Subjects{Deps: subjects.Deps{
		Templates:     tpl,
		Store:         st,
		SessionSecret: deps.Config.SessionSecret,
	}}
	checklistTemplates := &checklisttemplates.ChecklistTemplates{Deps: checklisttemplates.Deps{
		Templates:     tpl,
		Store:         st,
		SessionSecret: deps.Config.SessionSecret,
	}, EncryptionKey: adminSMTPKey}
	runsHandler := &runs.Runs{
		Deps: runs.Deps{
			Templates:     tpl,
			Store:         st,
			SessionSecret: deps.Config.SessionSecret,
		},
		EncryptionKey:  adminSMTPKey,
		AttachmentsDir: deps.Config.AttachmentsDir,
		BaseURL:        deps.Config.BaseURL,
		Webhooks:       webhookDispatcher,
		Notifications:  notificationsSvc,
	}
	myTasks := &mytasks.MyTasks{Deps: mytasks.Deps{
		Templates:     tpl,
		Store:         st,
		SessionSecret: deps.Config.SessionSecret,
	}}
	orgsHandler := &organizations.Organizations{Deps: organizations.Deps{
		Templates:     tpl,
		Store:         st,
		Sessions:      sessions,
		SessionSecret: deps.Config.SessionSecret,
		SecureCookies: deps.Config.SecureCookies(),
	}}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(appmiddleware.CapturePeerAddr) // before RealIP — DevAuth must see true peer
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(DevNoCache(deps.Config.Env))
	r.Use(appmiddleware.LoadUser(st))
	r.Use(appmiddleware.EnsureDevAuth(st, sessions, deps.Config.DevAuthEnabled(), deps.Config.DevAuthEmail))
	r.Use(appmiddleware.LoadActiveOrganization(st))
	r.Use(appmiddleware.LoadHeaderData(st))
	r.Use(appmiddleware.CSRF(deps.Config.SessionSecret))

	r.Get("/healthz", Health)
	r.Get("/", (&home.Home{Deps: home.Deps{
		Templates:     tpl,
		Store:         st,
		SessionSecret: deps.Config.SessionSecret,
	}}).ServeHTTP)
	r.Get("/sw.js", ServeServiceWorkerKill(staticFS))
	r.Handle("/static/*", http.StripPrefix("/static/", StaticHandler(http.FileServer(http.FS(staticFS)), deps.Config.Env)))

	r.Get("/login", authHandler.Login)
	r.Get("/auth/github/start", authHandler.StartGitHub)
	r.Get("/auth/github/callback", authHandler.Callback)
	r.Post("/logout", authHandler.Logout)

	r.Group(func(r chi.Router) {
		r.Use(appmiddleware.RequireAuth)
		r.Get("/org/new", orgsHandler.NewForm)
		r.Post("/org/new", orgsHandler.Create)
		r.Get("/org/select", orgsHandler.SelectForm)
		r.Post("/org/select", orgsHandler.Select)
		r.Post("/org/invitations/{id}/accept", orgsHandler.AcceptInvitation)
	})

	r.Group(func(r chi.Router) {
		r.Use(appmiddleware.RequireAuth)
		r.Post("/org/switch", orgsHandler.Switch)
		r.Get("/revues", runsHandler.List)
		r.Get("/subjects", subjectsHandler.List)
		r.Get("/subjects/new", subjectsHandler.NewForm)
		r.Post("/subjects", subjectsHandler.Create)
		r.Get("/subjects/{id}", subjectsHandler.Show)
		r.Get("/subjects/{id}/edit", subjectsHandler.EditForm)
		r.Post("/subjects/{id}", subjectsHandler.Update)
		r.Post("/subjects/{id}/archive", subjectsHandler.Archive)
		r.Get("/subjects/{id}/teams/preview", subjectsHandler.PreviewTeam)
		r.Post("/subjects/{id}/teams", subjectsHandler.AddTeam)
		r.Post("/subjects/{id}/teams/remove", subjectsHandler.RemoveTeam)
		r.Get("/subjects/{id}/modeles", checklistTemplates.List)
		r.Post("/subjects/{id}/revues", runsHandler.Create)
		r.Get("/revues/nouvelle", subjectsHandler.WizardNouvelle)
		r.Post("/revues/nouvelle", subjectsHandler.WizardNouvelleCreate)
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
		r.Get("/modeles/new", checklistTemplates.NewForm)
		r.Post("/modeles", checklistTemplates.Create)
		r.Get("/modeles/notion-import", checklistTemplates.NotionImportForm)
		r.Post("/modeles/notion-import", checklistTemplates.NotionImport)
		r.Get("/modeles/{tid}", checklistTemplates.Show)
		r.Get("/modeles/{tid}/edit", checklistTemplates.EditForm)
		r.Post("/modeles/{tid}", checklistTemplates.Save)
		r.Post("/modeles/{tid}/archive", checklistTemplates.Archive)
	})

	r.Group(func(r chi.Router) {
		r.Use(appmiddleware.RequireAuth)
		r.Use(appmiddleware.RequireOrgAdmin(st))
		r.Get("/admin", orgsHandler.AdminHub)
		r.Get("/admin/users", adminUsers.List)
		r.Post("/admin/users", adminUsers.Add)
		r.Post("/admin/users/remove", adminUsers.Remove)
		r.Get("/admin/teams", adminTeams.List)
		r.Post("/admin/teams", adminTeams.Create)
		r.Get("/admin/teams/{id}", adminTeams.Show)
		r.Post("/admin/teams/{id}/members", adminTeams.AddMember)
		r.Post("/admin/teams/{id}/members/remove", adminTeams.RemoveMember)
		r.Get("/admin/subjects", subjectsHandler.List)
		r.Get("/admin/subjects/new", subjectsHandler.NewForm)
		r.Post("/admin/subjects", subjectsHandler.Create)
		r.Get("/admin/subjects/{id}/edit", subjectsHandler.EditForm)
		r.Post("/admin/subjects/{id}", subjectsHandler.Update)
		r.Post("/admin/subjects/{id}/archive", subjectsHandler.Archive)
		r.Get("/admin/settings/labels", orgsHandler.SubjectLabelsShow)
		r.Post("/admin/settings/labels", orgsHandler.SubjectLabelsSave)
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
