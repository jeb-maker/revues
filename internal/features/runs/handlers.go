package runs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/features/subjects"
	"github.com/jeb-maker/revues/internal/integrations/notion"
	"github.com/jeb-maker/revues/internal/integrations/webhooks"
	"github.com/jeb-maker/revues/internal/notifications"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/middleware"
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
	data := viewtemplates.PageData{Title: title}
	if user, ok := middleware.UserFromContext(r.Context()); ok {
		data.User = user
		if token := middleware.SessionTokenFromContext(r); token != "" {
			data.CSRFToken = auth.CSRFToken(token, d.SessionSecret)
		}
	}
	viewtemplates.ApplyHeaderFromContext(r, &data)
	return data
}

// PageDataTab is PageData with ActiveTab set.
func (d *Deps) PageDataTab(r *http.Request, title, activeTab string) viewtemplates.PageData {
	data := d.PageData(r, title)
	data.ActiveTab = activeTab
	return data
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

// List shows draft and in-progress review runs for the current user.
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
func (h *Runs) Show(w http.ResponseWriter, r *http.Request) {
	run, project, user, access, ok := h.loadRun(w, r)
	if !ok {
		return
	}

	h.renderRunShow(w, r, run, project, user, access, viewtemplates.RunShowData{
		Message:   r.URL.Query().Get("msg"),
		ItemError: r.URL.Query().Get("item_error"),
	})
}

// UpdateItem changes status and comment on a run item.
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
func (h *Runs) Start(w http.ResponseWriter, r *http.Request) {
	run, _, user, access, ok := h.loadRun(w, r)
	if !ok {
		return
	}
	if !CanLaunchAccess(user, access) {
		http.NotFound(w, r)
		return
	}

	if err := h.Store.StartRun(r.Context(), run.ID); err != nil {
		if errors.Is(err, store.ErrInvalidRunStatus) {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		slog.Error("start run", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/runs/"+strconv.FormatInt(run.ID, 10)+"?msg=Revue+d%C3%A9marr%C3%A9e", http.StatusSeeOther)
}

// Complete moves a run from in_progress to done.
func (h *Runs) Complete(w http.ResponseWriter, r *http.Request) {
	run, _, user, access, ok := h.loadRun(w, r)
	if !ok {
		return
	}
	if !CanCompleteAccess(user, access) {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	closingNote := strings.TrimSpace(r.FormValue("closing_note"))

	if err := h.Store.CompleteRun(r.Context(), run.ID, closingNote); err != nil {
		if errors.Is(err, store.ErrInvalidRunStatus) {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		slog.Error("complete run", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	rows, err := h.Store.ListRunExportRows(r.Context(), run.ID)
	if err != nil {
		slog.Error("list run export rows for evidence", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	csvData, err := BuildRunCSV(rows)
	if err != nil {
		slog.Error("build run csv for evidence", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := h.Store.SealRunEvidenceHash(r.Context(), run.ID, SHA256Hex(csvData)); err != nil {
		slog.Error("seal run evidence hash", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if h.Webhooks != nil {
		h.Webhooks.EmitReviewCompleted(r.Context(), run.ID)
	}

	if h.Notifications != nil {
		h.Notifications.NotifyRunCompleted(r.Context(), run.ID)
	}

	if h.isHTMX(r) {
		w.Header().Set("HX-Redirect", "/runs/"+strconv.FormatInt(run.ID, 10)+"?msg=Revue+termin%C3%A9e")
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/runs/"+strconv.FormatInt(run.ID, 10)+"?msg=Revue+termin%C3%A9e", http.StatusSeeOther)
}

// ExportCSV downloads a CSV export for a completed run.
func (h *Runs) ExportCSV(w http.ResponseWriter, r *http.Request) {
	run, project, _, _, ok := h.loadRun(w, r)
	if !ok {
		return
	}
	if run.Status != store.RunStatusDone {
		http.NotFound(w, r)
		return
	}

	rows, err := h.Store.ListRunExportRows(r.Context(), run.ID)
	if err != nil {
		slog.Error("list run export rows", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	csvData, err := BuildRunCSV(rows)
	if err != nil {
		slog.Error("build run csv", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	filename := exportCSVFilename(h.runDisplayLabel(r.Context(), run, project), run.ID)
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	if _, writeErr := w.Write(csvData); writeErr != nil {
		slog.Error("write run csv export", "err", writeErr)
	}
}

// ExportEvidence downloads a sealed evidence ZIP for a completed run.
func (h *Runs) ExportEvidence(w http.ResponseWriter, r *http.Request) {
	run, project, _, _, ok := h.loadRun(w, r)
	if !ok {
		return
	}
	if run.Status != store.RunStatusDone || strings.TrimSpace(run.EvidenceCSVSHA256) == "" {
		http.NotFound(w, r)
		return
	}

	rows, err := h.Store.ListRunExportRows(r.Context(), run.ID)
	if err != nil {
		slog.Error("list run export rows for evidence zip", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	csvData, err := BuildRunCSV(rows)
	if err != nil {
		slog.Error("build run csv for evidence zip", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if got := SHA256Hex(csvData); got != run.EvidenceCSVSHA256 {
		slog.Error("evidence csv hash mismatch", "run_id", run.ID, "sealed", run.EvidenceCSVSHA256, "got", got)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	versionInfo, err := h.Store.TemplateVersionInfo(r.Context(), run.TemplateVersionID)
	if err != nil {
		slog.Error("template version info for evidence", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	completedAt := ""
	if run.CompletedAt.Valid {
		completedAt = run.CompletedAt.String
	}
	closedBy := ""
	if run.CreatedBy.Valid {
		if u, userErr := h.Store.UserByID(r.Context(), run.CreatedBy.Int64); userErr == nil && u != nil {
			closedBy = u.Login
		}
	}

	manifest := EvidenceManifest{
		RunID:        run.ID,
		SubjectName:  project.Name,
		TemplateName: versionInfo.Name,
		Version:      versionInfo.Version,
		Status:       store.RunStatusDone,
		CompletedAt:  completedAt,
		ClosedBy:     closedBy,
		CSVSHA256:    run.EvidenceCSVSHA256,
		GeneratedAt:  completedAt,
		Attachments:  h.evidenceAttachmentRefs(r.Context(), run.ID),
	}
	zipData, err := BuildEvidenceZIP(run.ID, csvData, manifest)
	if err != nil {
		slog.Error("build evidence zip", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("preuve-revue-%d.zip", run.ID)
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	if _, writeErr := w.Write(zipData); writeErr != nil {
		slog.Error("write evidence zip", "err", writeErr)
	}
}

func exportCSVFilename(displayLabel string, runID int64) string {
	safe := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			return r
		case r == '-', r == '_':
			return r
		default:
			return '-'
		}
	}, strings.TrimSpace(displayLabel))
	if safe == "" {
		return fmt.Sprintf("revue-%d.csv", runID)
	}
	return safe + ".csv"
}

func (h *Runs) renderRunShow(w http.ResponseWriter, r *http.Request, run *store.ChecklistRun, project *store.Subject, user *store.User, access store.SubjectAccess, extra viewtemplates.RunShowData) {
	runItems, err := h.Store.ListRunItems(r.Context(), run.ID)
	if err != nil {
		slog.Error("list run items", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	nokItems, err := h.Store.ListNokRunItems(r.Context(), run.ID)
	if err != nil {
		slog.Error("list nok run items", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	versionInfo, err := h.Store.TemplateVersionInfo(r.Context(), run.TemplateVersionID)
	if err != nil {
		slog.Error("template version info", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	displayLabel := h.runDisplayLabel(r.Context(), run, project)
	pd := h.PageData(r, displayLabel)
	showAssign := pd.ShowAssign

	var members []store.SubjectMember
	if showAssign && CanAssignAccess(user, access) {
		members, err = h.Store.ListSubjectMembers(r.Context(), project.ID)
		if err != nil {
			slog.Error("list project members", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	jiraLinks := h.loadJiraLinksForItems(r.Context(), runItems)
	attachmentsByItem := h.loadAttachmentsForItems(r.Context(), runItems)

	filterSection := strings.TrimSpace(r.URL.Query().Get("section"))
	filterStatus := strings.TrimSpace(r.URL.Query().Get("status"))

	sections := uniqueSections(runItems)

	items := runItems
	if filterSection != "" {
		var filtered []store.RunItem
		for _, it := range items {
			if it.Section == filterSection {
				filtered = append(filtered, it)
			}
		}
		items = filtered
	}
	if filterStatus != "" {
		var filtered []store.RunItem
		for _, it := range items {
			if it.Status == filterStatus {
				filtered = append(filtered, it)
			}
		}
		items = filtered
	}

	sectionGroups := buildRunItemSections(items)

	// Page H1: no #id; in SimpleUI drop subject (already obvious / single-subject).
	pageTitle := store.RunDisplayLabel(versionInfo.Name, project.Name, run.CreatedAt, 0)
	if pd.SimpleUI {
		pageTitle = store.RunDisplayLabel(versionInfo.Name, "", run.CreatedAt, 0)
	}
	if pageTitle == "" {
		pageTitle = displayLabel
	}
	pd.Title = pageTitle
	pd.Breadcrumbs = viewtemplates.BCRunShow(pageTitle, pd.Labels.Run)
	pd.ActiveTab = "runs"
	data := viewtemplates.RunShowData{
		PageData:          pd,
		Subject:           project,
		Run:               run,
		RunDisplayLabel:   displayLabel,
		Items:             items,
		ItemSections:      sectionGroups,
		NokItems:          nokItems,
		Sections:          sections,
		FilterSection:     filterSection,
		FilterStatus:      filterStatus,
		JiraLinks:         jiraLinks,
		Attachments:       attachmentsByItem,
		Members:           members,
		TemplateName:      versionInfo.Name,
		VersionNum:        versionInfo.Version,
		MemberRole:        subjects.DisplayRole(access),
		CanLaunch:         CanLaunchAccess(user, access),
		CanCheck:          CanUpdateAccess(user, access),
		CanAssign:         showAssign && CanAssignAccess(user, access),
		CanLinkJira:       CanLinkJiraAccess(user, access),
		JiraConfigured:    h.jiraConfigured(r.Context()),
		CanComplete:       CanCompleteAccess(user, access),
		NotionConfigured:  h.notionConfigured(r.Context()),
		CanExportNotion:   CanCompleteAccess(user, access) && run.Status == store.RunStatusDone && strings.TrimSpace(run.NotionURL) == "",
		CanExportEvidence: run.Status == store.RunStatusDone && strings.TrimSpace(run.EvidenceCSVSHA256) != "",
		Progress:          h.progressData(run.ID, runItems),
		Message:           extra.Message,
		ItemError:         extra.ItemError,
		AssignError:       extra.AssignError,
		CompleteError:     extra.CompleteError,
		NotionExportError: extra.NotionExportError,
		ClosingNote:       extra.ClosingNote,
		Error:             extra.Error,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	statusCode := http.StatusOK
	if extra.ItemError != "" || extra.CompleteError != "" || extra.AssignError != "" || extra.NotionExportError != "" {
		statusCode = http.StatusBadRequest
	}
	w.WriteHeader(statusCode)
	if err := h.Templates.ExecuteTemplate(w, "run_show", data); err != nil {
		slog.Error("render run show", "err", err)
	}
}

func buildRunItemSections(items []store.RunItem) []viewtemplates.RunItemSectionData {
	type agg struct {
		title string
		items []store.RunItem
	}
	order := make([]string, 0, 8)
	byTitle := make(map[string]*agg)

	for _, it := range items {
		title := strings.TrimSpace(it.Section)
		if title == "" {
			title = "Sans section"
		}
		a, ok := byTitle[title]
		if !ok {
			order = append(order, title)
			a = &agg{title: title}
			byTitle[title] = a
		}
		a.items = append(a.items, it)
	}

	out := make([]viewtemplates.RunItemSectionData, 0, len(order))
	for _, title := range order {
		a := byTitle[title]
		okCount := 0
		nonOK := 0
		for _, it := range a.items {
			switch it.Status {
			case store.RunItemStatusOK, store.RunItemStatusNA:
				okCount++
			default:
				nonOK++
			}
		}
		out = append(out, viewtemplates.RunItemSectionData{
			Title:      a.title,
			Items:      a.items,
			Total:      len(a.items),
			OKCount:    okCount,
			NonOKCount: nonOK,
			AllOKOrNA:  nonOK == 0,
		})
	}
	return out
}

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

func (h *Runs) progressData(runID int64, runItems []store.RunItem) viewtemplates.RunProgressData {
	done, total := Progress(runItems)
	percent := 0
	if total > 0 {
		percent = done * 100 / total
	}
	return viewtemplates.RunProgressData{
		RunID:   runID,
		Done:    done,
		Total:   total,
		Percent: percent,
	}
}

func uniqueSections(items []store.RunItem) []string {
	seen := make(map[string]bool)
	var sections []string
	for _, it := range items {
		if it.Section != "" && !seen[it.Section] {
			seen[it.Section] = true
			sections = append(sections, it.Section)
		}
	}
	return sections
}

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
		completeStatus := viewtemplates.RunCompleteStatusData{
			Run:      run,
			NokItems: nokItemsFromRunItems(runItems),
			Progress: progress,
		}
		if err := h.Templates.ExecuteTemplate(w, "run_complete_status_oob_fragment", completeStatus); err != nil {
			slog.Error("render run complete status oob fragment", "err", err)
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
