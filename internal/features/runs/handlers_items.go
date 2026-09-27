package runs

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	viewtemplates "github.com/jeb-maker/revues/internal/web/templates"

	"github.com/go-chi/chi/v5"
	"github.com/jeb-maker/revues/internal/store"
)

func (h *Runs) UpdateItem(w http.ResponseWriter, r *http.Request) {
	run, project, user, access, ok := h.loadRun(w, r)
	if !ok {
		return
	}
	if !CanUpdateAccess(user, access) {
		http.NotFound(w, r)
		return
	}
	if run.Status != store.RunStatusInProgress {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	itemID, err := strconv.ParseInt(chi.URLParam(r, "itemId"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	existing, err := h.Store.RunItemByID(r.Context(), run.ID, itemID)
	if errors.Is(err, store.ErrRunItemNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		slog.Error("load run item before update", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	status := strings.TrimSpace(r.FormValue("status"))
	comment := strings.TrimSpace(r.FormValue("comment"))

	if err := ValidateUpdate(status, comment); err != nil {
		switch {
		case errors.Is(err, ErrCommentRequired):
			if h.isHTMX(r) {
				h.renderRunItemHTMXError(w, r, run, project, user, access, itemID, "Un commentaire est obligatoire pour le statut Non validé.", "")
				return
			}
			h.renderRunShow(w, r, run, project, user, access, viewtemplates.RunShowData{
				ItemError: "Un commentaire est obligatoire pour le statut Non validé.",
			})
		case errors.Is(err, ErrInvalidStatus):
			if h.isHTMX(r) {
				h.renderRunItemHTMXError(w, r, run, project, user, access, itemID, "Statut invalide.", "")
				return
			}
			h.renderRunShow(w, r, run, project, user, access, viewtemplates.RunShowData{
				ItemError: "Statut invalide.",
			})
		default:
			http.Error(w, "Bad Request", http.StatusBadRequest)
		}
		return
	}

	if err := h.Store.UpdateRunItemStatus(r.Context(), run.ID, itemID, user.ID, status, comment); err != nil {
		if errors.Is(err, store.ErrRunItemNotFound) {
			http.NotFound(w, r)
			return
		}
		slog.Error("update run item", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if h.Webhooks != nil && status == store.RunItemStatusNOK && existing.Status != store.RunItemStatusNOK {
		h.Webhooks.EmitReviewItemNOK(r.Context(), run.ID, itemID)
	}

	if h.isHTMX(r) {
		h.renderRunItemHTMXSuccess(w, r, run, project, user, access, itemID, "", "")
		return
	}

	http.Redirect(w, r, "/runs/"+strconv.FormatInt(run.ID, 10)+"?msg=Point+mis+%C3%A0+jour", http.StatusSeeOther)
}

// ShowItem displays a run item and its status change history.
func (h *Runs) ShowItem(w http.ResponseWriter, r *http.Request) {
	run, project, user, access, ok := h.loadRun(w, r)
	if !ok {
		return
	}

	itemID, err := strconv.ParseInt(chi.URLParam(r, "itemId"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	h.renderRunItemShow(w, r, run, project, user, access, itemID, viewtemplates.RunItemShowData{})
}

// AssignItem sets or clears assignee on a run item.
func (h *Runs) AssignItem(w http.ResponseWriter, r *http.Request) {
	run, project, user, access, ok := h.loadRun(w, r)
	if !ok {
		return
	}
	if !CanAssignAccess(user, access) {
		http.NotFound(w, r)
		return
	}
	if run.Status != store.RunStatusInProgress {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	itemID, err := strconv.ParseInt(chi.URLParam(r, "itemId"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var assigneeID *int64
	// "0" = sentinelle « Non assigné » (l'option vide de mb-select vaut aussi désassignation).
	if raw := strings.TrimSpace(r.FormValue("assignee_id")); raw != "" && raw != "0" {
		id, parseErr := strconv.ParseInt(raw, 10, 64)
		if parseErr != nil {
			if h.isHTMX(r) {
				h.renderRunItemHTMXError(w, r, run, project, user, access, itemID, "", "Assigné invalide.")
				return
			}
			h.renderRunShow(w, r, run, project, user, access, viewtemplates.RunShowData{
				AssignError: "Assigné invalide.",
			})
			return
		}
		assigneeID = &id
	}

	if err := h.Store.AssignRunItem(r.Context(), run.ID, itemID, assigneeID); err != nil {
		if errors.Is(err, store.ErrInvalidAssignee) {
			if h.isHTMX(r) {
				h.renderRunItemHTMXError(w, r, run, project, user, access, itemID, "", "Le membre doit appartenir à l'organisation.")
				return
			}
			h.renderRunShow(w, r, run, project, user, access, viewtemplates.RunShowData{
				AssignError: "Le membre doit appartenir au sujet.",
			})
			return
		}
		if errors.Is(err, store.ErrRunItemNotFound) {
			http.NotFound(w, r)
			return
		}
		slog.Error("assign run item", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if assigneeID != nil && h.Notifications != nil {
		h.Notifications.NotifyItemAssigned(r.Context(), run.ID, itemID)
	}

	if h.isHTMX(r) {
		h.renderRunItemHTMXSuccess(w, r, run, project, user, access, itemID, "", "")
		return
	}

	http.Redirect(w, r, "/runs/"+strconv.FormatInt(run.ID, 10)+"?msg=Assignation+enregistr%C3%A9e", http.StatusSeeOther)
}

// Start moves a run from draft to in_progress.
func (h *Runs) renderRunItemHTMXSuccess(w http.ResponseWriter, r *http.Request, run *store.ChecklistRun, project *store.Subject, user *store.User, access store.SubjectAccess, itemID int64, itemErr, assignErr string) {
	runItems, err := h.Store.ListRunItems(r.Context(), run.ID)
	if err != nil {
		slog.Error("list run items for htmx", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	item, ok := findRunItem(runItems, itemID)
	if !ok {
		http.NotFound(w, r)
		return
	}

	h.renderRunItemHTMX(w, r, run, project, user, access, item, runItems, itemErr, assignErr, http.StatusOK)
}

func (h *Runs) renderRunItemHTMXError(w http.ResponseWriter, r *http.Request, run *store.ChecklistRun, project *store.Subject, user *store.User, access store.SubjectAccess, itemID int64, itemErr, assignErr string) {
	item, err := h.Store.RunItemByID(r.Context(), run.ID, itemID)
	if errors.Is(err, store.ErrRunItemNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		slog.Error("load run item for htmx error", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	runItems, err := h.Store.ListRunItems(r.Context(), run.ID)
	if err != nil {
		slog.Error("list run items for htmx error", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.renderRunItemHTMX(w, r, run, project, user, access, *item, runItems, itemErr, assignErr, http.StatusBadRequest)
}

func (h *Runs) renderRunItemHTMX(w http.ResponseWriter, r *http.Request, run *store.ChecklistRun, project *store.Subject, user *store.User, access store.SubjectAccess, item store.RunItem, runItems []store.RunItem, itemErr, assignErr string, statusCode int) {
	pd := h.PageData(r, "")
	showAssign := pd.ShowAssign
	canAssign := showAssign && CanAssignAccess(user, access)

	var members []store.SubjectMember
	var err error
	if canAssign {
		members, err = h.Store.ListSubjectMembers(r.Context(), project.ID)
		if err != nil {
			slog.Error("list project members for htmx", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	row := viewtemplates.RunItemRowData{
		RunID:       run.ID,
		RunStatus:   run.Status,
		Item:        item,
		SectionID:   viewtemplates.SectionIDForTitle(item.Section),
		Members:     members,
		CSRFToken:   pd.CSRFToken,
		CanCheck:    CanUpdateAccess(user, access),
		CanAssign:   canAssign,
		ShowAssign:  showAssign,
		ItemError:   itemErr,
		AssignError: assignErr,
	}
	if link, ok := h.loadJiraLinksForItems(r.Context(), []store.RunItem{item})[item.ID]; ok {
		row.JiraLink = link
	}
	if att, ok := h.loadAttachmentsForItems(r.Context(), []store.RunItem{item})[item.ID]; ok {
		row.Attachment = att
	}
	progress := h.progressData(run.ID, runItems)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if statusCode < 400 {
		w.Header().Set("HX-Trigger", `{"toast:success":{"message":"Point mis à jour"}}`)
	}
	w.WriteHeader(statusCode)

	var buf bytes.Buffer
	if err := h.Templates.ExecuteTemplate(&buf, "run_item_row_fragment", row); err != nil {
		slog.Error("render run item row fragment", "err", err)
		return
	}
	if _, err := w.Write(buf.Bytes()); err != nil {
		slog.Error("write run item row fragment", "err", err)
		return
	}
	if err := h.Templates.ExecuteTemplate(w, "run_progress_oob_fragment", progress); err != nil {
		slog.Error("render run progress oob fragment", "err", err)
	}
	if run.Status == store.RunStatusInProgress && CanCompleteAccess(user, access) {
		completeStatus := h.completeStatusData(r, run, runItems, "", "")
		if err := h.Templates.ExecuteTemplate(w, "run_complete_section_oob_fragment", completeStatus); err != nil {
			slog.Error("render run complete section oob fragment", "err", err)
		}
	}
}

func nokItemsFromRunItems(runItems []store.RunItem) []store.RunItem {
	var nok []store.RunItem
	for _, item := range runItems {
		if item.Status == store.RunItemStatusNOK {
			nok = append(nok, item)
		}
	}
	return nok
}

func findRunItem(runItems []store.RunItem, itemID int64) (store.RunItem, bool) {
	for _, item := range runItems {
		if item.ID == itemID {
			return item, true
		}
	}
	return store.RunItem{}, false
}

func parseRunListFilters(r *http.Request) (status, query string) {
	q := r.URL.Query()
	rawStatus := strings.TrimSpace(q.Get("status"))
	if store.ValidRunListStatus(rawStatus) {
		status = rawStatus
	}
	query = strings.TrimSpace(q.Get("q"))
	return status, query
}

func parseListPage(r *http.Request) int {
	page, err := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("page")))
	if err != nil || page < 1 {
		return 1
	}
	return page
}
