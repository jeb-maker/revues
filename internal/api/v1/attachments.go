package apiv1

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/jeb-maker/revues/internal/attachments"
	"github.com/jeb-maker/revues/internal/features/runs"
	"github.com/jeb-maker/revues/internal/store"
)

// GetRunItemAttachment serves GET /api/v1/runs/{runId}/items/{itemId}/attachments.
func (s *Server) GetRunItemAttachment(w http.ResponseWriter, r *http.Request, runID RunId, itemID RunItemId) {
	run, _, _, ok := s.ensureRunAccess(w, r, runID)
	if !ok {
		return
	}
	if _, err := s.Store.RunItemByID(r.Context(), run.ID, itemID); err != nil {
		if errors.Is(err, store.ErrRunItemNotFound) {
			writeAPIError(w, http.StatusNotFound, "not_found", "Point introuvable.")
			return
		}
		slog.Error("load run item for attachment meta", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	att, err := s.Store.AttachmentByRunItemID(r.Context(), itemID)
	if err != nil {
		if errors.Is(err, store.ErrAttachmentNotFound) {
			writeAPIError(w, http.StatusNotFound, "not_found", "Pièce jointe introuvable.")
			return
		}
		slog.Error("load attachment meta", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	writeJSON(w, http.StatusOK, mapAttachment(att))
}

// UploadRunItemAttachment serves POST /api/v1/runs/{runId}/items/{itemId}/attachments.
func (s *Server) UploadRunItemAttachment(w http.ResponseWriter, r *http.Request, runID RunId, itemID RunItemId) {
	run, user, access, ok := s.ensureRunAccess(w, r, runID)
	if !ok {
		return
	}
	if !runs.CanUpdateAccess(user, access) {
		writeAPIError(w, http.StatusNotFound, "not_found", "Revue introuvable.")
		return
	}
	if run.Status != store.RunStatusInProgress {
		writeAPIError(w, http.StatusConflict, "conflict", "Cette revue n'est plus éditable.")
		return
	}
	if _, err := s.Store.RunItemByID(r.Context(), run.ID, itemID); err != nil {
		if errors.Is(err, store.ErrRunItemNotFound) {
			writeAPIError(w, http.StatusNotFound, "not_found", "Point introuvable.")
			return
		}
		slog.Error("load run item for attachment upload", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if s.Attachments == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	if r.MultipartForm == nil {
		r.Body = http.MaxBytesReader(w, r.Body, attachments.MaxMultipartBodyBytes)
		if err := r.ParseMultipartForm(attachments.MaxMultipartBodyBytes); err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				writeAPIError(w, http.StatusRequestEntityTooLarge, "payload_too_large", "Fichier trop volumineux (max 5 Mo).")
				return
			}
			writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête multipart invalide.")
			return
		}
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		// Accept legacy field name used by the previous HTML form.
		file, header, err = r.FormFile("attachment")
	}
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Fichier manquant.")
		return
	}
	defer func() { _ = file.Close() }()

	data, err := attachments.ReadAllLimited(file, attachments.MaxUploadBytes)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", uploadErrorMessage(err))
		return
	}
	att, err := s.Attachments.Save(r.Context(), itemID, header.Filename, data)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", uploadErrorMessage(err))
		return
	}
	writeJSON(w, http.StatusCreated, mapAttachment(att))
}

// DownloadRunItemAttachment serves GET /api/v1/runs/{runId}/items/{itemId}/attachments/{attachmentId}.
func (s *Server) DownloadRunItemAttachment(w http.ResponseWriter, r *http.Request, runID RunId, itemID RunItemId, attachmentID AttachmentId) {
	run, _, _, ok := s.ensureRunAccess(w, r, runID)
	if !ok {
		return
	}
	if _, err := s.Store.RunItemByID(r.Context(), run.ID, itemID); err != nil {
		if errors.Is(err, store.ErrRunItemNotFound) {
			writeAPIError(w, http.StatusNotFound, "not_found", "Point introuvable.")
			return
		}
		slog.Error("load run item for attachment download", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if s.Attachments == nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	att, path, err := s.Attachments.Open(r.Context(), attachmentID)
	if err != nil {
		if errors.Is(err, store.ErrAttachmentNotFound) {
			writeAPIError(w, http.StatusNotFound, "not_found", "Pièce jointe introuvable.")
			return
		}
		slog.Error("open attachment", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if att.RunItemID != itemID {
		writeAPIError(w, http.StatusNotFound, "not_found", "Pièce jointe introuvable.")
		return
	}
	if strings.Contains(att.StoragePath, "..") {
		writeAPIError(w, http.StatusNotFound, "not_found", "Pièce jointe introuvable.")
		return
	}
	baseDir := filepath.Clean(s.Config.AttachmentsDir)
	clean := filepath.Clean(path)
	if clean != baseDir && !strings.HasPrefix(clean, baseDir+string(os.PathSeparator)) {
		writeAPIError(w, http.StatusNotFound, "not_found", "Pièce jointe introuvable.")
		return
	}

	w.Header().Set("Content-Type", att.MimeType)
	w.Header().Set("Content-Disposition", contentDispositionAttachment(att.Filename))
	http.ServeFile(w, r, clean)
}

func mapAttachment(att *store.Attachment) Attachment {
	return Attachment{
		Id:        att.ID,
		RunItemId: att.RunItemID,
		Filename:  att.Filename,
		MimeType:  att.MimeType,
		SizeBytes: att.SizeBytes,
		CreatedAt: att.CreatedAt,
		IsImage:   attachments.IsImageMime(att.MimeType),
	}
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
		if r >= 0x20 && r <= 0x7e && r != '"' && r != '\\' && unicode.IsPrint(r) {
			return r
		}
		return '_'
	}, filename)
	if safe == "" {
		safe = "attachment"
	}
	return `attachment; filename="` + safe + `"`
}
