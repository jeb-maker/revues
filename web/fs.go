// Package web holds embedded static assets and HTML templates.
package web

import "embed"

// Static contains CSS and JS served at /static/.
//
//go:embed all:static
var Static embed.FS

// Templates contains html/template files under templates/.
//
//go:embed all:templates
var Templates embed.FS
