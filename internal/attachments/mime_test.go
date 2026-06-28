package attachments_test

import (
	"testing"

	"github.com/jeb-maker/revues/internal/attachments"
)

func TestIsImageMime(t *testing.T) {
	tests := []struct {
		mime string
		want bool
	}{
		{"image/jpeg", true},
		{"image/png", true},
		{"image/webp", true},
		{"IMAGE/JPEG", true},
		{"application/pdf", false},
		{"text/plain", false},
	}
	for _, tt := range tests {
		if got := attachments.IsImageMime(tt.mime); got != tt.want {
			t.Errorf("IsImageMime(%q) = %v, want %v", tt.mime, got, tt.want)
		}
	}
}
