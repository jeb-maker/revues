package runs_test

import (
	"errors"
	"testing"

	"github.com/jeb-maker/revues/internal/features/runs"
)

func TestValidateUpdate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		status  string
		comment string
		wantErr error
	}{
		{"ok without comment", runs.StatusOK, "", nil},
		{"na without comment", runs.StatusNA, "", nil},
		{"nok with comment", runs.StatusNOK, "Problème détecté", nil},
		{"nok without comment", runs.StatusNOK, "", runs.ErrCommentRequired},
		{"nok whitespace comment", runs.StatusNOK, "   ", runs.ErrCommentRequired},
		{"invalid status", "bad", "x", runs.ErrInvalidStatus},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := runs.ValidateUpdate(tt.status, tt.comment)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateUpdate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
