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

// CreateJiraItem creates a Jira issue from a nok run item.
func (h *Runs) CreateJiraItem(w http.ResponseWriter, r *http.Request) {
	run, project, user, access, ok := h.loadRun(w, r)
	if !ok {
		return
	}
	if !CanLinkJiraAccess(user, access) {
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

	item, err := h.Store.RunItemByID(r.Context(), run.ID, itemID)
	if err != nil {
		if errors.Is(err, store.ErrRunItemNotFound) {
			http.NotFound(w, r)
			return
		}
		slog.Error("load run item for jira create", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	input := jira.CreateInput{
		Title:       strings.TrimSpace(r.FormValue("jira_title")),
		Description: strings.TrimSpace(r.FormValue("jira_description")),
	}
	itemCtx := h.jiraRunItemContext(r, run, project, itemID)

	createSvc := h.jiraCreateService()
	if _, err := createSvc.CreateRunItem(r.Context(), run.ID, itemID, input, itemCtx); err != nil {
		defaultTitle, defaultDesc := jira.DefaultIssueContent(item, itemCtx)
		title := input.Title
		description := input.Description
		if title == "" {
			title = defaultTitle
		}
		if description == "" {
			description = defaultDesc
		}
		h.renderRunItemShow(w, r, run, project, user, access, itemID, viewtemplates.RunItemShowData{
			CreateError:     createErrorMessage(err),
			JiraCreateTitle: title,
			JiraCreateDesc:  description,
			ShowJiraCreate:  true,
		})
		return
	}

	http.Redirect(w, r, "/runs/"+strconv.FormatInt(run.ID, 10)+"/items/"+strconv.FormatInt(itemID, 10)+"?msg=Ticket+Jira+cr%C3%A9%C3%A9", http.StatusSeeOther)
}

func (h *Runs) jiraCreateService() *jira.CreateService {
	s, _ := h.Store.(*store.Store)
	return &jira.CreateService{
		Store:         s,
		EncryptionKey: h.EncryptionKey,
	}
}

func (h *Runs) jiraRunItemContext(r *http.Request, run *store.ChecklistRun, subject *store.Subject, itemID int64) jira.RunItemContext {
	baseURL := strings.TrimRight(h.baseURL(r), "/")
	return jira.RunItemContext{
		SubjectName: subject.Name,
		RunTitle:    h.runDisplayLabel(r.Context(), run, subject),
		ItemURL:     baseURL + "/runs/" + strconv.FormatInt(run.ID, 10) + "/items/" + strconv.FormatInt(itemID, 10),
	}
}

func (h *Runs) baseURL(r *http.Request) string {
	if r != nil {
		scheme := "http"
		if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
			scheme = "https"
		}
		if host := r.Host; host != "" {
			return scheme + "://" + host
		}
	}
	return "http://localhost:8080"
}

func createErrorMessage(err error) string {
	switch {
	case errors.Is(err, jira.ErrNotConfigured):
		return "Jira n'est pas configuré. Contactez un administrateur."
	case errors.Is(err, jira.ErrProjectKeyMissing):
		return "Clé projet Jira manquante dans la configuration admin."
	case errors.Is(err, jira.ErrNotNOK):
		return "Seuls les points non validés peuvent générer un ticket Jira."
	case errors.Is(err, jira.ErrAlreadyLinked):
		return "Une issue Jira est déjà liée à ce point."
	case errors.Is(err, jira.ErrConnectionFailed):
		return "Impossible de contacter Jira. Réessayez plus tard."
	case errors.Is(err, jira.ErrCreateFailed):
		return "Jira a refusé la création du ticket. Vérifiez la configuration (projet, type d'issue)."
	default:
		if msg := err.Error(); msg != "" {
			return msg
		}
		return "Impossible de créer le ticket Jira."
	}
}

func (h *Runs) jiraCreateDefaults(ctx context.Context, run *store.ChecklistRun, subject *store.Subject, item *store.RunItem, itemURL string) (title, description string) {
	runTitle := RunDisplayLabel("", subject.Name, run.CreatedAt, run.ID)
	if versionInfo, err := h.Store.TemplateVersionInfo(ctx, run.TemplateVersionID); err == nil {
		runTitle = RunDisplayLabel(versionInfo.Name, subject.Name, run.CreatedAt, run.ID)
	}
	return jira.DefaultIssueContent(item, jira.RunItemContext{
		SubjectName: subject.Name,
		RunTitle:    runTitle,
		ItemURL:     itemURL,
	})
}
