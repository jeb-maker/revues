package apiv1

import (
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/features/runs"
	"github.com/jeb-maker/revues/internal/store"
)

// runItemConflictMessage is returned with 409 `conflict` when the client's updated_at is stale.
const runItemConflictMessage = "Ce point a été modifié entre-temps. Rechargez la page."

// ListRuns serves GET /api/v1/runs.
func (s *Server) ListRuns(w http.ResponseWriter, r *http.Request, params ListRunsParams) {
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
	limit := store.FilteredRunsPageSize
	if params.Limit != nil && *params.Limit > 0 {
		limit = *params.Limit
	}
	offset := 0
	if params.Offset != nil && *params.Offset > 0 {
		offset = *params.Offset
	}

	admin := auth.HasMinRole(user.Role, auth.RoleAdmin)
	summaries, total, err := s.Store.ListFilteredRunSummaries(r.Context(), user.ID, admin, status, q, limit, offset)
	if err != nil {
		slog.Error("list runs", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	out := make([]RunSummary, 0, len(summaries))
	for _, row := range summaries {
		out = append(out, mapRunListSummary(row))
	}
	writeJSON(w, http.StatusOK, RunListResponse{Runs: out, Total: total})
}

// ListSubjectRuns serves GET /api/v1/subjects/{subjectID}/runs.
func (s *Server) ListSubjectRuns(w http.ResponseWriter, r *http.Request, subjectID SubjectId) {
	subject, _, _, ok := s.ensureSubjectAccess(w, r, subjectID)
	if !ok {
		return
	}
	rows, err := s.Store.ListRunsWithProgressBySubject(r.Context(), subject.ID)
	if err != nil {
		slog.Error("list subject runs", "err", err, "subject_id", subjectID)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	out := make([]RunSummary, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapRunWithProgress(row, subject.Name))
	}
	writeJSON(w, http.StatusOK, RunListResponse{Runs: out, Total: len(out)})
}

// ListSubjectRunTemplates serves GET /api/v1/subjects/{subjectID}/run-templates.
func (s *Server) ListSubjectRunTemplates(w http.ResponseWriter, r *http.Request, subjectID SubjectId) {
	_, user, access, ok := s.ensureSubjectAccess(w, r, subjectID)
	if !ok {
		return
	}
	templates, err := s.Store.ListChecklistTemplates(r.Context(), subjectID)
	if err != nil {
		slog.Error("list run templates", "err", err, "subject_id", subjectID)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	out := make([]RunTemplateSummary, 0, len(templates))
	for _, t := range templates {
		domains := t.Tags
		if domains == nil {
			domains = []string{}
		}
		out = append(out, RunTemplateSummary{
			Id:            t.ID,
			Name:          t.Name,
			LatestVersion: t.LatestVersion,
			ItemCount:     t.ItemCount,
			Domains:       &domains,
		})
	}
	writeJSON(w, http.StatusOK, RunTemplateListResponse{
		Templates: out,
		CanLaunch: runs.CanLaunchAccess(user, access),
	})
}

// CreateRun serves POST /api/v1/subjects/{subjectID}/runs.
func (s *Server) CreateRun(w http.ResponseWriter, r *http.Request, subjectID SubjectId) {
	subject, user, access, ok := s.ensureSubjectAccess(w, r, subjectID)
	if !ok {
		return
	}
	if !runs.CanLaunchAccess(user, access) {
		writeAPIError(w, http.StatusNotFound, "not_found", "Sujet introuvable.")
		return
	}

	var req CreateRunRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
		return
	}
	if req.TemplateId <= 0 {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Modèle requis.")
		return
	}

	dueRaw := ""
	if req.DueDate != nil {
		dueRaw = strings.TrimSpace(*req.DueDate)
	}
	dueISO, err := runs.ParseDueDate(dueRaw)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Date d'échéance invalide (attendu YYYY-MM-DD).")
		return
	}

	run, err := s.Store.CreateChecklistRunWithDueDate(
		r.Context(), subject.ID, req.TemplateId, user.ID,
		sql.NullString{String: dueISO, Valid: dueISO != ""},
	)
	if err != nil {
		if errors.Is(err, store.ErrChecklistTemplateNotFound) {
			writeAPIError(w, http.StatusBadRequest, "validation_failed", "Modèle introuvable ou incompatible avec ce sujet.")
			return
		}
		slog.Error("create checklist run", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	detail, ok := s.buildRunDetail(w, r, run, user)
	if !ok {
		return
	}
	writeJSON(w, http.StatusCreated, detail)
}

// GetRun serves GET /api/v1/runs/{runID}.
func (s *Server) GetRun(w http.ResponseWriter, r *http.Request, runID RunId) {
	run, user, _, ok := s.ensureRunAccess(w, r, runID)
	if !ok {
		return
	}
	detail, ok := s.buildRunDetail(w, r, run, user)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

// CompleteRun serves POST /api/v1/runs/{runID}/complete.
func (s *Server) CompleteRun(w http.ResponseWriter, r *http.Request, runID RunId) {
	run, user, access, ok := s.ensureRunAccess(w, r, runID)
	if !ok {
		return
	}
	if !runs.CanCompleteAccess(user, access) {
		writeAPIError(w, http.StatusNotFound, "not_found", "Revue introuvable.")
		return
	}
	if run.Status != store.RunStatusInProgress {
		writeAPIError(w, http.StatusConflict, "conflict", "Cette revue n'est plus éditable.")
		return
	}

	var req CompleteRunRequest
	if r.Body != nil && r.ContentLength != 0 {
		if err := decodeJSONBody(r, &req); err != nil && !errors.Is(err, io.EOF) {
			writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
			return
		}
	}
	closingNote := ""
	if req.ClosingNote != nil {
		closingNote = strings.TrimSpace(*req.ClosingNote)
	}

	items, err := s.Store.ListRunItems(r.Context(), run.ID)
	if err != nil {
		slog.Error("list run items for complete", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	if err = runs.ValidateComplete(items); err != nil {
		if errors.Is(err, runs.ErrPendingRequired) {
			writeAPIError(w, http.StatusBadRequest, "validation_failed",
				"Des points obligatoires sont encore en attente.")
			return
		}
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Clôture impossible.")
		return
	}

	csvRows, err := s.Store.ListRunExportRows(r.Context(), run.ID)
	if err != nil {
		slog.Error("export rows before complete", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	csvData, err := runs.BuildRunCSV(csvRows)
	if err != nil {
		slog.Error("build csv before complete", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	hash := runs.SHA256Hex(csvData)

	if err = s.Store.CompleteRunWithEvidence(r.Context(), run.ID, closingNote, hash); err != nil {
		if errors.Is(err, store.ErrInvalidRunStatus) {
			writeAPIError(w, http.StatusConflict, "conflict", "Cette revue n'est plus éditable.")
			return
		}
		slog.Error("complete run", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	if s.Webhooks != nil {
		s.Webhooks.EmitReviewCompleted(r.Context(), run.ID)
	}
	if s.Notifier != nil {
		s.Notifier.NotifyRunCompleted(r.Context(), run.ID)
	}

	run, err = s.Store.RunByID(r.Context(), run.ID)
	if err != nil {
		slog.Error("reload completed run", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}
	detail, ok := s.buildRunDetail(w, r, run, user)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

// GetRunItem serves GET /api/v1/runs/{runID}/items/{itemID}.
func (s *Server) GetRunItem(w http.ResponseWriter, r *http.Request, runID RunId, itemID RunItemId) {
	run, user, access, ok := s.ensureRunAccess(w, r, runID)
	if !ok {
		return
	}
	detail, ok := s.buildRunItemDetail(w, r, run, itemID, user, access)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

// UpdateRunItem serves PATCH /api/v1/runs/{runID}/items/{itemID}.
func (s *Server) UpdateRunItem(w http.ResponseWriter, r *http.Request, runID RunId, itemID RunItemId) {
	run, user, access, ok := s.ensureRunAccess(w, r, runID)
	if !ok {
		return
	}
	if run.Status != store.RunStatusInProgress {
		writeAPIError(w, http.StatusConflict, "conflict", "Cette revue n'est plus éditable.")
		return
	}

	var req UpdateRunItemRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
		return
	}

	item, err := s.Store.RunItemByID(r.Context(), run.ID, itemID)
	if err != nil {
		if errors.Is(err, store.ErrRunItemNotFound) {
			writeAPIError(w, http.StatusNotFound, "not_found", "Point introuvable.")
			return
		}
		slog.Error("load run item", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	// Optimistic lock: the client echoes the updated_at it displayed; empty = unconditional write.
	expectedUpdatedAt := ""
	if req.UpdatedAt != nil {
		expectedUpdatedAt = strings.TrimSpace(*req.UpdatedAt)
	}

	statusChanged := false
	newStatus := item.Status
	newComment := item.Comment

	if req.Status != nil || req.Comment != nil {
		if !runs.CanUpdateAccess(user, access) {
			writeAPIError(w, http.StatusNotFound, "not_found", "Revue introuvable.")
			return
		}
		if req.Status != nil {
			newStatus = string(*req.Status)
		}
		if req.Comment != nil {
			newComment = *req.Comment
		}
		if err := runs.ValidateUpdate(newStatus, newComment); err != nil {
			if errors.Is(err, runs.ErrCommentRequired) {
				writeAPIError(w, http.StatusBadRequest, "validation_failed",
					"Un commentaire est obligatoire pour le statut Non validé.")
				return
			}
			if errors.Is(err, runs.ErrInvalidStatus) {
				writeAPIError(w, http.StatusBadRequest, "validation_failed", "Statut invalide.")
				return
			}
			writeAPIError(w, http.StatusBadRequest, "validation_failed", "Mise à jour invalide.")
			return
		}
		err := s.Store.UpdateRunItemStatusChecked(
			r.Context(), run.ID, itemID, user.ID, newStatus, newComment, expectedUpdatedAt,
		)
		if err != nil {
			if errors.Is(err, store.ErrRunItemConflict) {
				writeAPIError(w, http.StatusConflict, "conflict", runItemConflictMessage)
				return
			}
			if errors.Is(err, store.ErrRunNotEditable) {
				writeAPIError(w, http.StatusConflict, "conflict", "Cette revue n'est plus éditable.")
				return
			}
			if errors.Is(err, store.ErrRunItemNotFound) {
				writeAPIError(w, http.StatusNotFound, "not_found", "Point introuvable.")
				return
			}
			slog.Error("update run item status", "err", err)
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
			return
		}
		statusChanged = item.Status != newStatus

		if expectedUpdatedAt != "" {
			// The status write bumped updated_at: chain the lock to the fresh value for the assign step.
			fresh, err := s.Store.RunItemByID(r.Context(), run.ID, itemID)
			if err != nil {
				slog.Error("reload run item after status update", "err", err)
				writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
				return
			}
			expectedUpdatedAt = fresh.UpdatedAt
		}
	}

	newlyAssigned := false
	unassign := req.Unassign != nil && *req.Unassign
	if unassign || req.AssignedTo != nil {
		if !runs.CanAssignAccess(user, access) {
			writeAPIError(w, http.StatusForbidden, "forbidden", "Droits insuffisants pour assigner.")
			return
		}
		var assignee *int64
		if !unassign {
			assignee = req.AssignedTo
		}
		if err := s.Store.AssignRunItemChecked(r.Context(), run.ID, itemID, assignee, expectedUpdatedAt); err != nil {
			if errors.Is(err, store.ErrRunItemConflict) {
				writeAPIError(w, http.StatusConflict, "conflict", runItemConflictMessage)
				return
			}
			if errors.Is(err, store.ErrInvalidAssignee) {
				writeAPIError(w, http.StatusBadRequest, "validation_failed", "Assigné invalide (doit être membre du sujet).")
				return
			}
			if errors.Is(err, store.ErrRunNotEditable) {
				writeAPIError(w, http.StatusConflict, "conflict", "Cette revue n'est plus éditable.")
				return
			}
			if errors.Is(err, store.ErrRunItemNotFound) {
				writeAPIError(w, http.StatusNotFound, "not_found", "Point introuvable.")
				return
			}
			slog.Error("assign run item", "err", err)
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
			return
		}
		// Email only on an effective new assignee: the SPA re-sends assigned_to on every save,
		// so re-saving the same assignee (or unassigning) must not re-notify.
		newlyAssigned = assignee != nil && (!item.AssignedTo.Valid || item.AssignedTo.Int64 != *assignee)
	}

	if statusChanged && newStatus == runs.StatusNOK && s.Webhooks != nil {
		s.Webhooks.EmitReviewItemNOK(r.Context(), run.ID, itemID)
	}
	if newlyAssigned && s.Notifier != nil {
		s.Notifier.NotifyItemAssigned(r.Context(), run.ID, itemID)
	}

	detail, ok := s.buildRunItemDetail(w, r, run, itemID, user, access)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (s *Server) ensureRunAccess(w http.ResponseWriter, r *http.Request, runID int64) (*store.ChecklistRun, *store.User, store.SubjectAccess, bool) {
	user, ok := requireUser(w, r)
	if !ok {
		return nil, nil, store.SubjectAccess{}, false
	}
	if !requireOrg(w, r) {
		return nil, nil, store.SubjectAccess{}, false
	}

	run, err := s.Store.RunByID(r.Context(), runID)
	if err != nil {
		if errors.Is(err, store.ErrRunNotFound) {
			writeAPIError(w, http.StatusNotFound, "not_found", "Revue introuvable.")
			return nil, nil, store.SubjectAccess{}, false
		}
		slog.Error("load run", "err", err, "run_id", runID)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return nil, nil, store.SubjectAccess{}, false
	}

	access, err := s.Store.ResolveSubjectAccess(r.Context(), user.ID, run.SubjectID, user.Role)
	if err != nil {
		slog.Error("resolve subject access for run", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return nil, nil, store.SubjectAccess{}, false
	}
	if !runs.CanViewAccess(access) {
		writeAPIError(w, http.StatusNotFound, "not_found", "Revue introuvable.")
		return nil, nil, store.SubjectAccess{}, false
	}
	return run, user, access, true
}

func (s *Server) buildRunDetail(w http.ResponseWriter, r *http.Request, run *store.ChecklistRun, user *store.User) (RunDetail, bool) {
	access, err := s.Store.ResolveSubjectAccess(r.Context(), user.ID, run.SubjectID, user.Role)
	if err != nil {
		slog.Error("resolve access for run detail", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return RunDetail{}, false
	}

	subject, err := s.Store.SubjectByID(r.Context(), run.SubjectID)
	if err != nil {
		slog.Error("load subject for run", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return RunDetail{}, false
	}

	versionInfo, err := s.Store.TemplateVersionInfo(r.Context(), run.TemplateVersionID)
	if err != nil {
		slog.Error("template version info", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return RunDetail{}, false
	}

	items, err := s.Store.ListRunItems(r.Context(), run.ID)
	if err != nil {
		slog.Error("list run items", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return RunDetail{}, false
	}

	members, err := s.Store.ListSubjectMembers(r.Context(), run.SubjectID)
	if err != nil {
		slog.Error("list subject members for run", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return RunDetail{}, false
	}

	title, err := s.Store.RunDisplayLabelForRun(r.Context(), run)
	if err != nil {
		slog.Error("run display label", "err", err)
		title = versionInfo.Name
	}

	done, total := runs.Progress(items)
	pending := runs.PendingRequiredItems(items)
	pendingCount := len(pending)
	editable := run.Status == store.RunStatusInProgress

	assignees := mapAssignees(members)
	mappedItems := make([]RunItem, 0, len(items))
	for _, item := range items {
		mappedItems = append(mappedItems, mapRunItem(item))
	}

	detail := RunDetail{
		Id:              run.ID,
		Title:           title,
		SubjectId:       subject.ID,
		SubjectName:     subject.Name,
		Status:          RunDetailStatus(run.Status),
		ClosingNote:     ptrString(run.ClosingNote),
		TemplateId:      versionInfo.TemplateID,
		TemplateName:    versionInfo.Name,
		TemplateVersion: versionInfo.Version,
		CreatedAt:       run.CreatedAt,
		Progress:        RunProgress{Done: done, Total: total, Percent: progressPercent(done, total)},
		Items:           mappedItems,
		Assignees:       &assignees,
		Capabilities: RunCapabilities{
			CanUpdateItems: editable && runs.CanUpdateAccess(user, access),
			CanAssign:      editable && runs.CanAssignAccess(user, access),
			CanComplete:    editable && runs.CanCompleteAccess(user, access),
			CanExportNotion: run.Status == store.RunStatusDone &&
				runs.CanCompleteAccess(user, access) &&
				strings.TrimSpace(run.NotionURL) == "" &&
				s.notionExportReady(r),
		},
		PendingRequiredCount: &pendingCount,
	}
	if run.NotionURL != "" {
		detail.NotionUrl = &run.NotionURL
	}
	if run.DueDate.Valid {
		detail.DueDate = &run.DueDate.String
	}
	if run.StartedAt.Valid {
		detail.StartedAt = &run.StartedAt.String
	}
	if run.CompletedAt.Valid {
		detail.CompletedAt = &run.CompletedAt.String
	}
	return detail, true
}

func (s *Server) buildRunItemDetail(
	w http.ResponseWriter,
	r *http.Request,
	run *store.ChecklistRun,
	itemID int64,
	user *store.User,
	access store.SubjectAccess,
) (RunItemDetail, bool) {
	item, err := s.Store.RunItemByID(r.Context(), run.ID, itemID)
	if err != nil {
		if errors.Is(err, store.ErrRunItemNotFound) {
			writeAPIError(w, http.StatusNotFound, "not_found", "Point introuvable.")
			return RunItemDetail{}, false
		}
		slog.Error("load run item detail", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return RunItemDetail{}, false
	}

	events, err := s.Store.ListRunItemEvents(r.Context(), item.ID)
	if err != nil {
		slog.Error("list run item events", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return RunItemDetail{}, false
	}

	members, err := s.Store.ListSubjectMembers(r.Context(), run.SubjectID)
	if err != nil {
		slog.Error("list assignees for item", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return RunItemDetail{}, false
	}

	subject, err := s.Store.SubjectByID(r.Context(), run.SubjectID)
	if err != nil {
		slog.Error("subject for item detail", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return RunItemDetail{}, false
	}

	title, _ := s.Store.RunDisplayLabelForRun(r.Context(), run)
	editable := run.Status == store.RunStatusInProgress
	assignees := mapAssignees(members)
	mappedEvents := make([]RunItemEvent, 0, len(events))
	for _, ev := range events {
		mappedEvents = append(mappedEvents, mapRunItemEvent(ev))
	}

	detail := RunItemDetail{
		Item:         mapRunItem(*item),
		Events:       mappedEvents,
		Assignees:    &assignees,
		Capabilities: runCaps(user, access, editable),
		RunStatus:    ptrString(run.Status),
		RunTitle:     ptrString(title),
		SubjectId:    &subject.ID,
		SubjectName:  &subject.Name,
	}

	if att, err := s.Store.AttachmentByRunItemID(r.Context(), item.ID); err == nil {
		mapped := mapAttachment(att)
		detail.Attachment = &mapped
	} else if !errors.Is(err, store.ErrAttachmentNotFound) {
		slog.Error("load attachment for item detail", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return RunItemDetail{}, false
	}

	jiraConfigured := false
	if svc := s.jiraService(); svc != nil {
		cfg, ok, err := svc.Load(r.Context())
		if err != nil {
			slog.Error("load jira for item detail", "err", err)
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
			return RunItemDetail{}, false
		}
		jiraConfigured = ok && cfg.Configured()
	}
	detail.JiraConfigured = &jiraConfigured

	if link, err := s.Store.IntegrationLinkByRunItemAndType(r.Context(), item.ID, store.IntegrationTypeJira); err == nil && link != nil {
		mapped := mapJiraLink(link)
		detail.JiraLink = &mapped
	} else if err != nil && !errors.Is(err, store.ErrIntegrationLinkNotFound) {
		slog.Error("load jira link for item detail", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return RunItemDetail{}, false
	}

	return detail, true
}

func runCaps(user *store.User, access store.SubjectAccess, editable bool) RunCapabilities {
	return RunCapabilities{
		CanUpdateItems:  editable && runs.CanUpdateAccess(user, access),
		CanAssign:       editable && runs.CanAssignAccess(user, access),
		CanComplete:     editable && runs.CanCompleteAccess(user, access),
		CanExportNotion: false,
	}
}

func mapRunItem(item store.RunItem) RunItem {
	out := RunItem{
		Id:        item.ID,
		RunId:     item.RunID,
		Section:   item.Section,
		Position:  item.Position,
		Label:     item.Label,
		HelpText:  ptrString(item.HelpText),
		Required:  item.Required,
		Status:    RunItemStatus(item.Status),
		Comment:   item.Comment,
		UpdatedAt: item.UpdatedAt,
	}
	if item.AssignedTo.Valid {
		out.AssignedTo = &item.AssignedTo.Int64
	}
	if item.AssignedLogin != "" {
		out.AssignedLogin = &item.AssignedLogin
	}
	return out
}

func mapRunItemEvent(ev store.RunItemEvent) RunItemEvent {
	out := RunItemEvent{
		Id:        ev.ID,
		NewStatus: ev.NewStatus,
		Comment:   ptrString(ev.Comment),
		CreatedAt: ev.CreatedAt,
	}
	if ev.UserLogin != "" {
		out.UserLogin = &ev.UserLogin
	}
	if ev.OldStatus.Valid {
		out.OldStatus = &ev.OldStatus.String
	}
	return out
}

func mapAssignees(members []store.SubjectMember) []RunAssignee {
	out := make([]RunAssignee, 0, len(members))
	for _, m := range members {
		out = append(out, RunAssignee{
			UserId:      m.UserID,
			Login:       m.Login,
			DisplayName: m.DisplayName,
		})
	}
	return out
}

func mapRunListSummary(row store.RunListSummary) RunSummary {
	out := RunSummary{
		Id:          row.RunID,
		Title:       row.Title,
		SubjectId:   row.SubjectID,
		SubjectName: row.SubjectName,
		Status:      RunSummaryStatus(row.Status),
		CreatedAt:   row.CreatedAt,
		Progress:    RunProgress{Done: row.Done, Total: row.Total, Percent: row.Percent},
	}
	if row.DueDate.Valid {
		out.DueDate = &row.DueDate.String
	}
	if row.StartedAt.Valid {
		out.StartedAt = &row.StartedAt.String
	}
	if row.CompletedAt.Valid {
		out.CompletedAt = &row.CompletedAt.String
	}
	if row.CreatedByLogin.Valid {
		out.CreatedByLogin = &row.CreatedByLogin.String
	}
	return out
}

func mapRunWithProgress(row store.RunWithProgress, subjectName string) RunSummary {
	out := RunSummary{
		Id:          row.ID,
		Title:       row.DisplayLabel,
		SubjectId:   row.SubjectID,
		SubjectName: subjectName,
		Status:      RunSummaryStatus(row.Status),
		CreatedAt:   row.CreatedAt,
		Progress:    RunProgress{Done: row.Done, Total: row.Total, Percent: row.Percent},
	}
	if row.DueDate.Valid {
		out.DueDate = &row.DueDate.String
	}
	if row.StartedAt.Valid {
		out.StartedAt = &row.StartedAt.String
	}
	if row.CompletedAt.Valid {
		out.CompletedAt = &row.CompletedAt.String
	}
	return out
}

func progressPercent(done, total int) int {
	if total == 0 {
		return 0
	}
	return (done * 100) / total
}

func ptrString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
