package templates_test

import (
	"testing"

	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/templates"
)

func TestPageData_ReportsMetadataIncludesP3Caps(t *testing.T) {
	data := templates.PageData{
		User: &store.User{ID: 1, Login: "alice", Role: "editor"},
		ActiveOrganization: &store.Organization{
			ID:         7,
			Name:       "Acme",
			UIRunLabel: "revues",
		},
		SimpleUI:    true,
		HasJira:     true,
		HasNotion:   true,
		HasWebhooks: false,
		HasEvidence: true,
	}
	meta := data.ReportsMetadata()
	caps, ok := meta["ui_caps"].(map[string]any)
	if !ok {
		t.Fatalf("ui_caps missing: %#v", meta)
	}
	for _, key := range []string{"has_jira", "has_notion", "has_webhooks", "has_evidence"} {
		if _, present := caps[key]; !present {
			t.Fatalf("ui_caps missing %q: %#v", key, caps)
		}
	}
	if caps["has_jira"] != true || caps["has_notion"] != true {
		t.Fatalf("has_jira/has_notion = %#v", caps)
	}
	if caps["has_webhooks"] != false || caps["has_evidence"] != true {
		t.Fatalf("has_webhooks/has_evidence = %#v", caps)
	}
	if caps["simple_ui"] != true {
		t.Fatal("simple_ui should remain true alongside P3")
	}
}

func TestBugReportContext_UICapsMapIncludesP3(t *testing.T) {
	c := templates.BugReportContext{
		HasJira:     true,
		HasNotion:   false,
		HasWebhooks: true,
		HasEvidence: false,
	}
	m := c.UICapsMap()
	if m["has_jira"] != true || m["has_webhooks"] != true {
		t.Fatalf("UICapsMap = %#v", m)
	}
	if m["has_notion"] != false || m["has_evidence"] != false {
		t.Fatalf("UICapsMap negatives = %#v", m)
	}
}
