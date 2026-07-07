package smtp

import (
	"errors"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"

	"github.com/jeb-maker/revues/internal/features/admin/settings"
	"github.com/jeb-maker/revues/internal/notifications"
	"github.com/jeb-maker/revues/internal/web/handlers"
	"github.com/jeb-maker/revues/internal/web/middleware"
	"github.com/jeb-maker/revues/internal/web/templates"
)

// AdminSMTP manages encrypted SMTP relay settings.
type AdminSMTP struct {
	handlers.Deps
	EncryptionKey []byte
}

// Show renders the SMTP configuration form.
func (h *AdminSMTP) Show(w http.ResponseWriter, r *http.Request) {
	data := h.pageData(r)
	data.Message = r.URL.Query().Get("msg")

	if cfg, ok, err := h.settings().LoadSMTP(r.Context()); err != nil {
		slog.Error("load smtp settings", "err", err)
		data.Error = "Impossible de charger la configuration SMTP."
	} else if ok {
		data.Host = cfg.Host
		data.Port = cfg.Port
		data.TLS = cfg.TLS
		data.Username = cfg.Username
		data.From = cfg.From
		data.HasPassword = cfg.Password != ""
		data.Configured = cfg.Enabled()
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "admin_smtp", data); err != nil {
		slog.Error("render admin smtp", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Save stores SMTP settings or sends a test email.
func (h *AdminSMTP) Save(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	action := strings.TrimSpace(r.FormValue("action"))
	switch action {
	case "test":
		h.sendTest(w, r)
	default:
		h.saveConfig(w, r)
	}
}

func (h *AdminSMTP) saveConfig(w http.ResponseWriter, r *http.Request) {
	if len(h.EncryptionKey) == 0 {
		h.renderForm(w, r, templates.AdminSMTPData{
			Error: "REVUES_ENCRYPTION_KEY est requis pour enregistrer la configuration SMTP.",
		})
		return
	}

	port, err := settings.ParsePort(r.FormValue("port"))
	if err != nil {
		h.renderForm(w, r, templates.AdminSMTPData{Error: err.Error()})
		return
	}

	cfg := settings.SMTPConfig{
		Host:     strings.TrimSpace(r.FormValue("host")),
		Port:     port,
		TLS:      r.FormValue("tls") == "on" || r.FormValue("tls") == "1",
		Username: strings.TrimSpace(r.FormValue("username")),
		From:     strings.TrimSpace(r.FormValue("from")),
	}

	current, hasCurrent, err := h.settings().LoadSMTP(r.Context())
	if err != nil {
		slog.Error("load smtp for merge", "err", err)
		h.renderForm(w, r, templates.AdminSMTPData{Error: "Impossible de charger la configuration existante."})
		return
	}
	if hasCurrent {
		cfg.Password = settings.MergePassword(current, r.FormValue("password"))
	} else {
		cfg.Password = r.FormValue("password")
	}

	if err := h.settings().SaveSMTP(r.Context(), cfg); err != nil {
		var msg string
		switch {
		case errors.Is(err, settings.ErrEncryptionNotConfigured):
			msg = "REVUES_ENCRYPTION_KEY est requis pour enregistrer la configuration SMTP."
		default:
			msg = err.Error()
		}
		h.renderForm(w, r, templates.AdminSMTPData{
			Host:     cfg.Host,
			Port:     cfg.Port,
			TLS:      cfg.TLS,
			Username: cfg.Username,
			From:     cfg.From,
			Error:    msg,
		})
		return
	}

	http.Redirect(w, r, "/admin/settings/smtp?msg=Configuration+SMTP+enregistr%C3%A9e", http.StatusSeeOther)
}

func (h *AdminSMTP) sendTest(w http.ResponseWriter, r *http.Request) {
	cfg, ok, err := h.settings().LoadSMTP(r.Context())
	if err != nil {
		slog.Error("load smtp for test", "err", err)
		h.renderForm(w, r, templates.AdminSMTPData{Error: "Impossible de charger la configuration SMTP."})
		return
	}
	if !ok || !cfg.Enabled() {
		h.renderForm(w, r, templates.AdminSMTPData{Error: "Configurez et enregistrez le relais SMTP avant d'envoyer un email de test."})
		return
	}

	recipient := strings.TrimSpace(r.FormValue("test_recipient"))
	if recipient == "" {
		if user, ok := middleware.UserFromContext(r.Context()); ok {
			recipient = user.Email
		}
	}
	if _, err := mail.ParseAddress(recipient); err != nil {
		h.renderForm(w, r, templates.AdminSMTPData{Error: "Destinataire de test invalide."})
		return
	}

	mailer := notifications.Mailer{Config: cfg}
	if err := mailer.Send(r.Context(), recipient, "Test SMTP Revues", "Ceci est un email de test envoyé depuis Revues."); err != nil {
		slog.Error("send smtp test", "err", err)
		h.renderForm(w, r, templates.AdminSMTPData{
			Host:          cfg.Host,
			Port:          cfg.Port,
			TLS:           cfg.TLS,
			Username:      cfg.Username,
			From:          cfg.From,
			HasPassword:   cfg.Password != "",
			Configured:    true,
			TestRecipient: recipient,
			Error:         "Échec de l'envoi de l'email de test. Vérifiez hôte, port, TLS et identifiants.",
		})
		return
	}

	http.Redirect(w, r, "/admin/settings/smtp?msg=Email+de+test+envoy%C3%A9", http.StatusSeeOther)
}

func (h *AdminSMTP) renderForm(w http.ResponseWriter, r *http.Request, partial templates.AdminSMTPData) {
	data := h.pageData(r)
	data.Host = partial.Host
	data.Port = partial.Port
	data.TLS = partial.TLS
	data.Username = partial.Username
	data.From = partial.From
	data.HasPassword = partial.HasPassword
	data.Configured = partial.Configured
	data.TestRecipient = partial.TestRecipient
	data.Message = partial.Message
	data.Error = partial.Error

	if data.Port == 0 {
		data.Port = 587
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	status := http.StatusOK
	if data.Error != "" {
		status = http.StatusBadRequest
	}
	w.WriteHeader(status)
	if err := h.Templates.ExecuteTemplate(w, "admin_smtp", data); err != nil {
		slog.Error("render admin smtp form", "err", err)
	}
}

func (h *AdminSMTP) pageData(r *http.Request) templates.AdminSMTPData {
	data := templates.AdminSMTPData{
		PageData:   h.PageData(r, "Configuration SMTP"),
		Port:       587,
		CanEncrypt: len(h.EncryptionKey) > 0,
	}
	data.AdminSection = "smtp"
	if user, ok := middleware.UserFromContext(r.Context()); ok {
		data.TestRecipient = user.Email
	}
	return data
}

func (h *AdminSMTP) settings() *settings.SettingsService {
	return &settings.SettingsService{
		Store:         h.Store,
		EncryptionKey: h.EncryptionKey,
	}
}
