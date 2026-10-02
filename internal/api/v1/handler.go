package apiv1

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	authfeature "github.com/jeb-maker/revues/internal/features/auth"
)

// Server implements the OpenAPI ServerInterface for /api/v1.
type Server struct {
	Auth     *authfeature.Service
	Config   config.Config
	Sessions *auth.SessionManager
}

// NewServer returns the API v1 server implementation.
func NewServer(svc *authfeature.Service, cfg config.Config, sessions *auth.SessionManager) *Server {
	return &Server{Auth: svc, Config: cfg, Sessions: sessions}
}

// GetHealth serves GET /api/v1/health.
func (s *Server) GetHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{Status: Ok})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("write api json response", "err", err)
	}
}

func writeAPIError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, ErrorResponse{
		Error: ErrorBody{Code: code, Message: message},
	})
}
