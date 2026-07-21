package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// RateLimitConfig controls a sliding-window limiter.
type RateLimitConfig struct {
	// Max is the maximum number of requests allowed in Window.
	Max int
	// Window is the sliding window duration.
	Window time.Duration
	// Key extracts the rate-limit bucket key (default: client IP).
	Key func(*http.Request) string
}

// RateLimit returns middleware that answers 429 when the bucket is exhausted.
func RateLimit(cfg RateLimitConfig) func(http.Handler) http.Handler {
	if cfg.Max < 1 {
		cfg.Max = 60
	}
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}
	if cfg.Key == nil {
		cfg.Key = clientIP
	}
	lim := &slidingWindow{max: cfg.Max, window: cfg.Window}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := cfg.Key(r)
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}
			if !lim.allow(key, time.Now()) {
				w.Header().Set("Retry-After", "60")
				http.Error(w, "Trop de requêtes. Réessayez dans une minute.", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

type slidingWindow struct {
	mu     sync.Mutex
	max    int
	window time.Duration
	hits   map[string][]time.Time
}

func (s *slidingWindow) allow(key string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.hits == nil {
		s.hits = make(map[string][]time.Time)
	}
	cutoff := now.Add(-s.window)
	kept := s.hits[key][:0]
	for _, t := range s.hits[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= s.max {
		s.hits[key] = kept
		return false
	}
	s.hits[key] = append(kept, now)
	return true
}
