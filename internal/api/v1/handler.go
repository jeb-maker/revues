package apiv1

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// Server implements the OpenAPI ServerInterface for /api/v1.
type Server struct{}

// NewServer returns the API v1 server implementation.
func NewServer() *Server {
	return &Server{}
}

// GetHealth serves GET /api/v1/health.
func (s *Server) GetHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	resp := HealthResponse{Status: Ok}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("write api health response", "err", err)
	}
}
