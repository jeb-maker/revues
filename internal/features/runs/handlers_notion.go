package runs

import (
	"context"
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
			NotionExportError: notion.UserMessage(err),
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
