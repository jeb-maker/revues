package config_test

import (
	"os"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/config"
)

func TestLoadDBMaxOpenConns(t *testing.T) {
	t.Setenv("REVUES_DB_MAX_OPEN_CONNS", "15")
	cfg := config.Load()
	if cfg.DBMaxOpenConns != 15 {
		t.Fatalf("DBMaxOpenConns = %d, want 15", cfg.DBMaxOpenConns)
	}
}

func TestLoadDBMaxOpenConnsInvalidFallsBack(t *testing.T) {
	t.Setenv("REVUES_DB_MAX_OPEN_CONNS", "0")
	cfg := config.Load()
	if cfg.DBMaxOpenConns != 10 {
		t.Fatalf("DBMaxOpenConns = %d, want default 10", cfg.DBMaxOpenConns)
	}
}

func TestLoadDBMaxOpenConnsUnsetUsesDefault(t *testing.T) {
	os.Unsetenv("REVUES_DB_MAX_OPEN_CONNS")
	cfg := config.Load()
	if cfg.DBMaxOpenConns != 10 {
		t.Fatalf("DBMaxOpenConns = %d, want default 10", cfg.DBMaxOpenConns)
	}
}

func TestValidate_ProductionRejectsDefaultSessionSecret(t *testing.T) {
	t.Setenv("REVUES_ENV", "production")
	t.Setenv("REVUES_SESSION_SECRET", "change-me-32-random-bytes-minimum")
	cfg := config.Load()
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want rejection of default secret")
	}
}

func TestValidate_ProductionRejectsShortSessionSecret(t *testing.T) {
	t.Setenv("REVUES_ENV", "production")
	t.Setenv("REVUES_SESSION_SECRET", "too-short")
	cfg := config.Load()
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "at least 32") {
		t.Fatalf("Validate() error = %v, want length complaint", err)
	}
}

func TestValidate_ProductionAcceptsStrongSecret(t *testing.T) {
	t.Setenv("REVUES_ENV", "production")
	t.Setenv("REVUES_SESSION_SECRET", "unique-production-secret-32b-ok!!")
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidate_DevelopmentAllowsDefaultSecret(t *testing.T) {
	t.Setenv("REVUES_ENV", "development")
	os.Unsetenv("REVUES_SESSION_SECRET")
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestSecureCookies_HTTPSBaseURL(t *testing.T) {
	t.Setenv("REVUES_ENV", "development")
	t.Setenv("REVUES_BASE_URL", "https://revues.example.com")
	cfg := config.Load()
	if !cfg.SecureCookies() {
		t.Fatal("SecureCookies() = false, want true for https BaseURL")
	}
}
