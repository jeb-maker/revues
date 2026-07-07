package mytasks

import (
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/items"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/middleware"
	viewtemplates "github.com/jeb-maker/revues/internal/web/templates"
)

// Deps holds shared dependencies for the mytasks HTTP handlers.
//
// This mirrors internal/web/handlers.Deps but is local to the mytasks feature
// package to avoid an import cycle (features/mytasks must not import
// internal/web/handlers). A follow-up issue may extract a shared base Deps.
type Deps struct {
	Templates     *template.Template
	Store         *store.Store
	SessionSecret string
}

// PageData builds shared view data with user and CSRF from the request context.
func (d *Deps) PageData(r *http.Request, title string) viewtemplates.PageData {
	data := viewtemplates.PageData{Title: title}
	if user, ok := middleware.UserFromContext(r.Context()); ok {
		data.User = user
		if token := middleware.SessionTokenFromContext(r); token != "" {
			data.CSRFToken = auth.CSRFToken(token, d.SessionSecret)
		}
	}
	return data
}

// PageDataTab is PageData with ActiveTab set.
func (d *Deps) PageDataTab(r *http.Request, title, activeTab string) viewtemplates.PageData {
	data := d.PageData(r, title)
	data.ActiveTab = activeTab
	return data
}

// MyTasks lists run items assigned to the current user.
type MyTasks struct {
	Deps
}

// List shows assigned tasks with optional project and status filters.
func (h *MyTasks) List(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	projectFilter := int64(0)
	if raw := r.URL.Query().Get("project"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err == nil {
			projectFilter = id
		}
	}
	statusFilter := r.URL.Query().Get("status")
	if statusFilter != "" && !items.ValidStatus(statusFilter) {
		statusFilter = ""
	}

	tasks, err := h.Store.ListAssignedRunItems(r.Context(), user.ID, projectFilter, statusFilter)
	if err != nil {
		slog.Error("list assigned run items", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	admin := auth.HasMinRole(user.Role, auth.RoleAdmin)
	projects, err := h.Store.ListProjects(r.Context(), user.ID, admin)
	if err != nil {
		slog.Error("list projects for my tasks", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := viewtemplates.MyTasksData{
		PageData:        h.PageDataTab(r, "Mes tâches", "tasks"),
		Tasks:           tasks,
		Projects:        projects,
		FilterProjectID: projectFilter,
		FilterStatus:    statusFilter,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "my_tasks", data); err != nil {
		slog.Error("render my tasks", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
