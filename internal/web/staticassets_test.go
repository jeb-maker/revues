package web_test

import (
	"bytes"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"

	appweb "github.com/jeb-maker/revues/internal/web"
	webassets "github.com/jeb-maker/revues/web"
)

func TestStaticHandlerSetsCacheControl(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("development", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/vendor/jeb-maker-mb/mb-boot.js", nil)
		rec := httptest.NewRecorder()
		appweb.StaticHandler(inner, "development").ServeHTTP(rec, req)
		if got := rec.Header().Get("Cache-Control"); got != "no-cache" {
			t.Fatalf("Cache-Control = %q, want no-cache", got)
		}
	})

	t.Run("production", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/vendor/jeb-maker-mb/mb-boot.js", nil)
		rec := httptest.NewRecorder()
		appweb.StaticHandler(inner, "production").ServeHTTP(rec, req)
		if got := rec.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
			t.Fatalf("Cache-Control = %q", got)
		}
	})
}

func TestVendoredMBBundlePresent(t *testing.T) {
	staticFS, err := fs.Sub(webassets.Static, "static")
	if err != nil {
		t.Fatalf("fs.Sub(): %v", err)
	}
	for _, path := range []string{
		"vendor/jeb-maker-mb/mb-boot.js",
		"vendor/jeb-maker-mb/mb-bridge.css",
		"vendor/jeb-maker-mb/tokens/tokens-core.css",
		"vendor/jeb-maker-mb/tokens/reference.css",
		"vendor/jeb-maker-mb/tokens/semantic.css",
	} {
		data, readErr := fs.ReadFile(staticFS, path)
		if readErr != nil {
			t.Fatalf("ReadFile(%s): %v", path, readErr)
		}
		if len(data) == 0 {
			t.Fatalf("%s empty", path)
		}
	}
	boot, err := fs.ReadFile(staticFS, "vendor/jeb-maker-mb/mb-boot.js")
	if err != nil {
		t.Fatalf("ReadFile(mb-boot): %v", err)
	}
	for _, tag := range []string{
		"mb-button", "mb-select", "mb-combobox", "mb-progress", "mb-empty-state", "mb-segmented-control", "mb-pagination",
		"mb-tag", "mb-breadcrumbs", "mb-nav", "mb-nav-toggle", "mb-avatar", "mb-spinner", "mb-toolbar", "mb-card",
		"mb-table", "mb-table-row", "mb-table-cell",
	} {
		if !bytes.Contains(boot, []byte(tag)) {
			t.Fatalf("mb-boot.js missing registration for %s", tag)
		}
	}
}
