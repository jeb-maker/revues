package handlers

import (
	"log/slog"
	"net/http"
)

// Health responds to load-balancer and orchestrator probes.
func Health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		slog.Error("write healthz response", "err", err)
	}
}
