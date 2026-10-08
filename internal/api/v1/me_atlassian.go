package apiv1

import (
	"log/slog"
	"net/http"

	appmiddleware "github.com/jeb-maker/revues/internal/web/middleware"
)

// GetMeAtlassian serves GET /api/v1/me/atlassian.
func (s *Server) GetMeAtlassian(w http.ResponseWriter, r *http.Request) {
	user, ok := appmiddleware.UserFromContext(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthenticated", "Authentification requise.")
		return
	}

	enabled := s.Config.AtlassianOAuthConfigured()
	out := MeAtlassian{
		Enabled:      enabled,
		Connected:    false,
		SiteUrl:      "",
		AccountEmail: "",
		CloudId:      "",
	}
	if !enabled {
		writeJSON(w, http.StatusOK, out)
		return
	}

	tokens := s.atlassianTokenService()
	if tokens == nil {
		writeJSON(w, http.StatusOK, out)
		return
	}
	connected, row, err := tokens.Connected(r.Context(), user.ID)
	if err != nil {
		slog.Error("me atlassian status", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	out.Connected = connected
	if connected && row != nil {
		out.SiteUrl = row.SiteURL
		out.AccountEmail = row.AccountEmail
		out.CloudId = row.CloudID
	}
	writeJSON(w, http.StatusOK, out)
}

// DeleteMeAtlassian serves DELETE /api/v1/me/atlassian.
func (s *Server) DeleteMeAtlassian(w http.ResponseWriter, r *http.Request) {
	user, ok := appmiddleware.UserFromContext(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthenticated", "Authentification requise.")
		return
	}

	tokens := s.atlassianTokenService()
	if tokens == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err := tokens.Disconnect(r.Context(), user.ID); err != nil {
		slog.Error("disconnect atlassian", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
