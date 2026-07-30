package templates

import (
	"net/http"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/web/middleware"
)

// NewPageData builds shared view data with user, CSRF and header context.
func NewPageData(r *http.Request, title, sessionSecret string) PageData {
	data := PageData{Title: title}
	if user, ok := middleware.UserFromContext(r.Context()); ok {
		data.User = user
		if token := middleware.SessionTokenFromContext(r); token != "" {
			data.CSRFToken = auth.CSRFToken(token, sessionSecret)
		}
	}
	ApplyHeaderFromContext(r, &data)
	return data
}

// NewPageDataTab is NewPageData with ActiveTab set.
func NewPageDataTab(r *http.Request, title, activeTab, sessionSecret string) PageData {
	data := NewPageData(r, title, sessionSecret)
	data.ActiveTab = activeTab
	return data
}
