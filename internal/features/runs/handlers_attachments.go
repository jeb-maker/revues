package runs

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/jeb-maker/revues/internal/attachments"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/middleware"
	viewtemplates "github.com/jeb-maker/revues/internal/web/templates"
)

const multipartMaxMemory = 6 << 20

func (h *Runs) UploadAttachment(w http.ResponseWriter, r *http.Request) {
	run, project, user, memberRole, ok := h.loadRun(w, r)
	if !ok {
		return
	}
	if !CanUpdate(user, memberRole) {
		http.NotFound(w, r)
		return
	}
	if run.Status != store.RunStatusInProgress {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	itemID, err := strconv.ParseInt(chi.URLParam(r, "itemId"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if _, err = h.Store.RunItemByID(r.Context(), run.ID, itemID); err != nil {
		if errors.Is(err, store.ErrRunItemNotFound) {
			http.NotFound(w, r)
			return
		}
		slog.Error("load run item for attachment upload", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err = r.ParseMultipartForm(multipartMaxMemory); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("attachment")
	if err != nil {
		h.renderRunItemShow(w, r, run, project, user, memberRole, itemID, viewtemplates.RunItemShowData{UploadError: "Fichier manquant."})
		return
	}
	defer func() { _ = file.Close() }()
	data, err := attachments.ReadAllLimited(file, attachments.MaxUploadBytes)
	if err != nil {
		h.renderRunItemShow(w, r, run, project, user, memberRole, itemID, viewtemplates.RunItemShowData{UploadError: uploadErrorMessage(err)})
		return
	}
	if _, err := h.attachmentService().Save(r.Context(), itemID, header.Filename, data); err != nil {
		h.renderRunItemShow(w, r, run, project, user, memberRole, itemID, viewtemplates.RunItemShowData{UploadError: uploadErrorMessage(err)})
		return
	}
	http.Redirect(w, r, "/runs/"+strconv.FormatInt(run.ID, 10)+"/items/"+strconv.FormatInt(itemID, 10)+"?msg=Pi%C3%A8ce+jointe+enregistr%C3%A9e", http.StatusSeeOther)
}

func (h *Runs) DownloadAttachment(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	attachmentID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	runID, err := h.Store.RunIDForAttachment(r.Context(), attachmentID)
	if errors.Is(err, store.ErrAttachmentNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		slog.Error("run id for attachment", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	run, err := h.Store.RunByID(r.Context(), runID)
	if errors.Is(err, store.ErrRunNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		slog.Error("load run for attachment download", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	_, isMember, err := h.Store.MemberRole(r.Context(), run.ProjectID, user.ID)
	if err != nil {
		slog.Error("member role for attachment download", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !CanView(user, isMember) {
		http.NotFound(w, r)
		return
	}
	att, path, err := h.attachmentService().Open(r.Context(), attachmentID)
	if err != nil {
		if errors.Is(err, store.ErrAttachmentNotFound) {
			http.NotFound(w, r)
			return
		}
		slog.Error("open attachment", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if strings.Contains(att.StoragePath, "..") {
		http.NotFound(w, r)
		return
	}
	baseDir := filepath.Clean(h.AttachmentsDir)
	clean := filepath.Clean(path)
	if clean != baseDir && !strings.HasPrefix(clean, baseDir+string(os.PathSeparator)) {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", att.MimeType)
	if r.URL.Query().Get("inline") == "1" && attachments.IsImageMime(att.MimeType) {
		w.Header().Set("Content-Disposition", "inline")
	} else {
		w.Header().Set("Content-Disposition", contentDispositionAttachment(att.Filename))
	}
	http.ServeFile(w, r, clean)
}

func (h *Runs) loadAttachmentsForItems(ctx context.Context, runItems []store.RunItem) map[int64]*store.Attachment {
	itemIDs := make([]int64, len(runItems))
	for i, item := range runItems {
		itemIDs[i] = item.ID
	}
	attachmentsByItem, err := h.Store.ListAttachmentsByRunItemIDs(ctx, itemIDs)
	if err != nil {
		slog.Error("list attachments for run items", "err", err)
		return map[int64]*store.Attachment{}
	}
	return attachmentsByItem
}

func (h *Runs) attachmentService() *attachments.Service {
	s, _ := h.Store.(*store.Store)
	return &attachments.Service{Store: s, Dir: h.AttachmentsDir}
}

func uploadErrorMessage(err error) string {
	switch {
	case errors.Is(err, attachments.ErrTooLarge):
		return "Fichier trop volumineux (max 5 Mo)."
	case errors.Is(err, attachments.ErrUnsupportedType):
		return "Type de fichier non autorisé (JPEG, PNG, WebP ou PDF)."
	case errors.Is(err, attachments.ErrEmptyFile):
		return "Fichier vide."
	default:
		return "Impossible d'enregistrer la pièce jointe."
	}
}

func contentDispositionAttachment(filename string) string {
	safe := strings.Map(func(r rune) rune {
		if r >= 0x20 && r <= 0x7e && r != '"' && r != '\\' {
			return r
		}
		return '_'
	}, filename)
	if safe == "" {
		safe = "attachment"
	}
	return `attachment; filename="` + safe + `"`
}
