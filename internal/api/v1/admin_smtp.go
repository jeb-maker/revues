package apiv1

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"

	adminintegrations "github.com/jeb-maker/revues/internal/features/admin/integrations"
	adminsettings "github.com/jeb-maker/revues/internal/features/admin/settings"
	"github.com/jeb-maker/revues/internal/notifications"
)

// GetAdminSMTPSettings serves GET /api/v1/admin/settings/smtp.
func (s *Server) GetAdminSMTPSettings(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}
	if s.Settings == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	writeJSON(w, http.StatusOK, mapSMTPSettings(s.loadSMTPMasked(r)))
}

// PutAdminSMTPSettings serves PUT /api/v1/admin/settings/smtp.
func (s *Server) PutAdminSMTPSettings(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}
	if s.Settings == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	var req SMTPSettingsUpdate
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
		return
	}

	cfg := adminsettings.SMTPConfig{
		Host:     strings.TrimSpace(req.Host),
		Port:     req.Port,
		TLS:      true,
		Username: "",
		From:     strings.TrimSpace(req.From),
	}
	if req.Tls != nil {
		cfg.TLS = *req.Tls
	}
	if req.Username != nil {
		cfg.Username = strings.TrimSpace(*req.Username)
	}
	submittedPassword := ""
	if req.Password != nil {
		submittedPassword = *req.Password
	}

	current, hasCurrent, err := s.Settings.LoadSMTP(r.Context())
	if err != nil {
		slog.Error("load smtp for merge", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if hasCurrent {
		cfg.Password = adminsettings.MergePassword(current, submittedPassword)
	} else {
		cfg.Password = submittedPassword
	}

	if err := adminsettings.ValidateSMTP(cfg); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", err.Error())
		return
	}
	if err := s.Settings.SaveSMTP(r.Context(), cfg); err != nil {
		writeConfigSaveError(w, err, adminsettings.ErrEncryptionNotConfigured,
			"REVUES_ENCRYPTION_KEY est requis pour enregistrer la configuration SMTP.")
		return
	}

	writeJSON(w, http.StatusOK, mapSMTPSettings(s.loadSMTPMasked(r)))
}

// DeleteAdminSMTPSettings serves DELETE /api/v1/admin/settings/smtp.
func (s *Server) DeleteAdminSMTPSettings(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}
	if s.Settings == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if err := s.Settings.ClearSMTP(r.Context()); err != nil {
		slog.Error("clear smtp", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PostAdminSMTPTest serves POST /api/v1/admin/settings/smtp/test.
func (s *Server) PostAdminSMTPTest(w http.ResponseWriter, r *http.Request) {
	user, _, ok := s.requireOrgAdmin(w, r)
	if !ok {
		return
	}
	if s.Settings == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	var req SMTPTestRequest
	if r.Body != nil && r.ContentLength != 0 {
		if err := decodeJSONBody(r, &req); err != nil && !errors.Is(err, io.EOF) {
			writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
			return
		}
	}

	cfg, configured, err := s.Settings.LoadSMTP(r.Context())
	if err != nil {
		slog.Error("load smtp for test", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if !configured || !cfg.Enabled() {
		writeAPIError(w, http.StatusBadRequest, "validation_failed",
			"Configurez et enregistrez le relais SMTP avant d'envoyer un email de test.")
		return
	}

	recipient := user.Email
	if req.Recipient != nil && strings.TrimSpace(string(*req.Recipient)) != "" {
		recipient = strings.TrimSpace(string(*req.Recipient))
	}
	if _, err := mail.ParseAddress(recipient); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Destinataire de test invalide.")
		return
	}

	mailer := notifications.Mailer{Config: cfg}
	if err := mailer.Send(r.Context(), recipient, "Test SMTP Revues", "Ceci est un email de test envoyé depuis Revues."); err != nil {
		slog.Error("send smtp test", "err", err)
		writeAPIError(w, http.StatusBadRequest, "validation_failed",
			"Échec de l'envoi de l'email de test. Vérifiez hôte, port, TLS et identifiants.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListAdminIntegrations serves GET /api/v1/admin/integrations.
func (s *Server) ListAdminIntegrations(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}
	if s.Integrations == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	overview, err := s.Integrations.Overview(r.Context())
	if err != nil {
		slog.Error("integrations overview", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	items := make([]IntegrationSummary, 0, len(overview.Items))
	for _, row := range overview.Items {
		items = append(items, mapIntegrationSummary(row))
	}
	writeJSON(w, http.StatusOK, IntegrationsOverview{Items: items})
}

type smtpMasked struct {
	configured  bool
	enabled     bool
	host        string
	port        int
	tls         bool
	username    string
	from        string
	hasPassword bool
}

func (s *Server) loadSMTPMasked(r *http.Request) smtpMasked {
	cfg, ok, err := s.Settings.LoadSMTP(r.Context())
	if err != nil {
		slog.Error("load smtp settings", "err", err)
		return smtpMasked{port: 587, tls: true}
	}
	if !ok {
		return smtpMasked{port: 587, tls: true}
	}
	return smtpMasked{
		configured:  true,
		enabled:     cfg.Enabled(),
		host:        cfg.Host,
		port:        cfg.Port,
		tls:         cfg.TLS,
		username:    cfg.Username,
		from:        cfg.From,
		hasPassword: cfg.Password != "",
	}
}

func mapSMTPSettings(m smtpMasked) SMTPSettings {
	return SMTPSettings{
		Configured:  m.configured,
		Enabled:     m.enabled,
		Host:        m.host,
		Port:        m.port,
		Tls:         m.tls,
		Username:    m.username,
		From:        m.from,
		HasPassword: m.hasPassword,
	}
}

func mapIntegrationSummary(row adminintegrations.IntegrationSummary) IntegrationSummary {
	key := IntegrationSummaryKey(row.Key)
	switch row.Key {
	case adminintegrations.IntegrationKeySMTP:
		key = Smtp
	case adminintegrations.IntegrationKeyJira:
		key = Jira
	case adminintegrations.IntegrationKeyWebhooks:
		key = Webhooks
	}
	return IntegrationSummary{
		Key:         key,
		Name:        row.Name,
		Description: row.Description,
		Enabled:     row.Enabled,
		ConfigPath:  row.ConfigPath,
	}
}
