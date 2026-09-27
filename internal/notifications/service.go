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

const (
	sendTimeout = 30 * time.Second
	// MaxAttempts is the hard cap on SMTP tries per delivery (including the first).
	MaxAttempts = 5
	// DeliveryTTL is how long a pending email may stay in the queue.
	DeliveryTTL = 24 * time.Hour
	// DrainInterval is the in-process cron cadence for pending emails.
	DrainInterval = time.Minute
	// DrainBatchSize limits rows processed per drain tick.
	DrainBatchSize = 50
	baseBackoff    = time.Minute
	maxBackoff     = 30 * time.Minute
)

// SettingsLoader loads SMTP configuration for outbound email.
type SettingsLoader interface {
	LoadSMTP(ctx context.Context) (settings.SMTPConfig, bool, error)
}

// Service sends business notification emails via a durable SQL queue.
type Service struct {
	Store    *store.Store
	Settings SettingsLoader
	BaseURL  string
	Now      func() time.Time
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

func (s *Service) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

func (s *Service) dispatch(ctx context.Context, messages []emailMessage) {
	if len(messages) == 0 || s == nil || s.Store == nil || s.Settings == nil {
		return
	}

	orgID, hasOrg := orgctx.OrganizationID(ctx)
	go func() {
		bg := context.Background()
		if hasOrg {
			bg = orgctx.WithOrganizationID(bg, orgID)
		}
		bg, cancel := context.WithTimeout(bg, sendTimeout+2*time.Second)
		defer cancel()

		now := s.now()
		expires := now.Add(DeliveryTTL)
		for _, msg := range messages {
			id, err := s.Store.EnqueueEmailDelivery(bg, orgID, msg.to, msg.subject, msg.body, now, expires)
			if err != nil {
				slog.Error("enqueue notification email", "to", msg.to, "subject", msg.subject, "err", err)
				continue
			}
			del := store.EmailDelivery{
				ID:             id,
				OrganizationID: orgID,
				ToAddress:      msg.to,
				Subject:        msg.subject,
				Body:           msg.body,
				Attempts:       0,
				ExpiresAt:      expires.UTC().Format(time.RFC3339),
				State:          store.EmailDeliveryPending,
			}
			if err := s.processQueued(bg, del); err != nil {
				slog.Error("send notification email", "to", msg.to, "subject", msg.subject, "err", err)
			}
		}

		drainCtx, drainCancel := context.WithTimeout(context.Background(), sendTimeout*3)
		defer drainCancel()
		if err := s.Drain(drainCtx); err != nil {
			slog.Error("email opportunistic drain", "err", err)
		}
	}()
}

// Drain attempts due pending email deliveries (all orgs).
func (s *Service) Drain(ctx context.Context) error {
	if s == nil || s.Store == nil || s.Settings == nil {
		return nil
	}
	due, err := s.Store.ListDueEmailDeliveries(ctx, s.now(), DrainBatchSize)
	if err != nil {
		return err
	}
	for _, del := range due {
		orgCtx := ctx
		if del.OrganizationID > 0 {
			orgCtx = orgctx.WithOrganizationID(ctx, del.OrganizationID)
		}
		if err := s.processQueued(orgCtx, del); err != nil {
			slog.Error("email drain attempt failed", "delivery_id", del.ID, "to", del.ToAddress, "err", err)
		}
	}
	return nil
}

func (s *Service) processQueued(ctx context.Context, del store.EmailDelivery) error {
	now := s.now()
	if del.ExpiresAt != "" {
		if exp, err := time.Parse(time.RFC3339, del.ExpiresAt); err == nil && !now.Before(exp) {
			if updErr := s.Store.UpdateEmailDeliveryAttempt(ctx, del.ID, del.Attempts, nil, store.EmailDeliveryPoison, "ttl expired"); updErr != nil {
				return updErr
			}
			return fmt.Errorf("ttl expired")
		}
	}

	cfg, ok, err := s.Settings.LoadSMTP(ctx)
	if err != nil {
		return s.failAttempt(ctx, del, now, fmt.Errorf("load smtp: %w", err))
	}
	if !ok || !cfg.Enabled() {
		// SMTP disabled: keep pending with short backoff so admin can enable later within TTL.
		return s.failAttempt(ctx, del, now, fmt.Errorf("smtp not configured"))
	}

	sendCtx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()
	mailer := Mailer{Config: cfg}
	if err := mailer.Send(sendCtx, del.ToAddress, del.Subject, del.Body); err != nil {
		return s.failAttempt(ctx, del, now, err)
	}
	return s.Store.UpdateEmailDeliveryAttempt(ctx, del.ID, del.Attempts+1, nil, store.EmailDeliveryDone, "")
}

func (s *Service) failAttempt(ctx context.Context, del store.EmailDelivery, now time.Time, sendErr error) error {
	attempts := del.Attempts + 1
	lastErr := sendErr.Error()
	if attempts >= MaxAttempts {
		if updErr := s.Store.UpdateEmailDeliveryAttempt(ctx, del.ID, attempts, nil, store.EmailDeliveryPoison, lastErr); updErr != nil {
			return updErr
		}
		return sendErr
	}
	next := now.Add(emailBackoffAfterAttempt(attempts))
	if del.ExpiresAt != "" {
		if exp, parseErr := time.Parse(time.RFC3339, del.ExpiresAt); parseErr == nil && !next.Before(exp) {
			lastErr += "; next backoff past ttl"
			if updErr := s.Store.UpdateEmailDeliveryAttempt(ctx, del.ID, attempts, nil, store.EmailDeliveryPoison, lastErr); updErr != nil {
				return updErr
			}
			return sendErr
		}
	}
	if updErr := s.Store.UpdateEmailDeliveryAttempt(ctx, del.ID, attempts, &next, store.EmailDeliveryPending, lastErr); updErr != nil {
		return updErr
	}
	return sendErr
}

func emailBackoffAfterAttempt(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	shift := attempt - 1
	if shift > 8 {
		shift = 8
	}
	d := baseBackoff << shift
	if d > maxBackoff {
		return maxBackoff
	}
	return d
}
