package items_test

import (
	"errors"
	"testing"

	"github.com/jeb-maker/revues/internal/items"
)

func TestValidateUpdate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		status  string
		comment string
		wantErr error
	}{
		{"ok without comment", items.StatusOK, "", nil},
		{"na without comment", items.StatusNA, "", nil},
		{"nok with comment", items.StatusNOK, "Problème détecté", nil},
		{"nok without comment", items.StatusNOK, "", items.ErrCommentRequired},
		{"nok whitespace comment", items.StatusNOK, "   ", items.ErrCommentRequired},
		{"invalid status", "bad", "x", items.ErrInvalidStatus},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := items.ValidateUpdate(tt.status, tt.comment)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateUpdate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
