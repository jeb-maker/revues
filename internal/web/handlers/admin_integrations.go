package handlers

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/jeb-maker/revues/internal/admin"
	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/integrations/jira"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/middleware"
	"github.com/jeb-maker/revues/internal/web/templates"
)

// AdminIntegrations shows the unified integrations overview.
type AdminIntegrations struct {
	Templates     *template.Template
	Store         *store.Store
	SessionSecret string
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
		PageData: templates.PageData{Title: "Intégrations"},
	}
	if user, ok := middleware.UserFromContext(r.Context()); ok {
		data.User = user
		if token := middleware.SessionTokenFromContext(r); token != "" {
			data.CSRFToken = auth.CSRFToken(token, h.SessionSecret)
		}
	}
	return data
}

func (h *AdminIntegrations) service() *admin.IntegrationsService {
	return &admin.IntegrationsService{
		Settings: &admin.SettingsService{
			Store:         h.Store,
			EncryptionKey: h.EncryptionKey,
		},
		Jira: &jira.Service{
			Store:         h.Store,
			EncryptionKey: h.EncryptionKey,
		},
	}
}
