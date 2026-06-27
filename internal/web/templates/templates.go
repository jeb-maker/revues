package templates

import (
	"fmt"
	"html/template"
	"io/fs"

	webassets "github.com/jeb-maker/revues/web"
)

// PageData is shared view data for HTML pages.
type PageData struct {
	Title string
}

// Parse loads layout and page templates from the embedded filesystem.
func Parse() (*template.Template, error) {
	root, err := fs.Sub(webassets.Templates, "templates")
	if err != nil {
		return nil, fmt.Errorf("templates root: %w", err)
	}

	tpl, err := template.ParseFS(root,
		"layouts/base.html",
		"pages/*.html",
	)
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	return tpl, nil
}
