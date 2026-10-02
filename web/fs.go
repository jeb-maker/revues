// Package web holds embedded static assets (vendor mb, etc.).
package web

import "embed"

// Static contains CSS/JS served at /static/ (vendor mb kept for WP-005).
//
//go:embed all:static
var Static embed.FS
