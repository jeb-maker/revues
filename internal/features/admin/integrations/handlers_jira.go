package integrations

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jeb-maker/revues/internal/integrations/jira"
	"github.com/jeb-maker/revues/internal/web/templates"
)

// AdminJira manages encrypted Jira integration settings.
type AdminJira struct {
	Deps
	EncryptionKey []byte
	JiraClient    *jira.Client
}

// Show renders the Jira configuration form.
func (h *AdminJira) Show(w http.ResponseWriter, r *http.Request) {
	data := h.pageData(r)
	data.Message = r.URL.Query().Get("msg")

	if cfg, ok, err := h.service().Load(r.Context()); err != nil {
		slog.Error("load jira settings", "err", err)
		data.Error = "Impossible de charger la configuration Jira."
	} else if ok {
		data.InstanceType = cfg.InstanceType
		data.BaseURL = cfg.BaseURL
		data.Email = cfg.Email
		data.ProjectKey = cfg.ProjectKey
		data.IssueType = cfg.IssueType
		data.HasAPIToken = jira.HasSecret(cfg.APIToken)
		data.HasPAT = jira.HasSecret(cfg.PAT)
		data.Configured = cfg.Configured()
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "admin_jira", data); err != nil {
		slog.Error("render admin jira", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Save stores Jira settings or tests the connection.
func (h *AdminJira) Save(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	action := strings.TrimSpace(r.FormValue("action"))
	switch action {
	case "test":
		h.testConnection(w, r)
	default:
		h.saveConfig(w, r)
	}
}

func (h *AdminJira) saveConfig(w http.ResponseWriter, r *http.Request) {
	if len(h.EncryptionKey) == 0 {
		h.renderForm(w, r, templates.AdminJiraData{
			Error: "REVUES_ENCRYPTION_KEY est requis pour enregistrer la configuration Jira.",
		})
		return
	}

	cfg := h.configFromForm(r)

	current, hasCurrent, err := h.service().Load(r.Context())
	if err != nil {
		slog.Error("load jira for merge", "err", err)
		h.renderForm(w, r, templates.AdminJiraData{Error: "Impossible de charger la configuration existante."})
		return
	}
	if hasCurrent {
		cfg.APIToken = jira.MergeSecret(current.APIToken, r.FormValue("api_token"))
		cfg.PAT = jira.MergeSecret(current.PAT, r.FormValue("pat"))
	}

	if err := h.service().Save(r.Context(), cfg); err != nil {
		var msg string
		switch {
		case errors.Is(err, jira.ErrEncryptionNotConfigured):
			msg = "REVUES_ENCRYPTION_KEY est requis pour enregistrer la configuration Jira."
		default:
			msg = err.Error()
		}
		h.renderForm(w, r, templates.AdminJiraData{
			InstanceType: cfg.InstanceType,
			BaseURL:      cfg.BaseURL,
			Email:        cfg.Email,
			ProjectKey:   cfg.ProjectKey,
			IssueType:    cfg.IssueType,
			HasAPIToken:  cfg.APIToken != "",
			HasPAT:       cfg.PAT != "",
			Error:        msg,
		})
		return
	}

	http.Redirect(w, r, "/admin/integrations/jira?msg=Configuration+Jira+enregistr%C3%A9e", http.StatusSeeOther)
}

func (h *AdminJira) testConnection(w http.ResponseWriter, r *http.Request) {
	cfg, ok, err := h.service().Load(r.Context())
	if err != nil {
		slog.Error("load jira for test", "err", err)
		h.renderForm(w, r, templates.AdminJiraData{Error: "Impossible de charger la configuration Jira."})
		return
	}
	if !ok || !cfg.Configured() {
		h.renderForm(w, r, templates.AdminJiraData{Error: "Configurez et enregistrez Jira avant de tester la connexion."})
		return
	}

	client := h.jiraClient()
	if err := client.TestConnection(r.Context(), cfg); err != nil {
		slog.Error("jira test connection", "err", err)
		h.renderForm(w, r, templates.AdminJiraData{
			InstanceType: cfg.InstanceType,
			BaseURL:      cfg.BaseURL,
			Email:        cfg.Email,
			HasAPIToken:  jira.HasSecret(cfg.APIToken),
			HasPAT:       jira.HasSecret(cfg.PAT),
			Configured:   true,
			Error:        "Échec du test de connexion Jira. Vérifiez l'URL et les identifiants.",
		})
		return
	}

	http.Redirect(w, r, "/admin/integrations/jira?msg=Connexion+Jira+r%C3%A9ussie", http.StatusSeeOther)
}

func (h *AdminJira) configFromForm(r *http.Request) jira.Config {
	instanceType := strings.TrimSpace(r.FormValue("instance_type"))
	if instanceType != jira.InstanceServer {
		instanceType = jira.InstanceCloud
	}

	return jira.Config{
		InstanceType: instanceType,
		BaseURL:      strings.TrimSpace(r.FormValue("base_url")),
		Email:        strings.TrimSpace(r.FormValue("email")),
		APIToken:     r.FormValue("api_token"),
		PAT:          r.FormValue("pat"),
		ProjectKey:   strings.TrimSpace(r.FormValue("project_key")),
		IssueType:    strings.TrimSpace(r.FormValue("issue_type")),
	}
}

func (h *AdminJira) renderForm(w http.ResponseWriter, r *http.Request, partial templates.AdminJiraData) {
	data := h.pageData(r)
	data.InstanceType = partial.InstanceType
	data.BaseURL = partial.BaseURL
	data.Email = partial.Email
	data.ProjectKey = partial.ProjectKey
	data.IssueType = partial.IssueType
	data.HasAPIToken = partial.HasAPIToken
	data.HasPAT = partial.HasPAT
	data.Configured = partial.Configured
	data.Message = partial.Message
	data.Error = partial.Error

	if data.InstanceType == "" {
		data.InstanceType = jira.InstanceCloud
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	status := http.StatusOK
	if data.Error != "" {
		status = http.StatusBadRequest
	}
	w.WriteHeader(status)
	if err := h.Templates.ExecuteTemplate(w, "admin_jira", data); err != nil {
		slog.Error("render admin jira form", "err", err)
	}
}

func (h *AdminJira) pageData(r *http.Request) templates.AdminJiraData {
	data := templates.AdminJiraData{
		PageData:     templates.ApplyPageMeta(h.PageData(r, ""), templates.BCAdminJira()),
		InstanceType: jira.InstanceCloud,
		CanEncrypt:   len(h.EncryptionKey) > 0,
	}
	data.ActiveTab = "org"
	data.AdminSection = "jira"
	return data
}

func (h *AdminJira) service() *jira.Service {
	return &jira.Service{
		Store:         h.Store,
		EncryptionKey: h.EncryptionKey,
	}
}

func (h *AdminJira) jiraClient() *jira.Client {
	if h.JiraClient != nil {
		return h.JiraClient
	}
	return &jira.Client{}
}
