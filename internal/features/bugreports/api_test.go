package bugreports

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCompactClientPayload_OmitsScreenshotDataURL(t *testing.T) {
	raw := []byte(`{
		"schemaVersion":1,
		"title":"t",
		"message":"m",
		"screenshot":{"status":"captured","mime":"image/jpeg","dataUrl":"data:image/jpeg;base64,AAAA","bytes":4}
	}`)
	got := compactClientPayload(raw)
	if strings.Contains(string(got), "data:image/jpeg") {
		t.Fatalf("dataUrl still present: %s", got)
	}
	var m map[string]any
	if err := json.Unmarshal(got, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	shot, ok := m["screenshot"].(map[string]any)
	if !ok {
		t.Fatalf("screenshot type = %T", m["screenshot"])
	}
	if shot["omitted"] != true {
		t.Fatalf("omitted = %v", shot["omitted"])
	}
	if shot["status"] != "captured" {
		t.Fatalf("status = %v", shot["status"])
	}
}

func TestNormalizeReportTypeAndSeverity(t *testing.T) {
	if got := normalizeReportType("HELP"); got != "help" {
		t.Fatalf("normalizeReportType = %q", got)
	}
	if got := normalizeReportType("nope"); got != "bug" {
		t.Fatalf("normalizeReportType default = %q", got)
	}
	if got := severityFromReportType("bug"); got != severityNormal {
		t.Fatalf("bug severity = %q", got)
	}
	if got := severityFromReportType("suggestion"); got != severityLow {
		t.Fatalf("suggestion severity = %q", got)
	}
}
