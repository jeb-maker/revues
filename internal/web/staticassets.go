package web

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"net/http"
	"sort"
)

// StaticAssetVersion returns a short content hash for cache busting embedded assets.
func StaticAssetVersion(staticFS fs.FS) (string, error) {
	entries, err := fs.Glob(staticFS, "**/*")
	if err != nil {
		return "", fmt.Errorf("glob static assets: %w", err)
	}
	sort.Strings(entries)

	h := sha256.New()
	for _, path := range entries {
		info, statErr := fs.Stat(staticFS, path)
		if statErr != nil {
			return "", fmt.Errorf("stat static asset %q: %w", path, statErr)
		}
		if info.IsDir() {
			// fs.Glob may return nested directories (e.g. vendor/…); only hash files.
			continue
		}
		data, readErr := fs.ReadFile(staticFS, path)
		if readErr != nil {
			return "", fmt.Errorf("read static asset %q: %w", path, readErr)
		}
		_, _ = h.Write([]byte(path))
		_, _ = h.Write(data)
	}

	sum := hex.EncodeToString(h.Sum(nil))
	if len(sum) < 12 {
		return sum, nil
	}
	return sum[:12], nil
}

// ServeServiceWorkerKill serves a script that unregisters orphan service workers.
func ServeServiceWorkerKill(staticFS fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := fs.ReadFile(staticFS, "sw.js")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(data)
	}
}

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
