package runs

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/middleware"
)

func (h *Runs) loadSubjectForLaunch(w http.ResponseWriter, r *http.Request) (*store.Subject, *store.User, store.SubjectAccess, bool) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return nil, nil, store.SubjectAccess{}, false
	}

	projectID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return nil, nil, store.SubjectAccess{}, false
	}

	project, err := h.Store.SubjectByID(r.Context(), projectID)
	if errors.Is(err, store.ErrSubjectNotFound) {
		http.NotFound(w, r)
		return nil, nil, store.SubjectAccess{}, false
	}
	if err != nil {
		slog.Error("load project", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, store.SubjectAccess{}, false
	}

	access, err := h.Store.ResolveSubjectAccess(r.Context(), user.ID, projectID, user.Role)
	if err != nil {
		slog.Error("resolve subject access", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, store.SubjectAccess{}, false
	}

	if !CanLaunchAccess(user, access) {
		http.NotFound(w, r)
		return nil, nil, store.SubjectAccess{}, false
	}

	return project, user, access, true
}

func (h *Runs) loadRun(w http.ResponseWriter, r *http.Request) (*store.ChecklistRun, *store.Subject, *store.User, store.SubjectAccess, bool) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return nil, nil, nil, store.SubjectAccess{}, false
	}

	runID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return nil, nil, nil, store.SubjectAccess{}, false
	}

	run, err := h.Store.RunByID(r.Context(), runID)
	if errors.Is(err, store.ErrRunNotFound) {
		http.NotFound(w, r)
		return nil, nil, nil, store.SubjectAccess{}, false
	}
	if err != nil {
		slog.Error("load run", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, nil, store.SubjectAccess{}, false
	}

	project, err := h.Store.SubjectByID(r.Context(), run.SubjectID)
	if errors.Is(err, store.ErrSubjectNotFound) {
		http.NotFound(w, r)
		return nil, nil, nil, store.SubjectAccess{}, false
	}
	if err != nil {
		slog.Error("load run project", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, nil, store.SubjectAccess{}, false
	}

	access, err := h.Store.ResolveSubjectAccess(r.Context(), user.ID, project.ID, user.Role)
	if err != nil {
		slog.Error("resolve subject access", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, nil, store.SubjectAccess{}, false
	}

	if !CanViewAccess(access) {
		http.NotFound(w, r)
		return nil, nil, nil, store.SubjectAccess{}, false
	}

	return run, project, user, access, true
}

func (h *Runs) runDisplayLabel(ctx context.Context, run *store.ChecklistRun, subject *store.Subject) string {
	versionInfo, err := h.Store.TemplateVersionInfo(ctx, run.TemplateVersionID)
	if err != nil {
		return RunDisplayLabel("", subject.Name, run.CreatedAt, run.ID)
	}
	return RunDisplayLabel(versionInfo.Name, subject.Name, run.CreatedAt, run.ID)
}

func (h *Runs) isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") != ""
}
