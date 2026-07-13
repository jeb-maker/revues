package checklisttemplates

import (
	"fmt"
	"unicode/utf8"

	"github.com/jeb-maker/revues/internal/store"
	viewtemplates "github.com/jeb-maker/revues/internal/web/templates"
)

const (
	// MaxTemplateItemLabelLen is the max length for a checklist point label (list display).
	MaxTemplateItemLabelLen = 120
	// MaxTemplateItemHelpLen is the max length for optional point help text.
	MaxTemplateItemHelpLen = 2000
)

func validateTemplateItemFields(label, help string) string {
	if utf8.RuneCountInString(label) > MaxTemplateItemLabelLen {
		return fmt.Sprintf("Le libellé ne peut pas dépasser %d caractères.", MaxTemplateItemLabelLen)
	}
	if utf8.RuneCountInString(help) > MaxTemplateItemHelpLen {
		return fmt.Sprintf("Le texte d'aide ne peut pas dépasser %d caractères.", MaxTemplateItemHelpLen)
	}
	return ""
}

func validateTemplateItems(items []store.TemplateItemInput) string {
	for _, item := range items {
		if msg := validateTemplateItemFields(item.Label, item.HelpText); msg != "" {
			return msg
		}
	}
	return ""
}

func applyTemplateFormLimits(data *viewtemplates.ChecklistTemplateFormData) {
	data.MaxItemLabelLen = MaxTemplateItemLabelLen
	data.MaxItemHelpLen = MaxTemplateItemHelpLen
}
