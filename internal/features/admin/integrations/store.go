package integrations

import (
	"github.com/jeb-maker/revues/internal/features/admin/settings"
	"github.com/jeb-maker/revues/internal/integrations/jira"
	"github.com/jeb-maker/revues/internal/integrations/notion"
)

type AdminStore interface {
	settings.SettingStore
	jira.ConfigStore
	notion.ConfigStore
}
