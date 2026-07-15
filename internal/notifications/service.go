package notifications

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jeb-maker/revues/internal/features/admin/settings"
	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
)

const sendTimeout = 30 * time.Second

// SettingsLoader loads SMTP configuration for outbound email.
type SettingsLoader interface {
	LoadSMTP(ctx context.Context) (settings.SMTPConfig, bool, error)
}

// Service sends business notification emails asynchronously.
type Service struct {
	Store    *store.Store
	Settings SettingsLoader
	BaseURL  string
}

type emailMessage struct {
	to      string
	subject string
	body    string
}

// NotifyRunCompleted emails org members when a review is completed.
func (s *Service) NotifyRunCompleted(ctx context.Context, runID int64) {
	if s == nil || s.Store == nil || s.Settings == nil {
		return
	}

	run, err := s.Store.RunByID(ctx, runID)
	if err != nil {
		slog.Error("notification run completed load run", "run_id", runID, "err", err)
		return
	}

	subject, err := s.Store.SubjectByID(ctx, run.SubjectID)
	if err != nil {
		slog.Error("notification run completed load subject", "run_id", runID, "err", err)
		return
	}

	displayLabel, err := s.Store.RunDisplayLabelForRun(ctx, run)
	if err != nil {
		slog.Error("notification run completed display label", "run_id", runID, "err", err)
		return
	}

	members, err := s.Store.ListSubjectMembers(ctx, run.SubjectID)
	if err != nil {
		slog.Error("notification run completed list members", "run_id", runID, "err", err)
		return
	}

	seen := make(map[string]struct{})
	var messages []emailMessage
	for _, member := range members {
		to := strings.TrimSpace(member.Email)
		if to == "" {
			continue
		}
		if _, ok := seen[to]; ok {
			continue
		}
		seen[to] = struct{}{}

		subjectLine := fmt.Sprintf("Revue terminée : %s", displayLabel)
		body := fmt.Sprintf(
			"La revue « %s » du sujet « %s » est terminée.\n\n%s/runs/%d\n",
			displayLabel, subject.Name, strings.TrimRight(s.BaseURL, "/"), run.ID,
		)
		messages = append(messages, emailMessage{to: to, subject: subjectLine, body: body})
	}

	s.dispatch(ctx, messages)
}

// NotifyItemAssigned emails the assignee when a checklist point is assigned.
func (s *Service) NotifyItemAssigned(ctx context.Context, runID, itemID int64) {
	if s == nil || s.Store == nil || s.Settings == nil {
		return
	}

	run, err := s.Store.RunByID(ctx, runID)
	if err != nil {
		slog.Error("notification item assigned load run", "run_id", runID, "item_id", itemID, "err", err)
		return
	}

	item, err := s.Store.RunItemByID(ctx, runID, itemID)
	if err != nil {
		slog.Error("notification item assigned load item", "run_id", runID, "item_id", itemID, "err", err)
		return
	}
	if !item.AssignedTo.Valid {
		return
	}

	assignee, err := s.Store.UserByID(ctx, item.AssignedTo.Int64)
	if err != nil {
		slog.Error("notification item assigned load assignee", "run_id", runID, "item_id", itemID, "err", err)
		return
	}
	to := strings.TrimSpace(assignee.Email)
	if to == "" {
		return
	}

	subjectEntity, err := s.Store.SubjectByID(ctx, run.SubjectID)
	if err != nil {
		slog.Error("notification item assigned load subject", "run_id", runID, "item_id", itemID, "err", err)
		return
	}

	displayLabel, err := s.Store.RunDisplayLabelForRun(ctx, run)
	if err != nil {
		slog.Error("notification item assigned display label", "run_id", runID, "item_id", itemID, "err", err)
		return
	}

	subjectLine := fmt.Sprintf("Point assigné : %s", item.Label)
	body := fmt.Sprintf(
		"Le point « %s » de la revue « %s » (sujet « %s ») vous a été assigné.\n\n%s/runs/%d/items/%d\n",
		item.Label, displayLabel, subjectEntity.Name, strings.TrimRight(s.BaseURL, "/"), run.ID, item.ID,
	)
	s.dispatch(ctx, []emailMessage{{to: to, subject: subjectLine, body: body}})
}

// SendDueReminders emails run responsibles for reviews due tomorrow (J-1).
func (s *Service) SendDueReminders(ctx context.Context) error {
	if s == nil || s.Store == nil || s.Settings == nil {
		return nil
	}

	tomorrow := time.Now().UTC().Add(24 * time.Hour).Format("2006-01-02")
	runs, err := s.Store.ListRunsDueOn(ctx, tomorrow)
	if err != nil {
		return fmt.Errorf("list runs due on %s: %w", tomorrow, err)
	}

	for _, run := range runs {
		subjectEntity, err := s.Store.SubjectByIDUnscoped(ctx, run.SubjectID)
		if err != nil {
			slog.Error("notification due reminder load subject", "run_id", run.ID, "err", err)
			continue
		}
		runCtx := orgctx.WithOrganizationID(ctx, subjectEntity.OrganizationID)
		s.sendDueReminder(runCtx, &run, tomorrow)
	}
	return nil
}

func (s *Service) sendDueReminder(ctx context.Context, run *store.ChecklistRun, dueDay string) {
	to := s.runResponsibleEmail(ctx, run)
	if to == "" {
		return
	}

	subjectEntity, err := s.Store.SubjectByID(ctx, run.SubjectID)
	if err != nil {
		slog.Error("notification due reminder load subject", "run_id", run.ID, "err", err)
		return
	}

	displayLabel, err := s.Store.RunDisplayLabelForRun(ctx, run)
	if err != nil {
		slog.Error("notification due reminder display label", "run_id", run.ID, "err", err)
		return
	}

	dueLabel := dueDay
	if run.DueDate.Valid {
		dueLabel = run.DueDate.String
	}

	subjectLine := fmt.Sprintf("Échéance demain : %s", displayLabel)
	body := fmt.Sprintf(
		"La revue « %s » du sujet « %s » arrive à échéance demain (%s).\n\n%s/runs/%d\n",
		displayLabel, subjectEntity.Name, dueLabel, strings.TrimRight(s.BaseURL, "/"), run.ID,
	)
	s.dispatch(ctx, []emailMessage{{to: to, subject: subjectLine, body: body}})
}

func (s *Service) runResponsibleEmail(ctx context.Context, run *store.ChecklistRun) string {
	if run.CreatedBy.Valid {
		user, err := s.Store.UserByID(ctx, run.CreatedBy.Int64)
		if err == nil {
			if email := strings.TrimSpace(user.Email); email != "" {
				return email
			}
		}
	}

	members, err := s.Store.ListSubjectMembers(ctx, run.SubjectID)
	if err != nil {
		slog.Error("notification responsible list members", "run_id", run.ID, "err", err)
		return ""
	}
	for _, member := range members {
		if member.Role == "lead" {
			if email := strings.TrimSpace(member.Email); email != "" {
				return email
			}
		}
	}
	return ""
}

func (s *Service) dispatch(ctx context.Context, messages []emailMessage) {
	if len(messages) == 0 {
		return
	}

	go func() {
		sendCtx, cancel := context.WithTimeout(ctx, sendTimeout)
		defer cancel()

		cfg, ok, err := s.Settings.LoadSMTP(sendCtx)
		if err != nil {
			slog.Error("load smtp settings for notification", "err", err)
			return
		}
		if !ok || !cfg.Enabled() {
			return
		}

		mailer := Mailer{Config: cfg}
		for _, msg := range messages {
			if err := mailer.Send(sendCtx, msg.to, msg.subject, msg.body); err != nil {
				slog.Error("send notification email", "to", msg.to, "subject", msg.subject, "err", err)
			}
		}
	}()
}
