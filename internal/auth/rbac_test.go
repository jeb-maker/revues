package auth_test

import (
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
)

func TestHasMinRole(t *testing.T) {
	t.Parallel()

	tests := []struct {
		user string
		min  string
		want bool
	}{
		{"admin", "admin", true},
		{"admin", "editor", true},
		{"admin", "reader", true},
		{"editor", "admin", false},
		{"editor", "editor", true},
		{"editor", "reader", true},
		{"reader", "editor", false},
		{"reader", "reader", true},
		{"invalid", "reader", false},
		{"reader", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.user+"_needs_"+tt.min, func(t *testing.T) {
			t.Parallel()
			if got := auth.HasMinRole(tt.user, tt.min); got != tt.want {
				t.Errorf("HasMinRole(%q, %q) = %v, want %v", tt.user, tt.min, got, tt.want)
			}
		})
	}
}
