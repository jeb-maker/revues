package apiv1

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	adminsettings "github.com/jeb-maker/revues/internal/features/admin/settings"
	"github.com/jeb-maker/revues/internal/integrations/webhooks"
	"github.com/jeb-maker/revues/internal/store"
)

// GetAdminWebhooks serves GET /api/v1/admin/webhooks.
func (s *Server) GetAdminWebhooks(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}
	if s.Settings == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	writeJSON(w, http.StatusOK, mapWebhookSettings(s.loadWebhooksMasked(r)))
}

// PutAdminWebhooks serves PUT /api/v1/admin/webhooks.
func (s *Server) PutAdminWebhooks(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}
	if s.Settings == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	var req WebhookSettingsUpdate
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
		return
	}

	urls := adminsettings.ParseWebhookURLs(strings.Join(req.Urls, "\n"))
	devMode := s.Config.Env == "development"
	for _, raw := range urls {
		if err := webhooks.ValidateTargetURL(raw, devMode); err != nil {
			writeAPIError(w, http.StatusBadRequest, "validation_failed",
				"URL webhook refusée (scheme) : "+raw)
			return
		}
		if err := webhooks.ValidateTargetHost(r.Context(), raw, devMode); err != nil {
			writeAPIError(w, http.StatusBadRequest, "validation_failed",
				"URL webhook refusée (anti-SSRF) : "+raw)
			return
		}
	}

	cfg := adminsettings.WebhookConfig{
		URLs:            urls,
		ReviewCompleted: req.ReviewCompleted,
		ReviewItemNOK:   req.ReviewItemNok,
	}
	submittedSecret := ""
	if req.Secret != nil {
		submittedSecret = *req.Secret
	}

	current, hasCurrent, err := s.Settings.LoadWebhooks(r.Context())
	if err != nil {
		slog.Error("load webhooks for merge", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if hasCurrent {
		cfg.Secret = adminsettings.MergeWebhookSecret(current, submittedSecret)
	} else {
		cfg.Secret = submittedSecret
	}

	if err := adminsettings.ValidateWebhooks(cfg); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", err.Error())
		return
	}
	if err := s.Settings.SaveWebhooks(r.Context(), cfg); err != nil {
		writeConfigSaveError(w, err, adminsettings.ErrEncryptionNotConfigured,
			"REVUES_ENCRYPTION_KEY est requis pour enregistrer la configuration webhooks.")
		return
	}

	writeJSON(w, http.StatusOK, mapWebhookSettings(s.loadWebhooksMasked(r)))
}

// DeleteAdminWebhooks serves DELETE /api/v1/admin/webhooks.
func (s *Server) DeleteAdminWebhooks(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}
	if s.Settings == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if err := s.Settings.ClearWebhooks(r.Context()); err != nil {
		slog.Error("clear webhooks", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PostAdminWebhooksTest serves POST /api/v1/admin/webhooks/test.
func (s *Server) PostAdminWebhooksTest(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}
	if s.Settings == nil || s.Webhooks == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	cfg, configured, err := s.Settings.LoadWebhooks(r.Context())
	if err != nil {
		slog.Error("load webhooks for test", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if !configured || !cfg.Enabled() {
		writeAPIError(w, http.StatusBadRequest, "validation_failed",
			"Configurez et enregistrez les webhooks avant d'envoyer un test.")
		return
	}

	if err := s.Webhooks.SendTest(r.Context()); err != nil {
		slog.Error("send webhook test", "err", err)
		writeAPIError(w, http.StatusBadRequest, "validation_failed",
			"Échec du test webhook. Vérifiez les URLs (anti-SSRF) et la disponibilité des cibles.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListAdminWebhookDeliveries serves GET /api/v1/admin/webhooks/deliveries.
func (s *Server) ListAdminWebhookDeliveries(w http.ResponseWriter, r *http.Request, params ListAdminWebhookDeliveriesParams) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}
	if s.Store == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	limit := 50
	if params.Limit != nil && *params.Limit > 0 {
		limit = *params.Limit
	}
	rows, err := s.Store.ListOrgWebhookDeliveries(r.Context(), limit)
	if err != nil {
		slog.Error("list webhook deliveries", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	out := make([]WebhookDelivery, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapWebhookDelivery(row))
	}
	writeJSON(w, http.StatusOK, WebhookDeliveryList{Deliveries: out})
}

// PostAdminWebhookDeliveriesDrain serves POST /api/v1/admin/webhooks/deliveries/drain.
func (s *Server) PostAdminWebhookDeliveriesDrain(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}
	if s.Webhooks == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if err := s.Webhooks.Drain(r.Context()); err != nil {
		slog.Error("webhook drain", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PostAdminWebhookDeliveryRetry serves POST /api/v1/admin/webhooks/deliveries/{deliveryId}/retry.
//
//nolint:staticcheck // SA1003: deliveryId matches OpenAPI path param / oapi-codegen ServerInterface
func (s *Server) PostAdminWebhookDeliveryRetry(w http.ResponseWriter, r *http.Request, deliveryId WebhookDeliveryId) {
	if _, _, ok := s.requireOrgAdmin(w, r); !ok {
		return
	}
	if s.Store == nil || s.Webhooks == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	del, err := s.Store.WebhookDeliveryByID(r.Context(), deliveryId)
	if err != nil {
		if errors.Is(err, store.ErrWebhookDeliveryNotFound) {
			writeAPIError(w, http.StatusNotFound, "not_found", "Livraison introuvable.")
			return
		}
		slog.Error("load webhook delivery", "err", err, "delivery_id", deliveryId)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if del.State == store.WebhookDeliveryDone {
		writeAPIError(w, http.StatusBadRequest, "validation_failed",
			"Une livraison réussie ne peut pas être relancée.")
		return
	}

	now := time.Now().UTC()
	if err := s.Store.ResetWebhookDeliveryForRetry(r.Context(), deliveryId, now, now.Add(webhooks.DeliveryTTL)); err != nil {
		if errors.Is(err, store.ErrWebhookDeliveryNotFound) {
			writeAPIError(w, http.StatusNotFound, "not_found", "Livraison introuvable.")
			return
		}
		slog.Error("reset webhook delivery", "err", err, "delivery_id", deliveryId)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if err := s.Webhooks.Drain(r.Context()); err != nil {
		slog.Error("webhook drain after retry", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type webhooksMasked struct {
	configured      bool
	enabled         bool
	urls            []string
	hasSecret       bool
	reviewCompleted bool
	reviewItemNOK   bool
}

func (s *Server) loadWebhooksMasked(r *http.Request) webhooksMasked {
	cfg, ok, err := s.Settings.LoadWebhooks(r.Context())
	if err != nil {
		slog.Error("load webhooks settings", "err", err)
		return webhooksMasked{urls: []string{}}
	}
	if !ok {
		return webhooksMasked{urls: []string{}}
	}
	urls := cfg.URLs
	if urls == nil {
		urls = []string{}
	}
	return webhooksMasked{
		configured:      true,
		enabled:         cfg.Enabled(),
		urls:            urls,
		hasSecret:       strings.TrimSpace(cfg.Secret) != "",
		reviewCompleted: cfg.ReviewCompleted,
		reviewItemNOK:   cfg.ReviewItemNOK,
	}
}

func mapWebhookSettings(m webhooksMasked) WebhookSettings {
	urls := m.urls
	if urls == nil {
		urls = []string{}
	}
	return WebhookSettings{
		Configured:      m.configured,
		Enabled:         m.enabled,
		Urls:            urls,
		HasSecret:       m.hasSecret,
		ReviewCompleted: m.reviewCompleted,
		ReviewItemNok:   m.reviewItemNOK,
	}
}

func mapWebhookDelivery(d store.WebhookDelivery) WebhookDelivery {
	out := WebhookDelivery{
		Id:        d.ID,
		EventId:   d.EventID,
		EventType: d.EventType,
		Url:       d.URL,
		Success:   d.Success,
		Attempts:  d.Attempts,
		State:     WebhookDeliveryState(d.State),
		CreatedAt: d.CreatedAt,
	}
	if d.StatusCode.Valid {
		code := int(d.StatusCode.Int64)
		out.StatusCode = &code
	}
	if d.NextAttemptAt != "" {
		v := d.NextAttemptAt
		out.NextAttemptAt = &v
	}
	if d.ExpiresAt != "" {
		v := d.ExpiresAt
		out.ExpiresAt = &v
	}
	if d.LastError != "" {
		v := d.LastError
		out.LastError = &v
	}
	return out
}
