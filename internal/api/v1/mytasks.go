package apiv1

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/jeb-maker/revues/internal/features/mytasks"
	"github.com/jeb-maker/revues/internal/store"
)

// ListMyTasks serves GET /api/v1/me/tasks.
//
// IDOR: store.ListAssignedRunItems filters assigned_to = current user.
// Org scope: store also filters by active organization (orgctx) and excludes
// archived subjects/runs — same SoftLoad org middleware as other API routes.
func (s *Server) ListMyTasks(w http.ResponseWriter, r *http.Request, params ListMyTasksParams) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}
	if !requireOrg(w, r) {
		return
	}

	status := ""
	if params.Status != nil {
		status = string(*params.Status)
	}
	q := ""
	if params.Q != nil {
		q = strings.TrimSpace(*params.Q)
	}

	var tasksStore mytasks.AssignedTaskStore = s.Store
	rows, err := tasksStore.ListAssignedRunItems(r.Context(), user.ID, status, q)
	if err != nil {
		slog.Error("list my tasks", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	out := make([]MyTask, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapMyTask(row))
	}
	writeJSON(w, http.StatusOK, MyTaskListResponse{Tasks: out})
}

func mapMyTask(row store.AssignedRunItemSummary) MyTask {
	out := MyTask{
		Id:          row.ID,
		RunId:       row.RunID,
		Section:     row.Section,
		Position:    row.Position,
		Label:       row.Label,
		HelpText:    ptrString(row.HelpText),
		Required:    row.Required,
		Status:      MyTaskStatus(row.Status),
		Comment:     row.Comment,
		UpdatedAt:   row.UpdatedAt,
		RunTitle:    row.RunTitle,
		SubjectId:   row.SubjectID,
		SubjectName: row.SubjectName,
	}
	if row.AssignedTo.Valid {
		out.AssignedTo = &row.AssignedTo.Int64
	}
	if row.AssignedLogin != "" {
		out.AssignedLogin = &row.AssignedLogin
	}
	return out
}
