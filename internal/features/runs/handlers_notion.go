package runs

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/jeb-maker/revues/internal/integrations/notion"
	"github.com/jeb-maker/revues/internal/store"
	viewtemplates "github.com/jeb-maker/revues/internal/web/templates"
)

func (h *Runs) ExportNotion(w http.ResponseWriter, r *http.Request) {
	run, project, user, access, ok := h.loadRun(w, r)
	if !ok {
		return
	}
	if !CanCompleteAccess(user, access) {
		http.NotFound(w, r)
		return
	}
	if run.Status != store.RunStatusDone {
		http.NotFound(w, r)
		return
	}
	if _, err := h.notionExportService().ExportRun(r.Context(), run.ID); err != nil {
		h.renderRunShow(w, r, run, project, user, access, viewtemplates.RunShowData{
			NotionExportError: notionExportErrorMessage(err),
		})
		return
	}
	http.Redirect(w, r, "/runs/"+strconv.FormatInt(run.ID, 10)+"?msg=Revue+export%C3%A9e+vers+Notion", http.StatusSeeOther)
}

func (h *Runs) notionExportService() *notion.ExportService {
	s, _ := h.Store.(*store.Store)
	svc := &notion.ExportService{Store: s, EncryptionKey: h.EncryptionKey, BaseURL: h.BaseURL}
	if h.NotionClient != nil {
		svc.Client = h.NotionClient
	}
	return svc
}

func (h *Runs) notionConfigured(ctx context.Context) bool {
	s2, _ := h.Store.(*store.Store)
	cfg, ok, err := (&notion.Service{Store: s2, EncryptionKey: h.EncryptionKey}).Load(ctx)
	return err == nil && ok && notion.ExportReady(cfg)
}

func notionExportErrorMessage(err error) string {
	switch {
	case errors.Is(err, notion.ErrNotConfigured):
		return "Notion n'est pas configuré. Contactez un administrateur."
	case errors.Is(err, notion.ErrDatabaseMissing):
		return "Aucune base Notion par défaut configurée. Contactez un administrateur."
	case errors.Is(err, notion.ErrAlreadyExported):
		return "Cette revue a déjà été exportée vers Notion."
	case errors.Is(err, notion.ErrRunNotDone):
		return "Seules les revues terminées peuvent être exportées."
	case errors.Is(err, notion.ErrExportFailed), errors.Is(err, notion.ErrConnectionFailed):
		return "Impossible d'exporter vers Notion. Réessayez plus tard."
	default:
		if msg := err.Error(); msg != "" {
			return msg
		}
		return "Impossible d'exporter vers Notion."
	}
}
