package web

import (
	"net/http"
)

// DevNoCache disables caching in development (HTML and API responses).
func DevNoCache(env string) func(http.Handler) http.Handler {
	if env != "development" {
		return func(next http.Handler) http.Handler { return next }
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-cache")
			next.ServeHTTP(w, r)
		})
	}
}

// StaticHandler serves embedded static files with cache headers suited to the environment.
func StaticHandler(fileServer http.Handler, env string) http.Handler {
	cacheControl := "public, max-age=31536000, immutable"
	if env == "development" {
		cacheControl = "no-cache"
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", cacheControl)
		fileServer.ServeHTTP(w, r)
	})
}
