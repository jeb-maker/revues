package config_test

import (
	"os"
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
