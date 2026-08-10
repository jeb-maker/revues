package notion

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// APIError is a Notion HTTP API failure with enough detail for logging and UX mapping.
type APIError struct {
	Status int
	Code   string
	Detail string
	Body   string
	Err    error
}

func (e *APIError) Error() string {
	if e == nil {
		return "notion api error"
	}
	base := "notion api error"
	if e.Err != nil {
		base = e.Err.Error()
	}
	if e.Status == 0 {
		return base
	}
	msg := strings.TrimSpace(e.Detail)
	if msg == "" {
		msg = strings.TrimSpace(e.Body)
	}
	if msg == "" {
		return fmt.Sprintf("%s: status %d", base, e.Status)
	}
	return fmt.Sprintf("%s: status %d %s", base, e.Status, msg)
}

func (e *APIError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// parseAPIError reads a Notion error body (best-effort) and wraps fallback.
func parseAPIError(resp *http.Response, fallback error) error {
	if resp == nil {
		return fallback
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	var payload struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(body, &payload)
	return &APIError{
		Status: resp.StatusCode,
		Code:   strings.TrimSpace(payload.Code),
		Detail: strings.TrimSpace(payload.Message),
		Body:   strings.TrimSpace(string(body)),
		Err:    fallback,
	}
}

// UserMessage returns a French, actionable message for import/export UX.
// Technical payloads (JSON, status dumps) are never shown to end users.
func UserMessage(err error) string {
	if err == nil {
		return ""
	}
	switch {
	case errors.Is(err, ErrNotConfigured):
		return "Notion n'est pas configuré. Demandez à un administrateur de configurer l'intégration."
	case errors.Is(err, ErrDatabaseMissing):
		return "Aucune base Notion par défaut configurée. Demandez à un administrateur de renseigner l'identifiant de base."
	case errors.Is(err, ErrAlreadyExported):
		return "Cette revue a déjà été exportée vers Notion."
	case errors.Is(err, ErrRunNotDone):
		return "Seules les revues terminées peuvent être exportées vers Notion."
	case errors.Is(err, ErrDatabaseNotFound):
		return "Base Notion introuvable. Vérifiez l'URL ou l'identifiant, et le partage avec l'intégration."
	}

	var api *APIError
	if errors.As(err, &api) {
		switch api.Status {
		case http.StatusUnauthorized:
			return "Jeton Notion invalide ou expiré. Demandez à un administrateur de vérifier la configuration Notion."
		case http.StatusForbidden:
			return "Accès Notion refusé. Partagez la base avec l'intégration Notion, puis réessayez."
		case http.StatusNotFound:
			return "Ressource Notion introuvable. Vérifiez l'identifiant et le partage avec l'intégration."
		case http.StatusTooManyRequests:
			return "Notion limite temporairement les appels. Réessayez dans quelques minutes."
		case http.StatusBadRequest:
			if api.Code == "validation_error" {
				return "Notion a refusé la requête (colonnes incompatibles). Vérifiez que la base contient les propriétés attendues (Name, Sujet, Date, Lien Revues)."
			}
			return "Requête Notion invalide. Vérifiez la configuration de la base et réessayez."
		case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			return "Notion est temporairement indisponible. Réessayez plus tard."
		}
	}

	if errors.Is(err, ErrConnectionFailed) || errors.Is(err, ErrExportFailed) {
		return "Impossible de joindre Notion. Vérifiez la connectivité réseau et réessayez."
	}

	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		return "Erreur Notion. Réessayez ou contactez un administrateur."
	}
	// Pass through already-French validation / form messages; hide raw API dumps.
	if looksTechnicalNotionError(msg) {
		return "Erreur Notion. Réessayez ou contactez un administrateur."
	}
	return msg
}

func looksTechnicalNotionError(msg string) bool {
	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(lower, "status "),
		strings.Contains(lower, `"object"`),
		strings.Contains(lower, "notion api"),
		strings.Contains(lower, "invalid response"),
		strings.Contains(lower, "build notion"),
		strings.Contains(lower, "marshal "):
		return true
	default:
		return false
	}
}
