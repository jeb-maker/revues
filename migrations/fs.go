// Package migrations embeds goose SQL migration files.
package migrations

import "embed"

// Files contains numbered goose migrations.
//
//go:embed *.sql
var Files embed.FS
