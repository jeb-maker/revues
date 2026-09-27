package runs

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jeb-maker/revues/internal/store"
)

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
