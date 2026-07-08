package notion

import (
	"context"

	"github.com/jeb-maker/revues/internal/store"
)

type ConfigStore interface {
	GetIntegrationByType(ctx context.Context, integrationType string) (*store.Integration, error)
	UpsertIntegrationByType(ctx context.Context, integrationType string, enabled bool, configEncrypted []byte) error
}

var ErrIntegrationNotFound = store.ErrIntegrationNotFound
