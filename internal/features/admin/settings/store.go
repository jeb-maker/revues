package settings

import (
	"context"

	"github.com/jeb-maker/revues/internal/store"
)

type SettingStore interface {
	GetSetting(ctx context.Context, key string) ([]byte, error)
	UpsertSetting(ctx context.Context, key string, value []byte) error
	DeleteSetting(ctx context.Context, key string) error
}

var ErrSettingNotFound = store.ErrSettingNotFound
