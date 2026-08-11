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

func TestRateLimit_IgnoresSpoofedXForwardedFor(t *testing.T) {
	t.Parallel()
	lim := RateLimit(RateLimitConfig{Max: 1, Window: time.Minute})
	h := CapturePeerAddr(lim(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))

	// First request from real peer — allowed.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	req.RemoteAddr = "203.0.113.50:9999"
	req.Header.Set("X-Forwarded-For", "198.51.100.1")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("1st status = %d", rec.Code)
	}

	// Same peer, different spoofed XFF — still counted on peer key → 429.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/login", nil)
	req.RemoteAddr = "203.0.113.50:9999"
	req.Header.Set("X-Forwarded-For", "198.51.100.2")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("2nd status = %d, want 429 (peer key must ignore XFF)", rec.Code)
	}
}

func TestSlidingWindow_GCRemovesStaleBuckets(t *testing.T) {
	t.Parallel()
	s := &slidingWindow{max: 1, window: time.Millisecond}
	now := time.Now()
	if !s.allow("a", now) {
		t.Fatal("first allow")
	}
	s.ops = 63
	later := now.Add(2 * time.Millisecond)
	if !s.allow("b", later) {
		t.Fatal("second allow should trigger gc")
	}
	s.mu.Lock()
	_, hasA := s.hits["a"]
	s.mu.Unlock()
	if hasA {
		t.Fatal("stale bucket a should be garbage-collected")
	}
}
