package mytasks

import (
	"html/template"
	"log/slog"
	"net/http"
	"strings"

	runs "github.com/jeb-maker/revues/internal/features/runs"
	"github.com/jeb-maker/revues/internal/web/middleware"
	viewtemplates "github.com/jeb-maker/revues/internal/web/templates"
)

// Deps holds shared dependencies for the mytasks HTTP handlers.
//
// This mirrors internal/web/handlerdeps.HandlerDeps but is local to the
// mytasks feature package to avoid an import cycle.
type Deps struct {
	Templates     *template.Template
	Store         AssignedTaskStore
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

// MyTasks lists run items assigned to the current user.
type MyTasks struct {
	Deps
}

// List shows assigned tasks with optional search and status filters.
func (h *MyTasks) List(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	filterStatus, filterQuery := parseMyTasksFilters(r)

	tasks, err := h.Store.ListAssignedRunItems(r.Context(), user.ID, filterStatus, filterQuery)
	if err != nil {
		slog.Error("list assigned run items", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := viewtemplates.MyTasksData{
		PageData:         h.PageDataTab(r, "Mes tâches", "tasks"),
		Tasks:            tasks,
		FilterQuery:      filterQuery,
		FilterStatus:     filterStatus,
		HasActiveFilters: filterQuery != "" || filterStatus != "",
	}
	data.Breadcrumbs = viewtemplates.BCTasks()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "my_tasks", data); err != nil {
		slog.Error("render my tasks", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func parseMyTasksFilters(r *http.Request) (status, query string) {
	rawStatus := strings.TrimSpace(r.URL.Query().Get("status"))
	if runs.ValidStatus(rawStatus) {
		status = rawStatus
	}
	query = strings.TrimSpace(r.URL.Query().Get("q"))
	return status, query
}
