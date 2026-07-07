package runs_test

import (
	"errors"
	"testing"

	"github.com/jeb-maker/revues/internal/features/runs"
)

func TestParseDueDate(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr error
	}{
		{"empty", "", "", nil},
		{"whitespace", "  ", "", nil},
		{"valid date", "2026-07-15", "2026-07-15T00:00:00Z", nil},
		{"invalid", "not-a-date", "", runs.ErrInvalidDueDate},
		{"partial", "2026-13-40", "", runs.ErrInvalidDueDate},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := runs.ParseDueDate(tt.raw)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseDueDate() err = %v", err)
			}
			if got != tt.want {
				t.Fatalf("got = %q, want %q", got, tt.want)
			}
		})
	}
}
