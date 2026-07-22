package bugreports

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"github.com/jeb-maker/revues/internal/web/middleware"
)

const (
	maxAPIBodyBytes = 1 << 20 // 1 MiB (screenshot dataUrl + context)
	sourceForm      = "form"
	sourceWidget    = "widget"
)

// clientReport is the @jeb-maker/reports schemaVersion:1 payload (subset used for mapping).
type clientReport struct {
	SchemaVersion int         `json:"schemaVersion"`
	ID            string      `json:"id"`
	Type          string      `json:"type"`
	Title         string      `json:"title"`
	Message       string      `json:"message"`
	Page          *clientPage `json:"page"`
}

type clientPage struct {
	URL string `json:"url"`
}

// CreateAPI accepts a JSON report from the Reports widget (POST /signaler/api).
// Trusted identity fields are taken from the session, not from client metadata.
func (h *BugReports) CreateAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "unauthorized"})
		return
	}

	ct := r.Header.Get("Content-Type")
	if !strings.HasPrefix(strings.ToLower(ct), "application/json") {
		writeJSON(w, http.StatusUnsupportedMediaType, map[string]any{"ok": false, "error": "expected application/json"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxAPIBodyBytes)
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusRequestEntityTooLarge, map[string]any{"ok": false, "error": "payload too large"})
		return
	}
	if len(strings.TrimSpace(string(raw))) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "empty body"})
		return
	}

	var client clientReport
	if unmarshalErr := json.Unmarshal(raw, &client); unmarshalErr != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid json"})
		return
	}

	title := strings.TrimSpace(client.Title)
	description := strings.TrimSpace(client.Message)
	if title == "" || description == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "title and message are required"})
		return
	}
	if len(title) > maxTitleLen {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "title too long"})
		return
	}
	if len(description) > maxDescLen {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "message too long"})
		return
	}

	pageURL := ""
	if client.Page != nil {
		pageURL = resolvePageURL(client.Page.URL)
	}
	if pageURL == "" {
		pageURL = resolvePageURL(r.Referer())
	}

	ctxSummary := buildContextSummary(r, user)
	ctxSummary.PageURL = pageURL

	reportType := normalizeReportType(client.Type)
	now := time.Now().UTC()
	report := Report{
		ID:              uuid.NewString(),
		CreatedAt:       now.Format(time.RFC3339),
		Title:           title,
		Description:     description,
		Severity:        severityFromReportType(reportType),
		Source:          sourceWidget,
		ReportType:      reportType,
		ClientID:        strings.TrimSpace(client.ID),
		PageURL:         pageURL,
		UserID:          user.ID,
		UserLogin:       user.Login,
		UserEmail:       user.Email,
		UserDisplayName: user.DisplayName,
		UserRole:        user.Role,
		OrgID:           ctxSummary.OrgID,
		OrgName:         ctxSummary.OrgName,
		OrgRole:         ctxSummary.OrgRole,
		UIRunLabel:      ctxSummary.UIRunLabel,
		SimpleUI:        ctxSummary.SimpleUI,
		UICaps:          ctxSummary.UICapsMap(),
		UserAgent:       r.UserAgent(),
		RequestID:       chimw.GetReqID(r.Context()),
		Payload:         compactClientPayload(raw),
	}

	path, err := AppendReport(h.ReportsDir, report)
	if err != nil {
		slog.Error("persist bug report api", "err", err, "request_id", report.RequestID)
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "persist failed"})
		return
	}

	slog.Info("bug report submitted",
		"report_id", report.ID,
		"source", report.Source,
		"report_type", report.ReportType,
		"path", path,
		"user_id", report.UserID,
		"severity", report.Severity,
		"page_url", report.PageURL,
		"org_id", report.OrgID,
		"simple_ui", report.SimpleUI,
		"request_id", report.RequestID,
		"client_id", report.ClientID,
	)

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":         true,
		"id":         report.ID,
		"created_at": report.CreatedAt,
	})
}

func writeJSON(w http.ResponseWriter, status int, body map[string]any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func normalizeReportType(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "bug", "help", "suggestion", "question":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return "bug"
	}
}

func severityFromReportType(reportType string) string {
	switch reportType {
	case "bug":
		return severityNormal
	default:
		return severityLow
	}
}

// compactClientPayload stores diagnostics without embedding screenshot dataUrl bytes.
func compactClientPayload(raw json.RawMessage) json.RawMessage {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return raw
	}
	if shot, ok := m["screenshot"].(map[string]any); ok {
		compact := make(map[string]any, 6)
		for _, k := range []string{"status", "error", "mime", "bytes", "method", "width", "height"} {
			if v, exists := shot[k]; exists {
				compact[k] = v
			}
		}
		if dataURL, ok := shot["dataUrl"].(string); ok && dataURL != "" {
			compact["data_url_bytes"] = len(dataURL)
			compact["omitted"] = true
		}
		m["screenshot"] = compact
	}
	// Client metadata is untrusted for identity; keep for diagnostics but strip obvious id overrides noise.
	out, err := json.Marshal(m)
	if err != nil {
		return raw
	}
	return out
}
