package middleware

import (
	"context"
	"log/slog"

	"github.com/jeb-maker/revues/internal/features/admin/settings"
	"github.com/jeb-maker/revues/internal/integrations/jira"
	"github.com/jeb-maker/revues/internal/integrations/notion"
	"github.com/jeb-maker/revues/internal/store"
)

// resolveCapabilityCaps fills P3 org-level flags (integration configured).
// Independent of SimpleUI — configured integrations stay visible in particulier mode.
// HasEvidence is page-scoped (sealed run hash) and is set by handlers, not here.
func resolveCapabilityCaps(ctx context.Context, st *store.Store, encryptionKey []byte) (hasJira, hasNotion, hasWebhooks bool) {
	if st == nil || len(encryptionKey) == 0 {
		return false, false, false
	}
	if _, ok := OrganizationFromContext(ctx); !ok {
		return false, false, false
	}

	jiraOK, err := (&jira.LinkService{Store: st, EncryptionKey: encryptionKey}).Configured(ctx)
	if err != nil {
		slog.Debug("resolve capability jira", "err", err)
	} else {
		hasJira = jiraOK
	}

	cfg, ok, err := (&notion.Service{Store: st, EncryptionKey: encryptionKey}).Load(ctx)
	if err != nil {
		slog.Debug("resolve capability notion", "err", err)
	} else {
		hasNotion = ok && cfg.Configured()
	}

	wh, ok, err := (&settings.SettingsService{Store: st, EncryptionKey: encryptionKey}).LoadWebhooks(ctx)
	if err != nil {
		slog.Debug("resolve capability webhooks", "err", err)
	} else {
		hasWebhooks = ok && wh.Enabled()
	}
	return hasJira, hasNotion, hasWebhooks
}
