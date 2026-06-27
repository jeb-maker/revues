package middleware

import (
	"net/http"
	"strings"

	"github.com/jeb-maker/revues/internal/auth"
)

// CSRF validates CSRF tokens on mutating requests.
func CSRF(sessionSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !needsCSRF(r) {
				next.ServeHTTP(w, r)
				return
			}

			sessionToken, err := auth.SessionTokenFromRequest(r)
			if err != nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			token := r.Header.Get("X-CSRF-Token")
			if token == "" {
				if err := r.ParseForm(); err != nil {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
				token = r.FormValue("csrf_token")
			}

			if !auth.ValidateCSRF(sessionToken, sessionSecret, token) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func needsCSRF(r *http.Request) bool {
	switch r.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return false
	default:
	}

	path := r.URL.Path
	if strings.HasPrefix(path, "/auth/github/callback") {
		return false
	}

	return true
}
