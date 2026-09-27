package runs

import (
	"log/slog"
	"net/http"
	"strings"

	viewtemplates "github.com/jeb-maker/revues/internal/web/templates"

	"github.com/jeb-maker/revues/internal/features/subjects"
	"github.com/jeb-maker/revues/internal/store"
)

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
	pendingRequired := PendingRequiredItems(runItems)
	progress := h.progressData(run.ID, runItems)

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
	canExportEvidence := run.Status == store.RunStatusDone && strings.TrimSpace(run.EvidenceCSVSHA256) != ""
	pd.HasEvidence = canExportEvidence
	data := viewtemplates.RunShowData{
		PageData:             pd,
		Subject:              project,
		Run:                  run,
		RunDisplayLabel:      displayLabel,
		Items:                items,
		ItemSections:         sectionGroups,
		NokItems:             nokItems,
		PendingRequiredItems: pendingRequired,
		CanSubmitComplete:    len(pendingRequired) == 0,
		Sections:             sections,
		FilterSection:        filterSection,
		FilterStatus:         filterStatus,
		JiraLinks:            jiraLinks,
		Attachments:          attachmentsByItem,
		Members:              members,
		TemplateName:         versionInfo.Name,
		VersionNum:           versionInfo.Version,
		MemberRole:           subjects.DisplayRole(access),
		CanLaunch:            CanLaunchAccess(user, access),
		CanCheck:             CanUpdateAccess(user, access),
		CanAssign:            showAssign && CanAssignAccess(user, access),
		CanLinkJira:          CanLinkJiraAccess(user, access),
		JiraConfigured:       pd.HasJira,
		CanComplete:          CanCompleteAccess(user, access),
		NotionConfigured:     h.notionConfigured(r.Context()),
		CanExportNotion:      CanCompleteAccess(user, access) && run.Status == store.RunStatusDone && strings.TrimSpace(run.NotionURL) == "",
		CanExportEvidence:    canExportEvidence,
		Progress:             progress,
		Message:              extra.Message,
		ItemError:            extra.ItemError,
		AssignError:          extra.AssignError,
		CompleteError:        extra.CompleteError,
		NotionExportError:    extra.NotionExportError,
		ClosingNote:          extra.ClosingNote,
		Error:                extra.Error,
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
			ID:         viewtemplates.SectionIDForTitle(title),
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
