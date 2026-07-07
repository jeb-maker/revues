package webhooks

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jeb-maker/revues/internal/features/admin/settings"
	whdispatch "github.com/jeb-maker/revues/internal/integrations/webhooks"
	"github.com/jeb-maker/revues/internal/web/handlerdeps"
	"github.com/jeb-maker/revues/internal/web/templates"
)

// AdminWebhooks manages encrypted webhook notification settings.
type AdminWebhooks struct {
	handlerdeps.HandlerDeps
	EncryptionKey []byte
	Webhooks      *whdispatch.Dispatcher
}

// Show renders the webhook configuration form.
func (h *AdminWebhooks) Show(w http.ResponseWriter, r *http.Request) {
	data := h.pageData(r)
	data.Message = r.URL.Query().Get("msg")
	if cfg, ok, err := h.settings().LoadWebhooks(r.Context()); err != nil {
		slog.Error("load webhooks settings", "err", err)
		data.Error = "Impossible de charger la configuration webhooks."
	} else if ok {
		data.URLsText = strings.Join(cfg.URLs, "\n")
		data.HasSecret = cfg.Secret != ""
		data.ReviewCompleted = cfg.ReviewCompleted
		data.ReviewItemNOK = cfg.ReviewItemNOK
		data.Configured = cfg.Enabled()
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "admin_webhooks", data); err != nil {
		slog.Error("render admin webhooks", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Save stores webhook settings or sends a test delivery.
func (h *AdminWebhooks) Save(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(r.FormValue("action")) == "test" {
		h.sendTest(w, r)
		return
	}
	h.saveConfig(w, r)
}

func (h *AdminWebhooks) saveConfig(w http.ResponseWriter, r *http.Request) {
	if len(h.EncryptionKey) == 0 {
		h.renderForm(w, r, templates.AdminWebhooksData{Error: "REVUES_ENCRYPTION_KEY est requis pour enregistrer la configuration webhooks."})
		return
	}
	cfg := settings.WebhookConfig{URLs: settings.ParseWebhookURLs(r.FormValue("urls")), ReviewCompleted: r.FormValue("review_completed") == "on", ReviewItemNOK: r.FormValue("review_item_nok") == "on"}
	current, hasCurrent, err := h.settings().LoadWebhooks(r.Context())
	if err != nil {
		h.renderForm(w, r, templates.AdminWebhooksData{Error: "Impossible de charger la configuration existante."})
		return
	}
	if hasCurrent {
		cfg.Secret = settings.MergeWebhookSecret(current, r.FormValue("secret"))
	} else {
		cfg.Secret = r.FormValue("secret")
	}
	if err := h.settings().SaveWebhooks(r.Context(), cfg); err != nil {
		msg := err.Error()
		if errors.Is(err, settings.ErrEncryptionNotConfigured) {
			msg = "REVUES_ENCRYPTION_KEY est requis pour enregistrer la configuration webhooks."
		}
		h.renderForm(w, r, templates.AdminWebhooksData{URLsText: strings.Join(cfg.URLs, "\n"), ReviewCompleted: cfg.ReviewCompleted, ReviewItemNOK: cfg.ReviewItemNOK, Error: msg})
		return
	}
	http.Redirect(w, r, "/admin/settings/webhooks?msg=Configuration+webhooks+enregistr%C3%A9e", http.StatusSeeOther)
}

func (h *AdminWebhooks) sendTest(w http.ResponseWriter, r *http.Request) {
	if h.Webhooks == nil || h.Webhooks.SendTest(r.Context()) != nil {
		h.renderForm(w, r, templates.AdminWebhooksData{Error: "Échec de l'envoi du webhook de test. Consultez les logs serveur."})
		return
	}
	http.Redirect(w, r, "/admin/settings/webhooks?msg=Webhook+de+test+envoy%C3%A9", http.StatusSeeOther)
}

func (h *AdminWebhooks) renderForm(w http.ResponseWriter, r *http.Request, partial templates.AdminWebhooksData) {
	data := h.pageData(r)
	data.URLsText, data.HasSecret, data.ReviewCompleted, data.ReviewItemNOK, data.Configured = partial.URLsText, partial.HasSecret, partial.ReviewCompleted, partial.ReviewItemNOK, partial.Configured
	data.Message, data.Error = partial.Message, partial.Error
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	status := http.StatusOK
	if data.Error != "" {
		status = http.StatusBadRequest
	}
	w.WriteHeader(status)
	_ = h.Templates.ExecuteTemplate(w, "admin_webhooks", data)
}

func (h *AdminWebhooks) pageData(r *http.Request) templates.AdminWebhooksData {
	data := templates.AdminWebhooksData{
		PageData:   h.PageData(r, "Configuration webhooks"),
		CanEncrypt: len(h.EncryptionKey) > 0,
	}
	data.AdminSection = "webhooks"
	return data
}

func (h *AdminWebhooks) settings() *settings.SettingsService {
	return &settings.SettingsService{Store: h.Store, EncryptionKey: h.EncryptionKey}
}
