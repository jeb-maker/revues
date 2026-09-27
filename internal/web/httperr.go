package web

import (
	"html/template"
	"log"
	"net/http"

	viewtemplates "github.com/jeb-maker/revues/internal/web/templates"
)

// ErrorPageData is view data for branded HTTP error pages.
type ErrorPageData struct {
	viewtemplates.PageData
	Message string
}

func writeErrorPage(w http.ResponseWriter, r *http.Request, tpl *template.Template, sessionSecret string, status int, title, message string) {
	pd := viewtemplates.NewPageData(r, title, sessionSecret)
	pd.Breadcrumbs = []viewtemplates.Breadcrumb{{Label: title}}
	w.WriteHeader(status)
	if err := tpl.ExecuteTemplate(w, "error", ErrorPageData{PageData: pd, Message: message}); err != nil {
		log.Printf("render error page: %v", err)
	}
}
