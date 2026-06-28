package attachments

import "strings"

// IsImageMime reports whether mime is a supported inline image type.
func IsImageMime(mime string) bool {
	switch strings.ToLower(strings.TrimSpace(mime)) {
	case "image/jpeg", "image/png", "image/webp":
		return true
	default:
		return false
	}
}
