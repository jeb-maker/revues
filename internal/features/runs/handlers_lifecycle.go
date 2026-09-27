package runs

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	viewtemplates "github.com/jeb-maker/revues/internal/web/templates"

	"github.com/jeb-maker/revues/internal/store"
)

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
	run, project, user, access, ok := h.loadRun(w, r)
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

	runItems, err := h.Store.ListRunItems(r.Context(), run.ID)
	if err != nil {
		slog.Error("list run items for complete", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err = ValidateComplete(runItems); err != nil {
		completeErr := "Traitez tous les points obligatoires avant de clôturer."
		extra := viewtemplates.RunShowData{
			CompleteError: completeErr,
			ClosingNote:   closingNote,
		}
		if h.isHTMX(r) {
			h.renderCompleteSectionHTMX(w, r, run, runItems, completeErr, closingNote)
			return
		}
		h.renderRunShow(w, r, run, project, user, access, extra)
		return
	}

	if err = h.Store.CompleteRun(r.Context(), run.ID, closingNote); err != nil {
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

func (h *Runs) completeStatusData(r *http.Request, run *store.ChecklistRun, runItems []store.RunItem, completeErr, closingNote string) viewtemplates.RunCompleteStatusData {
	pendingRequired := PendingRequiredItems(runItems)
	pd := h.PageData(r, "")
	return viewtemplates.RunCompleteStatusData{
		Run:                  run,
		NokItems:             nokItemsFromRunItems(runItems),
		PendingRequiredItems: pendingRequired,
		Progress:             h.progressData(run.ID, runItems),
		CanSubmitComplete:    len(pendingRequired) == 0,
		CompleteError:        completeErr,
		ClosingNote:          closingNote,
		CSRFToken:            pd.CSRFToken,
		Labels:               pd.Labels,
	}
}

func (h *Runs) renderCompleteSectionHTMX(w http.ResponseWriter, r *http.Request, run *store.ChecklistRun, runItems []store.RunItem, completeErr, closingNote string) {
	data := h.completeStatusData(r, run, runItems, completeErr, closingNote)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	if err := h.Templates.ExecuteTemplate(w, "run_complete_section_fragment", data); err != nil {
		slog.Error("render run complete section fragment", "err", err)
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
