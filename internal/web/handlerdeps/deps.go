// Package handlerdeps provides a shared HandlerDeps struct so that
// internal/web/router.go and internal/features/admin/* can both use it
// without creating an import cycle (router imports features, features
// would import web back).
package handlerdeps

import (
	"html/template"
	"net/http"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/middleware"
	"github.com/jeb-maker/revues/internal/web/templates"
)

type HandlerDeps struct {
	Templates     *template.Template
	Store         *store.Store
	SessionSecret string
}

func (d *HandlerDeps) PageData(r *http.Request, title string) templates.PageData {
	data := templates.PageData{Title: title}
	if user, ok := middleware.UserFromContext(r.Context()); ok {
		data.User = user
		if token := middleware.SessionTokenFromContext(r); token != "" {
			data.CSRFToken = auth.CSRFToken(token, d.SessionSecret)
		}
	}
	return data
}

func (d *HandlerDeps) PageDataTab(r *http.Request, title, activeTab string) templates.PageData {
	data := d.PageData(r, title)
	data.ActiveTab = activeTab
	return data
}
