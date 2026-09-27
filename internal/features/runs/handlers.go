package runs

import (
	"html/template"
	"net/http"

	"github.com/jeb-maker/revues/internal/integrations/notion"
	"github.com/jeb-maker/revues/internal/integrations/webhooks"
	"github.com/jeb-maker/revues/internal/notifications"
	viewtemplates "github.com/jeb-maker/revues/internal/web/templates"
)

// Deps holds shared dependencies for the runs HTTP handlers.
//
// This mirrors internal/web/handlerdeps.HandlerDeps but is local to the runs
// feature package to avoid an import cycle.
type Deps struct {
	Templates     *template.Template
	Store         RunStore
	SessionSecret string
}

// PageData builds shared view data with user and CSRF from the request context.
func (d *Deps) PageData(r *http.Request, title string) viewtemplates.PageData {
	return viewtemplates.NewPageData(r, title, d.SessionSecret)
}

// PageDataTab is PageData with ActiveTab set.
func (d *Deps) PageDataTab(r *http.Request, title, activeTab string) viewtemplates.PageData {
	return viewtemplates.NewPageDataTab(r, title, activeTab, d.SessionSecret)
}

// Runs handles review launch wizard and run lifecycle.
type Runs struct {
	Deps
	EncryptionKey  []byte
	AttachmentsDir string
	BaseURL        string
	NotionClient   *notion.Client
	Webhooks       *webhooks.Dispatcher
	Notifications  *notifications.Service
}
