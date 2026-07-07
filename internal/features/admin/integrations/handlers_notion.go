package integrations

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/jeb-maker/revues/internal/integrations/notion"
	"github.com/jeb-maker/revues/internal/web/handlers"
	"github.com/jeb-maker/revues/internal/web/templates"
)

// AdminNotion manages encrypted Notion integration settings.
type AdminNotion struct {
	handlers.Deps
	EncryptionKey []byte
	NotionClient  *notion.Client
}

// Show renders the Notion configuration form.
func (h *AdminNotion) Show(w http.ResponseWriter, r *http.Request) {
	data := templates.AdminNotionData{
		PageData:   h.PageData(r, "Configuration Notion"),
		CanEncrypt: len(h.EncryptionKey) > 0,
		Message:    r.URL.Query().Get("msg"),
	}
	data.AdminSection = "notion"
	if cfg, ok, err := h.service().Load(r.Context()); err != nil {
		data.Error = "Impossible de charger la configuration Notion."
	} else if ok {
		data.WorkspaceName = cfg.WorkspaceName
		data.DefaultDatabaseID = cfg.DefaultDatabaseID
		data.HasAPIToken = notion.HasSecret(cfg.APIToken)
		data.Configured = cfg.Configured()
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "admin_notion", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Save stores Notion settings or tests the connection.
func (h *AdminNotion) Save(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(r.FormValue("action")) == "test" {
		h.testConnection(w, r)
		return
	}
	if len(h.EncryptionKey) == 0 {
		h.renderError(w, r, "REVUES_ENCRYPTION_KEY est requis pour enregistrer la configuration Notion.")
		return
	}
	cfg := notion.Config{
		APIToken:          r.FormValue("api_token"),
		WorkspaceName:     strings.TrimSpace(r.FormValue("workspace_name")),
		DefaultDatabaseID: strings.TrimSpace(r.FormValue("default_database_id")),
	}
	if current, hasCurrent, err := h.service().Load(r.Context()); err == nil && hasCurrent {
		cfg.APIToken = notion.MergeSecret(current.APIToken, cfg.APIToken)
	}
	if err := h.service().Save(r.Context(), cfg); err != nil {
		h.renderError(w, r, err.Error())
		return
	}
	http.Redirect(w, r, "/admin/integrations/notion?msg=Configuration+Notion+enregistr%C3%A9e", http.StatusSeeOther)
}

func (h *AdminNotion) testConnection(w http.ResponseWriter, r *http.Request) {
	cfg, ok, err := h.service().Load(r.Context())
	if err != nil || !ok || !cfg.Configured() {
		h.renderError(w, r, "Configurez et enregistrez Notion avant de tester la connexion.")
		return
	}
	info, err := h.client().TestConnection(r.Context(), cfg)
	if err != nil {
		h.renderError(w, r, "Échec du test de connexion Notion. Vérifiez le jeton d'intégration.")
		return
	}
	msg := "Connexion Notion réussie"
	if info.WorkspaceName != "" {
		msg += " (" + info.WorkspaceName + ")"
	}
	http.Redirect(w, r, "/admin/integrations/notion?msg="+url.QueryEscape(msg), http.StatusSeeOther)
}

func (h *AdminNotion) renderError(w http.ResponseWriter, r *http.Request, msg string) {
	data := templates.AdminNotionData{
		PageData:   h.PageData(r, "Configuration Notion"),
		CanEncrypt: len(h.EncryptionKey) > 0,
		Error:      msg,
	}
	data.AdminSection = "notion"
	w.WriteHeader(http.StatusBadRequest)
	_ = h.Templates.ExecuteTemplate(w, "admin_notion", data)
}

func (h *AdminNotion) service() *notion.Service {
	return &notion.Service{Store: h.Store, EncryptionKey: h.EncryptionKey}
}

func (h *AdminNotion) client() *notion.Client {
	if h.NotionClient != nil {
		return h.NotionClient
	}
	return &notion.Client{}
}
