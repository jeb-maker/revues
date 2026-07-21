package bugreports

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Report is one persisted bug report (JSONL record).
type Report struct {
	ID              string          `json:"id"`
	CreatedAt       string          `json:"created_at"`
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	Steps           string          `json:"steps,omitempty"`
	Severity        string          `json:"severity"`
	Source          string          `json:"source,omitempty"` // form | widget
	ReportType      string          `json:"report_type,omitempty"`
	ClientID        string          `json:"client_id,omitempty"`
	PageURL         string          `json:"page_url,omitempty"`
	UserID          int64           `json:"user_id"`
	UserLogin       string          `json:"user_login"`
	UserEmail       string          `json:"user_email,omitempty"`
	UserDisplayName string          `json:"user_display_name,omitempty"`
	UserRole        string          `json:"user_role"`
	OrgID           int64           `json:"org_id,omitempty"`
	OrgName         string          `json:"org_name,omitempty"`
	OrgRole         string          `json:"org_role,omitempty"`
	UIRunLabel      string          `json:"ui_run_label,omitempty"`
	SimpleUI        bool            `json:"simple_ui"`
	UICaps          map[string]any  `json:"ui_caps,omitempty"`
	UserAgent       string          `json:"user_agent,omitempty"`
	RequestID       string          `json:"request_id,omitempty"`
	Payload         json.RawMessage `json:"payload,omitempty"` // widget schemaVersion:1 (compact)
}

// AppendReport writes report as one JSON line under dir (creates dir if needed).
// Returns the file path written.
func AppendReport(dir string, report Report) (string, error) {
	if dir == "" {
		return "", fmt.Errorf("bug reports dir is empty")
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", fmt.Errorf("create bug reports dir: %w", err)
	}

	day := time.Now().UTC().Format("2006-01-02")
	path := filepath.Join(dir, "bug-reports-"+day+".jsonl")

	line, err := json.Marshal(report)
	if err != nil {
		return "", fmt.Errorf("marshal bug report: %w", err)
	}
	line = append(line, '\n')

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o640)
	if err != nil {
		return "", fmt.Errorf("open bug reports file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(line); err != nil {
		return "", fmt.Errorf("write bug report: %w", err)
	}
	return path, nil
}
