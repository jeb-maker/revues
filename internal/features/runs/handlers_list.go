package runs

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	viewtemplates "github.com/jeb-maker/revues/internal/web/templates"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/features/subjects"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/middleware"
)

func (h *Runs) List(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	admin := auth.HasMinRole(user.Role, auth.RoleAdmin)
	subjectItems, err := h.Store.ListSubjects(r.Context(), user.ID, admin, "")
	if err != nil {
		slog.Error("list subjects for runs page", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	filterStatus, filterQuery := parseRunListFilters(r)
	page := parseListPage(r)
	pageSize := store.FilteredRunsPageSize
	offset := (page - 1) * pageSize
	runs, total, err := h.Store.ListFilteredRunSummaries(r.Context(), user.ID, admin, filterStatus, filterQuery, pageSize, offset)
	if err != nil {
		slog.Error("list filtered runs", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	if totalPages > 0 && page > totalPages {
		page = totalPages
		offset = (page - 1) * pageSize
		runs, total, err = h.Store.ListFilteredRunSummaries(r.Context(), user.ID, admin, filterStatus, filterQuery, pageSize, offset)
		if err != nil {
			slog.Error("list filtered runs", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	orgRole, orgMember, _ := h.Store.OrganizationMemberRole(r.Context(), 0, user.ID)
	if org, ok := middleware.OrganizationFromContext(r.Context()); ok {
		orgRole, orgMember, _ = h.Store.OrganizationMemberRole(r.Context(), org.ID, user.ID)
	}

	hasSubjects := len(subjectItems) > 0
	canLaunch := subjects.CanLaunchRun(user, orgMember) && (hasSubjects || subjects.CanCreateSubject(user))
	pagination := viewtemplates.NewPagination(page, pageSize, total, func(p int) string {
		return viewtemplates.RunsListURL(filterStatus, filterQuery, p)
	})
	pd := h.PageDataTab(r, "", "runs")
	data := viewtemplates.RunsListData{
		PageData:          viewtemplates.ApplyPageMeta(pd, viewtemplates.BCRevues(pd.Labels.Run)),
		Runs:              runs,
		FilterQuery:       filterQuery,
		FilterStatus:      filterStatus,
		HasActiveFilters:  filterQuery != "" || filterStatus != "",
		HasSubjects:       hasSubjects,
		CanCreate:         subjects.CanCreateSubject(user),
		CanLaunch:         canLaunch,
		CanManageOrgUsers: subjects.CanManageOrgUsers(user, orgRole, orgMember),
		Pagination:        pagination,
		Message:           r.URL.Query().Get("msg"),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "runs_list", data); err != nil {
		slog.Error("render runs list", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Create stores a new run with item snapshot.
func (h *Runs) Create(w http.ResponseWriter, r *http.Request) {
	project, user, access, ok := h.loadSubjectForLaunch(w, r)
	if !ok {
		return
	}
	_ = access
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	templateID, err := strconv.ParseInt(r.FormValue("template_id"), 10, 64)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	template, err := h.Store.ChecklistTemplateByID(r.Context(), templateID)
	if errors.Is(err, store.ErrChecklistTemplateNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		slog.Error("load template for run create", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if template.ArchivedAt.Valid {
		http.NotFound(w, r)
		return
	}

	matches, err := h.Store.TemplateMatchesSubject(r.Context(), project.ID, template.ID)
	if err != nil {
		slog.Error("check template matches project", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !matches {
		http.NotFound(w, r)
		return
	}

	run, err := h.Store.CreateChecklistRun(r.Context(), project.ID, template.ID, user.ID)
	if err != nil {
		slog.Error("create checklist run", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/runs/"+strconv.FormatInt(run.ID, 10), http.StatusSeeOther)
}

// Show displays run detail and snapshot items.
