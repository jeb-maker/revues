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

func TestStaticAssetVersionStable(t *testing.T) {
	staticFS, err := fs.Sub(webassets.Static, "static")
	if err != nil {
		t.Fatalf("fs.Sub(): %v", err)
	}

	v1, err := appweb.StaticAssetVersion(staticFS)
	if err != nil {
		t.Fatalf("StaticAssetVersion(): %v", err)
	}
	if len(v1) != 12 {
		t.Fatalf("version len = %d, want 12", len(v1))
	}

	v2, err := appweb.StaticAssetVersion(staticFS)
	if err != nil {
		t.Fatalf("StaticAssetVersion(2): %v", err)
	}
	if v1 != v2 {
		t.Fatalf("version = %q vs %q, want stable", v1, v2)
	}
}

func TestStaticHandlerSetsCacheControl(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("development", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/css/app.css", nil)
		rec := httptest.NewRecorder()
		appweb.StaticHandler(inner, "development").ServeHTTP(rec, req)
		if got := rec.Header().Get("Cache-Control"); got != "no-cache" {
			t.Fatalf("Cache-Control = %q, want no-cache", got)
		}
	})

	t.Run("production", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/css/app.css", nil)
		rec := httptest.NewRecorder()
		appweb.StaticHandler(inner, "production").ServeHTTP(rec, req)
		if got := rec.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
			t.Fatalf("Cache-Control = %q", got)
		}
	})
}

func TestSplitCSSAssetsPresent(t *testing.T) {
	staticFS, err := fs.Sub(webassets.Static, "static")
	if err != nil {
		t.Fatalf("fs.Sub(): %v", err)
	}
	for _, path := range []string{"css/app.css", "css/run.css", "css/editor.css"} {
		data, readErr := fs.ReadFile(staticFS, path)
		if readErr != nil {
			t.Fatalf("ReadFile(%s): %v", path, readErr)
		}
		if len(data) == 0 {
			t.Fatalf("%s empty", path)
		}
	}
}

func TestVendoredReportsBundlePresent(t *testing.T) {
	staticFS, err := fs.Sub(webassets.Static, "static")
	if err != nil {
		t.Fatalf("fs.Sub(): %v", err)
	}
	for _, path := range []string{
		"vendor/jeb-maker-reports/reports.min.js",
		"vendor/jeb-maker-reports/init.js",
	} {
		data, readErr := fs.ReadFile(staticFS, path)
		if readErr != nil {
			t.Fatalf("ReadFile(%s): %v", path, readErr)
		}
		if len(data) == 0 {
			t.Fatalf("%s empty", path)
		}
	}
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
		"vendor/jeb-maker-mb/components/button.js",
		"vendor/jeb-maker-mb/components/select.js",
		"vendor/jeb-maker-mb/components/progress.js",
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
	for _, tag := range []string{"mb-button", "mb-select", "mb-progress", "mb-empty-state", "mb-segmented-control", "mb-pagination"} {
		if !bytes.Contains(boot, []byte(tag)) {
			t.Fatalf("mb-boot.js missing registration for %s", tag)
		}
	}
}

func TestServeServiceWorkerKill(t *testing.T) {
	staticFS, err := fs.Sub(webassets.Static, "static")
	if err != nil {
		t.Fatalf("fs.Sub(): %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/sw.js", nil)
	rec := httptest.NewRecorder()
	appweb.ServeServiceWorkerKill(staticFS)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/javascript; charset=utf-8" {
		t.Fatalf("Content-Type = %q", ct)
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", cc)
	}
	if len(rec.Body.String()) == 0 {
		t.Fatal("empty body")
	}
}
