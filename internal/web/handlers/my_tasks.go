package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jeb-maker/revues/internal/auth"
	runs "github.com/jeb-maker/revues/internal/features/runs"
	"github.com/jeb-maker/revues/internal/web/middleware"
	viewtemplates "github.com/jeb-maker/revues/internal/web/templates"
)

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
	if statusFilter != "" && !runs.ValidStatus(statusFilter) {
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
