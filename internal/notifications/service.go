package notifications

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jeb-maker/revues/internal/features/admin/settings"
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

// NotifyRunCompleted emails all project members when a review is completed.
func (s *Service) NotifyRunCompleted(ctx context.Context, runID int64) {
	if s == nil || s.Store == nil || s.Settings == nil {
		return
	}

	run, err := s.Store.RunByID(ctx, runID)
	if err != nil {
		slog.Error("notification run completed load run", "run_id", runID, "err", err)
		return
	}

	project, err := s.Store.ProjectByID(ctx, run.ProjectID)
	if err != nil {
		slog.Error("notification run completed load project", "run_id", runID, "err", err)
		return
	}

	members, err := s.Store.ListProjectMembers(ctx, run.ProjectID)
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

		subject := fmt.Sprintf("Revue terminée : %s", run.Title)
		body := fmt.Sprintf(
			"La revue « %s » du projet « %s » est terminée.\n\n%s/runs/%d\n",
			run.Title, project.Name, strings.TrimRight(s.BaseURL, "/"), run.ID,
		)
		messages = append(messages, emailMessage{to: to, subject: subject, body: body})
	}

	s.dispatch(messages)
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

	project, err := s.Store.ProjectByID(ctx, run.ProjectID)
	if err != nil {
		slog.Error("notification item assigned load project", "run_id", runID, "item_id", itemID, "err", err)
		return
	}

	subject := fmt.Sprintf("Point assigné : %s", item.Label)
	body := fmt.Sprintf(
		"Le point « %s » de la revue « %s » (projet « %s ») vous a été assigné.\n\n%s/runs/%d/items/%d\n",
		item.Label, run.Title, project.Name, strings.TrimRight(s.BaseURL, "/"), run.ID, item.ID,
	)
	s.dispatch([]emailMessage{{to: to, subject: subject, body: body}})
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
		s.sendDueReminder(ctx, &run, tomorrow)
	}
	return nil
}

func (s *Service) sendDueReminder(ctx context.Context, run *store.ChecklistRun, dueDay string) {
	to := s.runResponsibleEmail(ctx, run)
	if to == "" {
		return
	}

	project, err := s.Store.ProjectByID(ctx, run.ProjectID)
	if err != nil {
		slog.Error("notification due reminder load project", "run_id", run.ID, "err", err)
		return
	}

	dueLabel := dueDay
	if run.DueDate.Valid {
		dueLabel = run.DueDate.String
	}

	subject := fmt.Sprintf("Échéance demain : %s", run.Title)
	body := fmt.Sprintf(
		"La revue « %s » du projet « %s » arrive à échéance demain (%s).\n\n%s/runs/%d\n",
		run.Title, project.Name, dueLabel, strings.TrimRight(s.BaseURL, "/"), run.ID,
	)
	s.dispatch([]emailMessage{{to: to, subject: subject, body: body}})
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

	members, err := s.Store.ListProjectMembers(ctx, run.ProjectID)
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

func (s *Service) dispatch(messages []emailMessage) {
	if len(messages) == 0 {
		return
	}

	go func() {
		sendCtx, cancel := context.WithTimeout(context.Background(), sendTimeout)
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
