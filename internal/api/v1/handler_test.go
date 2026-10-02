package apiv1_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	apiv1 "github.com/jeb-maker/revues/internal/api/v1"
)

func TestGetHealth(t *testing.T) {
	t.Parallel()

	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		apiv1.HandlerFromMux(apiv1.NewServer(), r)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want application/json; charset=utf-8", ct)
	}

	var body apiv1.HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json decode: %v (body=%q)", err, rec.Body.String())
	}
	if body.Status != apiv1.Ok {
		t.Errorf("status = %q, want %q", body.Status, apiv1.Ok)
	}
}
