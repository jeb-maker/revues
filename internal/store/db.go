package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"

	"github.com/jeb-maker/revues/migrations"
)

var gooseMu sync.Mutex

// DefaultMaxOpenConns is the SQLite connection pool size when none is configured.
const DefaultMaxOpenConns = 10

// DefaultBusyTimeoutMS is how long SQLite waits on SQLITE_BUSY before failing.
const DefaultBusyTimeoutMS = 5000

// Open connects to SQLite at path and applies connection pragmas.
// maxOpenConns <= 0 uses DefaultMaxOpenConns.
func Open(ctx context.Context, path string, maxOpenConns int) (*sql.DB, error) {
	if err := ensureParentDir(path); err != nil {
		return nil, fmt.Errorf("ensure database directory: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(%d)", path, DefaultBusyTimeoutMS)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if maxOpenConns <= 0 {
		maxOpenConns = DefaultMaxOpenConns
	}
	db.SetMaxOpenConns(maxOpenConns)
	idle := maxOpenConns / 2
	if idle < 1 {
		idle = 1
	}
	db.SetMaxIdleConns(idle)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	if err := configureSQLite(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

// Migrate runs pending goose migrations embedded in the migrations package.
func Migrate(ctx context.Context, db *sql.DB) error {
	gooseMu.Lock()
	defer gooseMu.Unlock()

	goose.SetBaseFS(migrations.Files)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("goose dialect: %w", err)
	}

	if err := goose.UpContext(ctx, db, "."); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	return nil
}

// Down rolls back all migrations (used in tests).
func Down(ctx context.Context, db *sql.DB) error {
	gooseMu.Lock()
	defer gooseMu.Unlock()

	goose.SetBaseFS(migrations.Files)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("goose dialect: %w", err)
	}

	if err := goose.DownToContext(ctx, db, ".", 0); err != nil {
		return fmt.Errorf("goose down: %w", err)
	}

	return nil
}

func configureSQLite(ctx context.Context, db *sql.DB) error {
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		fmt.Sprintf("PRAGMA busy_timeout = %d", DefaultBusyTimeoutMS),
	}

	for _, q := range pragmas {
		if _, err := db.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("exec %q: %w", q, err)
		}
	}

	return nil
}

func ensureParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "" || dir == "." {
		return nil
	}

	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}

	return nil
}
