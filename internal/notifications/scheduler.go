package notifications

import (
	"context"
	"log/slog"
	"time"
)

// StartDueReminderScheduler runs J-1 due reminders on startup and at each interval.
func StartDueReminderScheduler(ctx context.Context, svc *Service, interval time.Duration) {
	if svc == nil {
		return
	}
	if interval <= 0 {
		interval = 24 * time.Hour
	}

	go func() {
		run := func() {
			if err := svc.SendDueReminders(ctx); err != nil {
				slog.Error("send due reminders", "err", err)
			}
		}
		run()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				run()
			}
		}
	}()
}
