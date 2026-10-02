package checklisttemplates

import (
	"fmt"
	"unicode/utf8"

	"github.com/jeb-maker/revues/internal/store"
)

const (
	// MaxTemplateItemLabelLen is the max length for a checklist point label.
	MaxTemplateItemLabelLen = 120
	// MaxTemplateItemHelpLen is the max length for optional point help text.
	MaxTemplateItemHelpLen = 2000
	// MaxTemplateNameLen is the max length for a template display name.
	MaxTemplateNameLen = 200
)

func validateTemplateItemFields(label, help string) string {
	if utf8.RuneCountInString(label) > MaxTemplateItemLabelLen {
		return fmt.Sprintf("Le libellé ne peut pas dépasser %d caractères", MaxTemplateItemLabelLen)
	}
	if utf8.RuneCountInString(help) > MaxTemplateItemHelpLen {
		return fmt.Sprintf("Le texte d'aide ne peut pas dépasser %d caractères", MaxTemplateItemHelpLen)
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
