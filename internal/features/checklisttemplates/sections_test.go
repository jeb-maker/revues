package checklisttemplates

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/store"
)

func TestParseTemplateItems_GroupedSections(t *testing.T) {
	form := url.Values{}
	form.Add("section_idx", "0")
	form.Add("section_idx", "1")
	form.Add("section_title", "Général")
	form.Add("section_title", "Sécurité")
	form.Add("item_section_idx", "0")
	form.Add("item_section_idx", "0")
	form.Add("item_section_idx", "1")
	form.Add("item_row_idx", "0")
	form.Add("item_row_idx", "1")
	form.Add("item_row_idx", "2")
	form.Add("item_label", "Point A")
	form.Add("item_label", "Point B")
	form.Add("item_label", "Point C")
	form.Add("item_help", "")
	form.Add("item_help", "Aide")
	form.Add("item_help", "")
	form.Add("item_required", "0")

	req := &http.Request{Method: http.MethodPost, Form: form}

	items, errMsg := parseTemplateItems(req)
	if errMsg != "" {
		t.Fatalf("parse error: %s", errMsg)
	}
	if len(items) != 3 {
		t.Fatalf("len(items) = %d, want 3", len(items))
	}
	if items[0].Section != "Général" || items[1].Section != "Général" {
		t.Fatalf("expected first points in Général, got %q and %q", items[0].Section, items[1].Section)
	}
	if items[2].Section != "Sécurité" {
		t.Fatalf("third section = %q, want Sécurité", items[2].Section)
	}
	if !items[0].Required || items[1].Required || items[2].Required {
		t.Fatal("only first item should be required")
	}
}

func TestItemsToEditorSections_GroupsConsecutive(t *testing.T) {
	sections := itemsToEditorSections([]store.TemplateItem{
		{Section: "Général", Label: "A"},
		{Section: "Général", Label: "B"},
		{Section: "Sécurité", Label: "C"},
	})
	if len(sections) != 2 {
		t.Fatalf("len(sections) = %d, want 2", len(sections))
	}
	if sections[0].Title != "Général" || len(sections[0].Items) != 2 {
		t.Fatalf("first section: title=%q items=%d", sections[0].Title, len(sections[0].Items))
	}
	if sections[1].Title != "Sécurité" || len(sections[1].Items) != 1 {
		t.Fatalf("second section: title=%q items=%d", sections[1].Title, len(sections[1].Items))
	}
}

func TestParseTemplateItems_LegacyFlatRows(t *testing.T) {
	form := url.Values{}
	form.Add("item_section", "Général")
	form.Add("item_label", "Point A")
	form.Add("item_help", "")
	form.Add("item_required", "0")

	req := &http.Request{Method: http.MethodPost, Form: form}
	items, errMsg := parseTemplateItems(req)
	if errMsg != "" {
		t.Fatalf("parse error: %s", errMsg)
	}
	if len(items) != 1 || items[0].Section != "Général" {
		t.Fatalf("item = %+v", items[0])
	}
}

func TestParseTemplateItems_RejectsLongLabel(t *testing.T) {
	form := url.Values{}
	form.Add("item_label", strings.Repeat("a", MaxTemplateItemLabelLen+1))
	form.Add("item_help", "")

	req := &http.Request{Method: http.MethodPost, Form: form}
	_, errMsg := parseTemplateItems(req)
	if errMsg == "" {
		t.Fatal("expected label length error")
	}
}

func TestParseTemplateItems_RejectsLongHelp(t *testing.T) {
	form := url.Values{}
	form.Add("item_label", "Point")
	form.Add("item_help", strings.Repeat("b", MaxTemplateItemHelpLen+1))

	req := &http.Request{Method: http.MethodPost, Form: form}
	_, errMsg := parseTemplateItems(req)
	if errMsg == "" {
		t.Fatal("expected help length error")
	}
}

func TestValidateTemplateItems_AcceptsLimits(t *testing.T) {
	items := []store.TemplateItemInput{
		{Label: strings.Repeat("a", MaxTemplateItemLabelLen), HelpText: strings.Repeat("b", MaxTemplateItemHelpLen)},
	}
	if msg := validateTemplateItems(items); msg != "" {
		t.Fatalf("unexpected error: %s", msg)
	}
}
