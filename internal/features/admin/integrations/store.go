package integrations

import (
	"github.com/jeb-maker/revues/internal/features/admin/settings"
	"github.com/jeb-maker/revues/internal/integrations/jira"
)

type AdminStore interface {
	settings.SettingStore
	jira.ConfigStore
}
