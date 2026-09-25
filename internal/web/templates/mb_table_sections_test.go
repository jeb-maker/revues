package templates_test

import (
	"testing"

	"github.com/jeb-maker/revues/internal/store"
	viewtemplates "github.com/jeb-maker/revues/internal/web/templates"
)

func TestSectionIDForTitleStable(t *testing.T) {
	a := viewtemplates.SectionIDForTitle("")
	b := viewtemplates.SectionIDForTitle("Sans section")
	c := viewtemplates.SectionIDForTitle("Ops")
	if a != b {
		t.Fatalf("empty and Sans section should share id: %q vs %q", a, b)
	}
	if a == c {
		t.Fatalf("Ops should differ from Sans section")
	}
	if viewtemplates.SectionIDForTitle("Ops") != c {
		t.Fatalf("id not stable across calls")
	}
}

func TestToMBTableSections(t *testing.T) {
	secs := viewtemplates.ToMBTableSections([]viewtemplates.RunItemSectionData{
		{
			ID:        viewtemplates.SectionIDForTitle("Ops"),
			Title:     "Ops",
			Items:     []store.RunItem{{}, {}},
			Total:     2,
			OKCount:   1,
			AllOKOrNA: false,
		},
		{
			ID:        viewtemplates.SectionIDForTitle("Sans section"),
			Title:     "Sans section",
			Total:     1,
			OKCount:   1,
			AllOKOrNA: true,
		},
	})
	if len(secs) != 2 {
		t.Fatalf("len=%d", len(secs))
	}
	if secs[0].Label != "Ops" || secs[0].Count != false || secs[0].Meta == "" {
		t.Fatalf("ops section unexpected: %+v", secs[0])
	}
	if secs[1].Label != "Point" || !secs[1].Collapsed {
		t.Fatalf("sans section unexpected: %+v", secs[1])
	}
}
