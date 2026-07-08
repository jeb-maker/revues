package runs

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/jeb-maker/revues/internal/integrations/jira"
	"github.com/jeb-maker/revues/internal/store"
	viewtemplates "github.com/jeb-maker/revues/internal/web/templates"
)

// LinkJiraItem associates a Jira issue with a run item.
func (h *Runs) LinkJiraItem(w http.ResponseWriter, r *http.Request) {
	run, project, user, memberRole, ok := h.loadRun(w, r)
	if !ok {
		return
	}
	if !CanLinkJira(user, memberRole) {
		http.NotFound(w, r)
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

	if _, err := h.Store.RunItemByID(r.Context(), run.ID, itemID); err != nil {
		if errors.Is(err, store.ErrRunItemNotFound) {
			http.NotFound(w, r)
			return
		}
		slog.Error("load run item for jira link", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	input := strings.TrimSpace(r.FormValue("jira_issue"))
	linkSvc := h.jiraLinkService()
	if _, err := linkSvc.LinkRunItem(r.Context(), itemID, input); err != nil {
		h.renderRunItemShow(w, r, run, project, user, memberRole, itemID, viewtemplates.RunItemShowData{
			LinkError:      linkErrorMessage(err),
			JiraIssueInput: input,
		})
		return
	}

	http.Redirect(w, r, "/runs/"+strconv.FormatInt(run.ID, 10)+"/items/"+strconv.FormatInt(itemID, 10)+"?msg=Lien+Jira+enregistr%C3%A9", http.StatusSeeOther)
}

func (h *Runs) jiraLinkService() *jira.LinkService {
	return &jira.LinkService{
		Store:         h.Store.(*store.Store),
		EncryptionKey: h.EncryptionKey,
	}
}

func (h *Runs) jiraConfigured(ctx context.Context) bool {
	ok, err := h.jiraLinkService().Configured(ctx)
	return err == nil && ok
}

func linkErrorMessage(err error) string {
	switch {
	case errors.Is(err, jira.ErrNotConfigured):
		return "Jira n'est pas configuré. Contactez un administrateur."
	case errors.Is(err, jira.ErrInvalidIssueReference):
		return "Clé ou URL Jira invalide (ex. PROJ-123)."
	case errors.Is(err, jira.ErrIssueNotFound):
		return "Issue Jira introuvable."
	case errors.Is(err, jira.ErrConnectionFailed):
		return "Impossible de contacter Jira. Réessayez plus tard."
	default:
		if msg := err.Error(); msg != "" {
			return msg
		}
		return "Impossible de lier l'issue Jira."
	}
}

func (h *Runs) loadJiraLinksForItems(ctx context.Context, runItems []store.RunItem) map[int64]store.IntegrationLink {
	itemIDs := make([]int64, len(runItems))
	for i, item := range runItems {
		itemIDs[i] = item.ID
	}
	links, err := h.Store.ListIntegrationLinksByRunItemIDs(ctx, itemIDs, store.IntegrationTypeJira)
	if err != nil {
		slog.Error("list jira links for run items", "err", err)
		return map[int64]store.IntegrationLink{}
	}
	return links
}

func (h *Runs) renderRunItemShow(w http.ResponseWriter, r *http.Request, run *store.ChecklistRun, project *store.Project, user *store.User, memberRole string, itemID int64, extra viewtemplates.RunItemShowData) {
	item, err := h.Store.RunItemByID(r.Context(), run.ID, itemID)
	if errors.Is(err, store.ErrRunItemNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		slog.Error("load run item", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	events, err := h.Store.ListRunItemEvents(r.Context(), item.ID)
	if err != nil {
		slog.Error("list run item events", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	jiraLink, _ := h.Store.IntegrationLinkByRunItemAndType(r.Context(), item.ID, store.IntegrationTypeJira)
	var attachment *store.Attachment
	if att, attErr := h.Store.AttachmentByRunItemID(r.Context(), item.ID); attErr == nil {
		attachment = att
	} else if !errors.Is(attErr, store.ErrAttachmentNotFound) {
		slog.Error("load attachment for run item show", "err", attErr)
	}

	data := viewtemplates.RunItemShowData{
		PageData:        h.PageData(r, item.Label),
		Project:         project,
		Run:             run,
		Item:            item,
		Events:          events,
		JiraLink:        jiraLink,
		Attachment:      attachment,
		MemberRole:      memberRole,
		CanCheck:        CanUpdate(user, memberRole),
		CanUpload:       CanUpdate(user, memberRole) && run.Status == store.RunStatusInProgress,
		CanLinkJira:     CanLinkJira(user, memberRole),
		JiraConfigured:  h.jiraConfigured(r.Context()),
		Message:         extra.Message,
		LinkError:       extra.LinkError,
		JiraIssueInput:  extra.JiraIssueInput,
		CreateError:     extra.CreateError,
		ShowJiraCreate:  extra.ShowJiraCreate,
		JiraCreateTitle: extra.JiraCreateTitle,
		JiraCreateDesc:  extra.JiraCreateDesc,
		UploadError:     extra.UploadError,
	}
	if data.Message == "" {
		data.Message = r.URL.Query().Get("msg")
	}
	if item.Status == store.RunItemStatusNOK && data.JiraLink == nil && data.CanLinkJira && data.JiraConfigured && !data.ShowJiraCreate {
		data.ShowJiraCreate = true
		itemURL := h.jiraRunItemContext(r, run, project, itemID).ItemURL
		data.JiraCreateTitle, data.JiraCreateDesc = h.jiraCreateDefaults(r.Context(), run, project, item, itemURL)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	statusCode := http.StatusOK
	if extra.LinkError != "" || extra.CreateError != "" || extra.UploadError != "" {
		statusCode = http.StatusBadRequest
	}
	w.WriteHeader(statusCode)
	if err := h.Templates.ExecuteTemplate(w, "run_item_show", data); err != nil {
		slog.Error("render run item show", "err", err)
	}
}
