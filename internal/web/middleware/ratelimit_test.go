package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimit_AllowsThenBlocks(t *testing.T) {
	t.Parallel()
	lim := RateLimit(RateLimitConfig{Max: 2, Window: time.Minute})
	var hits int
	h := lim(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusNoContent)
	}))

	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/auth/github/start", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("request %d: status = %d, want 204", i+1, rec.Code)
		}
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/github/start", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("3rd request status = %d, want 429", rec.Code)
	}
	if hits != 2 {
		t.Fatalf("hits = %d, want 2", hits)
	}
}

func TestRateLimit_SeparateKeys(t *testing.T) {
	t.Parallel()
	lim := RateLimit(RateLimitConfig{Max: 1, Window: time.Minute})
	h := lim(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for _, ip := range []string{"203.0.113.1:1", "203.0.113.2:1"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		req.RemoteAddr = ip
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("ip %s: status = %d", ip, rec.Code)
		}
	}
}
