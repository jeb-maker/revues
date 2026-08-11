package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jeb-maker/revues/internal/features/admin/settings"
	"github.com/jeb-maker/revues/internal/safehttp"
	"github.com/jeb-maker/revues/internal/store"
)

const (
	EventReviewCompleted = "review.completed"
	EventReviewItemNOK   = "review.item.nok"
	EventTest            = "webhook.test"

	requestTimeout = 5 * time.Second
	maxRedirects   = 1

	// MaxAttempts is the hard cap on HTTP tries per delivery (including the first).
	MaxAttempts = 5
	// DeliveryTTL is how long a pending delivery may stay in the queue.
	DeliveryTTL = 24 * time.Hour
	// DrainInterval is the in-process cron cadence for pending deliveries.
	DrainInterval = time.Minute
	// DrainBatchSize limits rows processed per drain tick.
	DrainBatchSize = 50
	// maxBackoff caps exponential backoff between attempts.
	maxBackoff = 30 * time.Minute
	// baseBackoff is the delay after the first failure (attempt 1 → 2).
	baseBackoff = time.Minute
)

type SettingsLoader interface {
	LoadWebhooks(ctx context.Context) (settings.WebhookConfig, bool, error)
}

type DeliveryStore interface {
	EnqueueWebhookDelivery(ctx context.Context, eventID, eventType, url string, payload []byte, nextAttemptAt, expiresAt time.Time) (int64, error)
	ListDueWebhookDeliveries(ctx context.Context, now time.Time, limit int) ([]store.WebhookDelivery, error)
	UpdateWebhookDeliveryAttempt(ctx context.Context, id int64, statusCode int, success bool, attempts int, nextAttemptAt *time.Time, state, lastError string) error
}

type RunLoader interface {
	RunByID(ctx context.Context, id int64) (*store.ChecklistRun, error)
	SubjectByID(ctx context.Context, id int64) (*store.Subject, error)
	RunItemByID(ctx context.Context, runID, itemID int64) (*store.RunItem, error)
	ListRunItems(ctx context.Context, runID int64) ([]store.RunItem, error)
	RunDisplayLabelForRun(ctx context.Context, run *store.ChecklistRun) (string, error)
}

type Dispatcher struct {
	Settings SettingsLoader
	Store    DeliveryStore
	Runs     RunLoader
	DevMode  bool
	Now      func() time.Time
	Client   *http.Client
}

func (d *Dispatcher) EmitReviewCompleted(ctx context.Context, runID int64) {
	d.emitAsync(ctx, EventReviewCompleted, func(ctx context.Context) (any, error) {
		return d.buildReviewCompletedPayload(ctx, runID)
	})
}

func (d *Dispatcher) EmitReviewItemNOK(ctx context.Context, runID, itemID int64) {
	d.emitAsync(ctx, EventReviewItemNOK, func(ctx context.Context) (any, error) {
		return d.buildReviewItemNOKPayload(ctx, runID, itemID)
	})
}

func (d *Dispatcher) SendTest(ctx context.Context) error {
	cfg, ok, err := d.Settings.LoadWebhooks(ctx)
	if err != nil {
		return fmt.Errorf("load webhooks: %w", err)
	}
	if !ok || !cfg.Enabled() {
		return fmt.Errorf("webhooks non configurés")
	}
	eventID := newEventID()
	body, err := json.Marshal(Envelope{EventID: eventID, EventType: EventTest, OccurredAt: d.now().UTC().Format(time.RFC3339), Data: TestData{Message: "Ceci est un événement de test depuis Revues."}})
	if err != nil {
		return fmt.Errorf("marshal test payload: %w", err)
	}
	var firstErr error
	for _, target := range cfg.URLs {
		if err := d.enqueueAndAttempt(ctx, cfg, target, eventID, EventTest, body); err != nil {
			slog.Error("webhook test delivery failed", "event_id", eventID, "url", redactURL(target), "err", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// Drain processes due pending deliveries (cron or opportunistic request path).
func (d *Dispatcher) Drain(ctx context.Context) error {
	if d == nil || d.Settings == nil || d.Store == nil {
		return nil
	}
	cfg, ok, err := d.Settings.LoadWebhooks(ctx)
	if err != nil {
		return fmt.Errorf("load webhooks for drain: %w", err)
	}
	if !ok || !cfg.Enabled() {
		return nil
	}
	due, err := d.Store.ListDueWebhookDeliveries(ctx, d.now(), DrainBatchSize)
	if err != nil {
		return err
	}
	for _, del := range due {
		if err := d.processQueued(ctx, cfg.Secret, del); err != nil {
			slog.Error("webhook drain attempt failed", "delivery_id", del.ID, "event_id", del.EventID, "url", redactURL(del.URL), "err", err)
		}
	}
	return nil
}

// StartDrainScheduler runs Drain on startup and every interval in the same process.
func StartDrainScheduler(ctx context.Context, d *Dispatcher, interval time.Duration) {
	if d == nil {
		return
	}
	if interval <= 0 {
		interval = DrainInterval
	}
	go func() {
		run := func() {
			drainCtx, cancel := context.WithTimeout(context.Background(), requestTimeout*time.Duration(DrainBatchSize)+5*time.Second)
			defer cancel()
			if err := d.Drain(drainCtx); err != nil {
				slog.Error("webhook drain", "err", err)
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

// BackoffAfterAttempt returns the delay before the next attempt after a failed try.
// attempt is the number of failed attempts so far (1 after first failure).
func BackoffAfterAttempt(attempt int) time.Duration {
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

func (d *Dispatcher) emitAsync(ctx context.Context, eventType string, build func(context.Context) (any, error)) {
	if d == nil || d.Settings == nil || d.Store == nil || d.Runs == nil {
		return
	}
	cfg, ok, err := d.Settings.LoadWebhooks(ctx)
	if err != nil {
		slog.Error("load webhooks for emit", "event_type", eventType, "err", err)
		return
	}
	if !ok || !cfg.Enabled() || !cfg.EventEnabled(eventType) {
		return
	}
	go func() {
		bg, cancel := context.WithTimeout(context.Background(), requestTimeout+2*time.Second)
		defer cancel()
		data, err := build(bg)
		if err != nil {
			slog.Error("build webhook payload", "event_type", eventType, "err", err)
			return
		}
		eventID := newEventID()
		body, err := json.Marshal(Envelope{EventID: eventID, EventType: eventType, OccurredAt: d.now().UTC().Format(time.RFC3339), Data: data})
		if err != nil {
			slog.Error("marshal webhook payload", "event_type", eventType, "err", err)
			return
		}
		for _, target := range cfg.URLs {
			if err := d.enqueueAndAttempt(bg, cfg, target, eventID, eventType, body); err != nil {
				slog.Error("webhook delivery failed", "event_id", eventID, "event_type", eventType, "url", redactURL(target), "err", err)
			}
		}
		// Opportunistic drain of other due rows on the request/emit path.
		drainCtx, drainCancel := context.WithTimeout(context.Background(), requestTimeout*3)
		defer drainCancel()
		if err := d.Drain(drainCtx); err != nil {
			slog.Error("webhook opportunistic drain", "err", err)
		}
	}()
}

func (d *Dispatcher) enqueueAndAttempt(ctx context.Context, cfg settings.WebhookConfig, target, eventID, eventType string, body []byte) error {
	now := d.now()
	expires := now.Add(DeliveryTTL)
	id, err := d.Store.EnqueueWebhookDelivery(ctx, eventID, eventType, target, body, now, expires)
	if err != nil {
		return err
	}
	del := store.WebhookDelivery{
		ID:        id,
		EventID:   eventID,
		EventType: eventType,
		URL:       target,
		Payload:   body,
		Attempts:  0,
		ExpiresAt: expires.UTC().Format(time.RFC3339),
		State:     store.WebhookDeliveryPending,
	}
	return d.processQueued(ctx, cfg.Secret, del)
}

func (d *Dispatcher) processQueued(ctx context.Context, secret string, del store.WebhookDelivery) error {
	now := d.now()
	if del.ExpiresAt != "" {
		if exp, err := time.Parse(time.RFC3339, del.ExpiresAt); err == nil && !now.Before(exp) {
			if updErr := d.Store.UpdateWebhookDeliveryAttempt(ctx, del.ID, 0, false, del.Attempts, nil, store.WebhookDeliveryPoison, "ttl expired"); updErr != nil {
				return updErr
			}
			return fmt.Errorf("ttl expired")
		}
	}

	statusCode, err := d.deliverOnce(ctx, secret, del.URL, del.EventID, del.EventType, del.Payload)
	attempts := del.Attempts + 1
	success := err == nil && statusCode >= 200 && statusCode < 300
	if success {
		return d.Store.UpdateWebhookDeliveryAttempt(ctx, del.ID, statusCode, true, attempts, nil, store.WebhookDeliveryDone, "")
	}

	lastErr := "unexpected status"
	if err != nil {
		lastErr = err.Error()
	} else {
		lastErr = fmt.Sprintf("unexpected status %d", statusCode)
	}

	persist := func(next *time.Time, state string) error {
		if updErr := d.Store.UpdateWebhookDeliveryAttempt(ctx, del.ID, statusCode, false, attempts, next, state, lastErr); updErr != nil {
			return updErr
		}
		return fmt.Errorf("%s", lastErr)
	}

	// Permanent policy failures (anti-SSRF) become poison immediately.
	if err != nil && isPermanentDeliveryError(err) {
		return persist(nil, store.WebhookDeliveryPoison)
	}

	if attempts >= MaxAttempts {
		return persist(nil, store.WebhookDeliveryPoison)
	}

	next := now.Add(BackoffAfterAttempt(attempts))
	if del.ExpiresAt != "" {
		if exp, parseErr := time.Parse(time.RFC3339, del.ExpiresAt); parseErr == nil && !next.Before(exp) {
			lastErr += "; next backoff past ttl"
			return persist(nil, store.WebhookDeliveryPoison)
		}
	}
	return persist(&next, store.WebhookDeliveryPending)
}

func isPermanentDeliveryError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "scheme not allowed") ||
		strings.Contains(msg, "blocked ip") ||
		strings.Contains(msg, "localhost not allowed") ||
		strings.Contains(msg, "invalid url")
}

func (d *Dispatcher) deliverOnce(ctx context.Context, secret, target, eventID, eventType string, body []byte) (int, error) {
	// Anti-SSRF re-checked on every attempt (URL + resolved IPs + dial).
	if err := ValidateTargetURL(target, d.DevMode); err != nil {
		return 0, err
	}
	if err := validateResolvedIPs(ctx, hostnameFromURL(target), d.DevMode); err != nil {
		return 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Revues-Webhooks/1.0")
	req.Header.Set("X-Revues-Event-Id", eventID)
	req.Header.Set("X-Revues-Event-Type", eventType)
	req.Header.Set("X-Revues-Signature", SignBody(secret, body))
	resp, err := d.httpClient().Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _, _ = io.Copy(io.Discard, resp.Body); _ = resp.Body.Close() }()
	return resp.StatusCode, nil
}

func (d *Dispatcher) httpClient() *http.Client {
	if d.Client != nil {
		return d.Client
	}
	return NewSafeClient(d.DevMode)
}

func (d *Dispatcher) now() time.Time {
	if d.Now != nil {
		return d.Now()
	}
	return time.Now()
}

func (d *Dispatcher) buildReviewCompletedPayload(ctx context.Context, runID int64) (ReviewCompletedData, error) {
	run, err := d.Runs.RunByID(ctx, runID)
	if err != nil {
		return ReviewCompletedData{}, fmt.Errorf("load run: %w", err)
	}
	project, err := d.Runs.SubjectByID(ctx, run.SubjectID)
	if err != nil {
		return ReviewCompletedData{}, fmt.Errorf("load subject: %w", err)
	}
	items, err := d.Runs.ListRunItems(ctx, runID)
	if err != nil {
		return ReviewCompletedData{}, fmt.Errorf("list run items: %w", err)
	}
	displayLabel, err := d.Runs.RunDisplayLabelForRun(ctx, run)
	if err != nil {
		return ReviewCompletedData{}, fmt.Errorf("run display label: %w", err)
	}
	return ReviewCompletedData{Review: ReviewRef{ID: run.ID, DisplayLabel: displayLabel, Status: run.Status, SubjectID: project.ID, SubjectName: project.Name, ClosingNote: run.ClosingNote, CompletedAt: nullString(run.CompletedAt)}, Items: itemSummary(items)}, nil
}

func (d *Dispatcher) buildReviewItemNOKPayload(ctx context.Context, runID, itemID int64) (ReviewItemNOKData, error) {
	run, err := d.Runs.RunByID(ctx, runID)
	if err != nil {
		return ReviewItemNOKData{}, fmt.Errorf("load run: %w", err)
	}
	project, err := d.Runs.SubjectByID(ctx, run.SubjectID)
	if err != nil {
		return ReviewItemNOKData{}, fmt.Errorf("load subject: %w", err)
	}
	item, err := d.Runs.RunItemByID(ctx, runID, itemID)
	if err != nil {
		return ReviewItemNOKData{}, fmt.Errorf("load run item: %w", err)
	}
	displayLabel, err := d.Runs.RunDisplayLabelForRun(ctx, run)
	if err != nil {
		return ReviewItemNOKData{}, fmt.Errorf("run display label: %w", err)
	}
	return ReviewItemNOKData{Review: ReviewRef{ID: run.ID, DisplayLabel: displayLabel, Status: run.Status, SubjectID: project.ID, SubjectName: project.Name}, Item: ItemRef{ID: item.ID, Section: item.Section, Label: item.Label, Status: item.Status, Comment: item.Comment}}, nil
}

func SignBody(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func VerifySignature(secret string, body []byte, signature string) bool {
	return hmac.Equal([]byte(SignBody(secret, body)), []byte(strings.TrimSpace(signature)))
}

func ValidateTargetURL(raw string, devMode bool) error {
	if err := safehttp.ValidateURL(raw, devMode); err != nil {
		if strings.Contains(err.Error(), "scheme") {
			return fmt.Errorf("webhook url scheme not allowed")
		}
		return err
	}
	return nil
}

func NewSafeClient(devMode bool) *http.Client {
	return safehttp.NewClient(safehttp.Options{
		AllowDevLocalhost: devMode,
		Timeout:           requestTimeout,
		DialTimeout:       requestTimeout,
		MaxRedirects:      maxRedirects,
	})
}

func validateResolvedIPs(ctx context.Context, host string, devMode bool) error {
	_, err := safehttp.ResolveAllowedIP(ctx, host, devMode)
	return err
}

func newEventID() string { return uuid.NewString() }

func nullString(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}

func itemSummary(items []store.RunItem) ItemsSummary {
	var s ItemsSummary
	for _, item := range items {
		switch item.Status {
		case store.RunItemStatusOK:
			s.OK++
		case store.RunItemStatusNOK:
			s.NOK++
		case store.RunItemStatusNA:
			s.NA++
		default:
			s.Pending++
		}
	}
	s.Total = len(items)
	return s
}

func redactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	if u.User != nil {
		u.User = url.UserPassword("redacted", "redacted")
	}
	return u.String()
}

func hostnameFromURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	return u.Hostname()
}
