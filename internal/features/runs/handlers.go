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
	runs, err := h.Store.ListFilteredRunSummaries(r.Context(), user.ID, admin, filterStatus, filterQuery)
	if err != nil {
		slog.Error("list filtered runs", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	orgRole, orgMember, _ := h.Store.OrganizationMemberRole(r.Context(), 0, user.ID)
	if org, ok := middleware.OrganizationFromContext(r.Context()); ok {
		orgRole, orgMember, _ = h.Store.OrganizationMemberRole(r.Context(), org.ID, user.ID)
	}

	hasSubjects := len(subjectItems) > 0
	canLaunch := subjects.CanLaunchRun(user, orgMember) && (hasSubjects || subjects.CanCreateSubject(user))
	data := viewtemplates.RunsListData{
		PageData:          h.PageDataTab(r, "Revues", "runs"),
		Runs:              runs,
		FilterQuery:       filterQuery,
		FilterStatus:      filterStatus,
		HasActiveFilters:  filterQuery != "" || filterStatus != "",
		HasSubjects:       hasSubjects,
		CanCreate:         subjects.CanCreateSubject(user),
		CanLaunch:         canLaunch,
		CanManageOrgUsers: subjects.CanManageOrgUsers(user, orgRole, orgMember),
		Message:           r.URL.Query().Get("msg"),
	}
	data.Breadcrumbs = viewtemplates.BCRevues()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "runs_list", data); err != nil {
		slog.Error("render runs list", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Create stores a new run with item snapshot.
func (h *Runs) Create(w http.ResponseWriter, r *http.Request) {
	project, user, _, isMember, ok := h.loadSubjectForLaunch(w, r)
	if !ok {
		return
	}
	if !CanLaunch(user, isMember) {
		http.NotFound(w, r)
		return
	}
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

	http.Redirect(w, r, "/runs/"+strconv.FormatInt(run.ID, 10)+"?msg=Revue+cr%C3%A9%C3%A9e", http.StatusSeeOther)
}

// Show displays run detail and snapshot items.
func (h *Runs) Show(w http.ResponseWriter, r *http.Request) {
	run, project, user, memberRole, isMember, ok := h.loadRun(w, r)
	if !ok {
		return
	}

	h.renderRunShow(w, r, run, project, user, memberRole, isMember, viewtemplates.RunShowData{
		Message:   r.URL.Query().Get("msg"),
		ItemError: r.URL.Query().Get("item_error"),
	})
}

// UpdateItem changes status and comment on a run item.
func (h *Runs) UpdateItem(w http.ResponseWriter, r *http.Request) {
	run, project, user, memberRole, isMember, ok := h.loadRun(w, r)
	if !ok {
		return
	}
	if !CanUpdate(user, isMember) {
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
				h.renderRunItemHTMXError(w, r, run, project, user, memberRole, itemID, "Un commentaire est obligatoire pour le statut Non validé.", "")
				return
			}
			h.renderRunShow(w, r, run, project, user, memberRole, isMember, viewtemplates.RunShowData{
				ItemError: "Un commentaire est obligatoire pour le statut Non validé.",
			})
		case errors.Is(err, ErrInvalidStatus):
			if h.isHTMX(r) {
				h.renderRunItemHTMXError(w, r, run, project, user, memberRole, itemID, "Statut invalide.", "")
				return
			}
			h.renderRunShow(w, r, run, project, user, memberRole, isMember, viewtemplates.RunShowData{
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
		h.renderRunItemHTMXSuccess(w, r, run, project, user, memberRole, itemID, "", "")
		return
	}

	http.Redirect(w, r, "/runs/"+strconv.FormatInt(run.ID, 10)+"?msg=Point+mis+%C3%A0+jour", http.StatusSeeOther)
}

// ShowItem displays a run item and its status change history.
func (h *Runs) ShowItem(w http.ResponseWriter, r *http.Request) {
	run, project, user, memberRole, _, ok := h.loadRun(w, r)
	if !ok {
		return
	}

	itemID, err := strconv.ParseInt(chi.URLParam(r, "itemId"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	h.renderRunItemShow(w, r, run, project, user, memberRole, itemID, viewtemplates.RunItemShowData{})
}

// AssignItem sets or clears assignee on a run item.
func (h *Runs) AssignItem(w http.ResponseWriter, r *http.Request) {
	run, project, user, memberRole, isMember, ok := h.loadRun(w, r)
	if !ok {
		return
	}
	orgRole, orgMember, err := h.Store.OrganizationMemberRole(r.Context(), project.OrganizationID, user.ID)
	if err != nil {
		slog.Error("caller org role", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !CanAssign(user, orgRole, orgMember) {
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
	if raw := strings.TrimSpace(r.FormValue("assignee_id")); raw != "" {
		id, parseErr := strconv.ParseInt(raw, 10, 64)
		if parseErr != nil {
			if h.isHTMX(r) {
				h.renderRunItemHTMXError(w, r, run, project, user, memberRole, itemID, "", "Assigné invalide.")
				return
			}
			h.renderRunShow(w, r, run, project, user, memberRole, isMember, viewtemplates.RunShowData{
				AssignError: "Assigné invalide.",
			})
			return
		}
		assigneeID = &id
	}

	if err := h.Store.AssignRunItem(r.Context(), run.ID, itemID, assigneeID); err != nil {
		if errors.Is(err, store.ErrInvalidAssignee) {
			if h.isHTMX(r) {
				h.renderRunItemHTMXError(w, r, run, project, user, memberRole, itemID, "", "Le membre doit appartenir à l'organisation.")
				return
			}
			h.renderRunShow(w, r, run, project, user, memberRole, isMember, viewtemplates.RunShowData{
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
		h.renderRunItemHTMXSuccess(w, r, run, project, user, memberRole, itemID, "", "")
		return
	}

	http.Redirect(w, r, "/runs/"+strconv.FormatInt(run.ID, 10)+"?msg=Assignation+enregistr%C3%A9e", http.StatusSeeOther)
}

// Start moves a run from draft to in_progress.
func (h *Runs) Start(w http.ResponseWriter, r *http.Request) {
	run, _, user, _, isMember, ok := h.loadRun(w, r)
	if !ok {
		return
	}
	if !CanLaunch(user, isMember) {
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
	run, project, user, _, _, ok := h.loadRun(w, r)
	if !ok {
		return
	}
	orgRole, orgMember, err := h.Store.OrganizationMemberRole(r.Context(), project.OrganizationID, user.ID)
	if err != nil {
		slog.Error("caller org role", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !CanComplete(user, orgRole, orgMember) {
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
	run, project, _, _, _, ok := h.loadRun(w, r)
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

func (h *Runs) renderRunShow(w http.ResponseWriter, r *http.Request, run *store.ChecklistRun, project *store.Subject, user *store.User, memberRole string, isMember bool, extra viewtemplates.RunShowData) {
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

	orgRole, orgMember, err := h.Store.OrganizationMemberRole(r.Context(), project.OrganizationID, user.ID)
	if err != nil {
		slog.Error("caller org role", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var members []store.SubjectMember
	if CanAssign(user, orgRole, orgMember) {
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

	displayLabel := h.runDisplayLabel(r.Context(), run, project)
	pd := h.PageData(r, displayLabel)
	pd.Breadcrumbs = viewtemplates.BCRunShow(displayLabel)
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
		MemberRole:        memberRole,
		CanLaunch:         CanLaunch(user, isMember),
		CanCheck:          CanUpdate(user, isMember),
		CanAssign:         CanAssign(user, orgRole, orgMember),
		CanLinkJira:       CanLinkJira(user, isMember),
		JiraConfigured:    h.jiraConfigured(r.Context()),
		CanComplete:       CanComplete(user, orgRole, orgMember),
		NotionConfigured:  h.notionConfigured(r.Context()),
		CanExportNotion:   CanComplete(user, orgRole, orgMember) && run.Status == store.RunStatusDone && strings.TrimSpace(run.NotionURL) == "",
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

func (h *Runs) loadSubjectForLaunch(w http.ResponseWriter, r *http.Request) (*store.Subject, *store.User, string, bool, bool) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return nil, nil, "", false, false
	}

	projectID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return nil, nil, "", false, false
	}

	project, err := h.Store.SubjectByID(r.Context(), projectID)
	if errors.Is(err, store.ErrSubjectNotFound) {
		http.NotFound(w, r)
		return nil, nil, "", false, false
	}
	if err != nil {
		slog.Error("load project", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, "", false, false
	}

	memberRole, isMember, err := h.Store.MemberRole(r.Context(), projectID, user.ID)
	if err != nil {
		slog.Error("member role", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, "", false, false
	}

	if !CanLaunch(user, isMember) {
		http.NotFound(w, r)
		return nil, nil, "", false, false
	}

	return project, user, memberRole, isMember, true
}

func (h *Runs) loadRun(w http.ResponseWriter, r *http.Request) (*store.ChecklistRun, *store.Subject, *store.User, string, bool, bool) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return nil, nil, nil, "", false, false
	}

	runID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return nil, nil, nil, "", false, false
	}

	run, err := h.Store.RunByID(r.Context(), runID)
	if errors.Is(err, store.ErrRunNotFound) {
		http.NotFound(w, r)
		return nil, nil, nil, "", false, false
	}
	if err != nil {
		slog.Error("load run", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, nil, "", false, false
	}

	project, err := h.Store.SubjectByID(r.Context(), run.SubjectID)
	if errors.Is(err, store.ErrSubjectNotFound) {
		http.NotFound(w, r)
		return nil, nil, nil, "", false, false
	}
	if err != nil {
		slog.Error("load run project", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, nil, "", false, false
	}

	memberRole, isMember, err := h.Store.MemberRole(r.Context(), project.ID, user.ID)
	if err != nil {
		slog.Error("member role", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, nil, "", false, false
	}

	if !CanView(user, isMember) {
		http.NotFound(w, r)
		return nil, nil, nil, "", false, false
	}

	return run, project, user, memberRole, isMember, true
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

func (h *Runs) renderRunItemHTMXSuccess(w http.ResponseWriter, r *http.Request, run *store.ChecklistRun, project *store.Subject, user *store.User, memberRole string, itemID int64, itemErr, assignErr string) {
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

	h.renderRunItemHTMX(w, r, run, project, user, memberRole, item, runItems, itemErr, assignErr, http.StatusOK)
}

func (h *Runs) renderRunItemHTMXError(w http.ResponseWriter, r *http.Request, run *store.ChecklistRun, project *store.Subject, user *store.User, memberRole string, itemID int64, itemErr, assignErr string) {
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

	h.renderRunItemHTMX(w, r, run, project, user, memberRole, *item, runItems, itemErr, assignErr, http.StatusBadRequest)
}

func (h *Runs) renderRunItemHTMX(w http.ResponseWriter, r *http.Request, run *store.ChecklistRun, project *store.Subject, user *store.User, memberRole string, item store.RunItem, runItems []store.RunItem, itemErr, assignErr string, statusCode int) {
	orgRole, orgMember, err := h.Store.OrganizationMemberRole(r.Context(), project.OrganizationID, user.ID)
	if err != nil {
		slog.Error("caller org role", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	_, isMember, err := h.Store.MemberRole(r.Context(), project.ID, user.ID)
	if err != nil {
		slog.Error("member role", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var members []store.SubjectMember
	if CanAssign(user, orgRole, orgMember) {
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
		CSRFToken:   h.PageData(r, "").CSRFToken,
		CanCheck:    CanUpdate(user, isMember),
		CanAssign:   CanAssign(user, orgRole, orgMember),
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
