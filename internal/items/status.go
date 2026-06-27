package items

import (
	"errors"
	"strings"
)

const (
	StatusPending = "pending"
	StatusOK      = "ok"
	StatusNOK     = "nok"
	StatusNA      = "na"
)

// ErrCommentRequired is returned when status nok has no comment.
var ErrCommentRequired = errors.New("comment required for nok status")

// ErrInvalidStatus is returned for unknown item statuses.
var ErrInvalidStatus = errors.New("invalid item status")

var validStatuses = map[string]struct{}{
	StatusPending: {},
	StatusOK:      {},
	StatusNOK:     {},
	StatusNA:      {},
}

// ValidStatus reports whether status is an allowed run item status.
func ValidStatus(status string) bool {
	_, ok := validStatuses[status]
	return ok
}

// ValidateUpdate checks status and comment business rules.
func ValidateUpdate(status, comment string) error {
	if !ValidStatus(status) {
		return ErrInvalidStatus
	}
	if status == StatusNOK && strings.TrimSpace(comment) == "" {
		return ErrCommentRequired
	}
	return nil
}
