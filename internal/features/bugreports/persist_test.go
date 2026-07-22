package bugreports

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAppendReport_CreatesDirAndJSONL(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "bug-reports")
	report := Report{
		ID:          "abc",
		CreatedAt:   "2026-07-19T12:00:00Z",
		Title:       "t",
		Description: "d",
		Severity:    "normal",
		UserID:      1,
		UserLogin:   "u",
		UserRole:    "reader",
	}
	path, err := AppendReport(dir, report)
	if err != nil {
		t.Fatalf("AppendReport: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var got Report
	if err := json.Unmarshal(raw[:len(raw)-1], &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.ID != "abc" || got.Title != "t" {
		t.Fatalf("got %+v", got)
	}
}

func TestReportsDirFromAttachments(t *testing.T) {
	got := ReportsDirFromAttachments("data/attachments")
	if got != "data/bug-reports" {
		t.Fatalf("got %q", got)
	}
}
