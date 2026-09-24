package templates_test

import (
	"testing"

	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/templates"
)

func TestFormatAccessSource(t *testing.T) {
	teamNames := map[int64]string{7: "Squad A"}
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{name: "direct", source: store.AccessSourceDirect, want: "accès direct"},
		{name: "org admin", source: store.AccessSourceOrgAdmin, want: "admin organisation"},
		{name: "global admin", source: store.AccessSourceGlobalAdmin, want: "admin global"},
		{name: "org member legacy", source: store.AccessSourceOrgMemberLegacy, want: "membre organisation"},
		{name: "team named", source: "team:7", want: "via équipe Squad A"},
		{name: "team unknown", source: "team:99", want: "via équipe"},
		{name: "passthrough", source: "other", want: "other"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := templates.FormatAccessSource(tt.source, teamNames)
			if got != tt.want {
				t.Fatalf("FormatAccessSource(%q) = %q, want %q", tt.source, got, tt.want)
			}
		})
	}
}
