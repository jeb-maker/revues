package integrations

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/features/admin/settings"
	"github.com/jeb-maker/revues/internal/integrations/jira"
	"github.com/jeb-maker/revues/internal/integrations/notion"
	"github.com/jeb-maker/revues/internal/web/middleware"
	"github.com/jeb-maker/revues/internal/web/templates"
)

// Deps holds dependencies for admin integrations handlers.
type Deps struct {
	Templates     *template.Template
	Store         AdminStore
	SessionSecret string
}

func (d *Deps) PageData(r *http.Request, title string) templates.PageData {
	data := templates.PageData{Title: title}
	if user, ok := middleware.UserFromContext(r.Context()); ok {
		data.User = user
		if token := middleware.SessionTokenFromContext(r); token != "" {
			data.CSRFToken = auth.CSRFToken(token, d.SessionSecret)
		}
	}
	templates.ApplyHeaderFromContext(r, &data)
	return data
}

// AdminIntegrations shows the unified integrations overview.
type AdminIntegrations struct {
	Deps
	EncryptionKey []byte
}

// Show renders integration status and links to configuration pages.
func (h *AdminIntegrations) Show(w http.ResponseWriter, r *http.Request) {
	data := h.pageData(r)

	overview, err := h.service().Overview(r.Context())
	if err != nil {
		slog.Error("load integrations overview", "err", err)
		data.Error = "Impossible de charger le statut des intégrations."
	} else {
		data.Integrations = make([]templates.AdminIntegrationRow, 0, len(overview.Items))
		for _, item := range overview.Items {
			data.Integrations = append(data.Integrations, templates.AdminIntegrationRow{
				Name:        item.Name,
				Description: item.Description,
				Enabled:     item.Enabled,
				ConfigPath:  item.ConfigPath,
			})
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "admin_integrations", data); err != nil {
		slog.Error("render admin integrations", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *AdminIntegrations) pageData(r *http.Request) templates.AdminIntegrationsData {
	data := templates.AdminIntegrationsData{
		PageData: templates.ApplyPageMeta(h.PageData(r, ""), templates.BCAdminIntegrations()),
	}
	data.ActiveTab = "org"
	data.AdminSection = "integrations"
	return data
}

func (h *AdminIntegrations) service() *IntegrationsService {
	return &IntegrationsService{
		Settings: &settings.SettingsService{
			Store:         h.Store,
			EncryptionKey: h.EncryptionKey,
		},
		Jira:   &jira.Service{Store: h.Store, EncryptionKey: h.EncryptionKey},
		Notion: &notion.Service{Store: h.Store, EncryptionKey: h.EncryptionKey},
	}
}
