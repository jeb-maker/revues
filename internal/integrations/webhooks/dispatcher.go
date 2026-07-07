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
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jeb-maker/revues/internal/features/admin/settings"
	"github.com/jeb-maker/revues/internal/store"
)

const (
	EventReviewCompleted = "review.completed"
	EventReviewItemNOK   = "review.item.nok"
	EventTest            = "webhook.test"
	maxAttempts          = 3
	requestTimeout       = 5 * time.Second
	maxRedirects         = 1
)

type SettingsLoader interface {
	LoadWebhooks(ctx context.Context) (settings.WebhookConfig, bool, error)
}

type DeliveryStore interface {
	InsertWebhookDelivery(ctx context.Context, eventID, eventType, url string, statusCode int, success bool) error
}

type RunLoader interface {
	RunByID(ctx context.Context, id int64) (*store.ChecklistRun, error)
	ProjectByID(ctx context.Context, id int64) (*store.Project, error)
	RunItemByID(ctx context.Context, runID, itemID int64) (*store.RunItem, error)
	ListRunItems(ctx context.Context, runID int64) ([]store.RunItem, error)
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
		if err := d.deliverWithRetry(ctx, cfg, target, eventID, EventTest, body); err != nil {
			slog.Error("webhook test delivery failed", "event_id", eventID, "url", redactURL(target), "err", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
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
		bg, cancel := context.WithTimeout(context.Background(), requestTimeout*time.Duration(maxAttempts)+2*time.Second)
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
			if err := d.deliverWithRetry(bg, cfg, target, eventID, eventType, body); err != nil {
				slog.Error("webhook delivery failed", "event_id", eventID, "event_type", eventType, "url", redactURL(target), "err", err)
			}
		}
	}()
}

func (d *Dispatcher) deliverWithRetry(ctx context.Context, cfg settings.WebhookConfig, target, eventID, eventType string, body []byte) error {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		statusCode, err := d.deliverOnce(ctx, cfg.Secret, target, eventID, eventType, body)
		success := err == nil && statusCode >= 200 && statusCode < 300
		if logErr := d.Store.InsertWebhookDelivery(ctx, eventID, eventType, target, statusCode, success); logErr != nil {
			slog.Error("log webhook delivery", "event_id", eventID, "err", logErr)
		}
		if success {
			return nil
		}
		if err != nil {
			lastErr = err
		} else {
			lastErr = fmt.Errorf("unexpected status %d", statusCode)
		}
		if attempt < maxAttempts {
			time.Sleep(time.Duration(attempt) * 200 * time.Millisecond)
		}
	}
	return lastErr
}

func (d *Dispatcher) deliverOnce(ctx context.Context, secret, target, eventID, eventType string, body []byte) (int, error) {
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
	project, err := d.Runs.ProjectByID(ctx, run.ProjectID)
	if err != nil {
		return ReviewCompletedData{}, fmt.Errorf("load project: %w", err)
	}
	items, err := d.Runs.ListRunItems(ctx, runID)
	if err != nil {
		return ReviewCompletedData{}, fmt.Errorf("list run items: %w", err)
	}
	return ReviewCompletedData{Review: ReviewRef{ID: run.ID, Title: run.Title, Status: run.Status, ProjectID: project.ID, ProjectName: project.Name, ClosingNote: run.ClosingNote, CompletedAt: nullString(run.CompletedAt)}, Items: itemSummary(items)}, nil
}

func (d *Dispatcher) buildReviewItemNOKPayload(ctx context.Context, runID, itemID int64) (ReviewItemNOKData, error) {
	run, err := d.Runs.RunByID(ctx, runID)
	if err != nil {
		return ReviewItemNOKData{}, fmt.Errorf("load run: %w", err)
	}
	project, err := d.Runs.ProjectByID(ctx, run.ProjectID)
	if err != nil {
		return ReviewItemNOKData{}, fmt.Errorf("load project: %w", err)
	}
	item, err := d.Runs.RunItemByID(ctx, runID, itemID)
	if err != nil {
		return ReviewItemNOKData{}, fmt.Errorf("load run item: %w", err)
	}
	return ReviewItemNOKData{Review: ReviewRef{ID: run.ID, Title: run.Title, Status: run.Status, ProjectID: project.ID, ProjectName: project.Name}, Item: ItemRef{ID: item.ID, Section: item.Section, Label: item.Label, Status: item.Status, Comment: item.Comment}}, nil
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
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme == "https" {
		return nil
	}
	if devMode && u.Scheme == "http" && isLocalhostHost(u.Hostname()) {
		return nil
	}
	return fmt.Errorf("webhook url scheme not allowed")
}

func NewSafeClient(devMode bool) *http.Client {
	return &http.Client{Timeout: requestTimeout, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= maxRedirects {
			return fmt.Errorf("too many redirects")
		}
		if err := ValidateTargetURL(req.URL.String(), devMode); err != nil {
			return err
		}
		return validateResolvedIPs(req.Context(), req.URL.Hostname(), devMode)
	}, Transport: &http.Transport{DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			host = addr
		}
		if err := validateResolvedIPs(ctx, host, devMode); err != nil {
			return nil, err
		}
		return (&net.Dialer{Timeout: requestTimeout}).DialContext(ctx, network, addr)
	}}}
}

func validateResolvedIPs(ctx context.Context, host string, devMode bool) error {
	if isLocalhostHost(host) {
		if devMode {
			return nil
		}
		return fmt.Errorf("localhost not allowed")
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return fmt.Errorf("resolve host: %w", err)
	}
	for _, ipAddr := range ips {
		if blockedIP(ipAddr.IP) {
			return fmt.Errorf("blocked ip address")
		}
	}
	return nil
}

func blockedIP(ip net.IP) bool {
	if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsPrivate() || ip.IsUnspecified() {
		return true
	}
	return ip.Equal(net.ParseIP("169.254.169.254"))
}

func isLocalhostHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
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
