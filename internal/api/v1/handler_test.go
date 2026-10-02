package apiv1_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	apiv1 "github.com/jeb-maker/revues/internal/api/v1"
	"github.com/jeb-maker/revues/internal/config"
)

func TestGetHealth(t *testing.T) {
	t.Parallel()

	handler := newTestRouter(t, config.Config{
		SessionSecret: "test-secret-at-least-thirty-two-bytes",
		Env:           "development",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body apiv1.HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json decode: %v (body=%q)", err, rec.Body.String())
	}
	if body.Status != apiv1.HealthResponseStatusOk {
		t.Errorf("status = %q, want %q", body.Status, apiv1.HealthResponseStatusOk)
	}
}
