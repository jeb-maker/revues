package checklisttemplates

import (
	"github.com/jeb-maker/revues/internal/store"
)

// Store wraps a *store.Store to expose checklist-template-related queries from
// the checklisttemplates feature package. It embeds *store.Store so existing
// checklist template SQL methods (CreateChecklistTemplate,
// ChecklistTemplateByID, ListChecklistTemplates, etc.) are promoted and
// accessible through the checklisttemplates feature namespace.
//
// The underlying SQL stays in internal/store for now because it is shared by
// other features (runs, integrations) and the webhooks RunLoader interface.
// A dedicated issue will migrate the SQL into this package once the depending
// features are themselves extracted (mirrors the projects feature strategy).
type Store struct {
	*store.Store
}

// New returns a checklisttemplates Store backed by the given store.Store.
func New(s *store.Store) *Store {
	return &Store{Store: s}
}

// ChecklistTemplate re-exports store.ChecklistTemplate under the
// checklisttemplates namespace.
type ChecklistTemplate = store.ChecklistTemplate

// ChecklistTemplateSummary re-exports store.ChecklistTemplateSummary under the
// checklisttemplates namespace.
type ChecklistTemplateSummary = store.ChecklistTemplateSummary

// TemplateVersion re-exports store.TemplateVersion under the
// checklisttemplates namespace.
type TemplateVersion = store.TemplateVersion

// TemplateItem re-exports store.TemplateItem under the checklisttemplates
// namespace.
type TemplateItem = store.TemplateItem

// TemplateItemInput re-exports store.TemplateItemInput under the
// checklisttemplates namespace.
type TemplateItemInput = store.TemplateItemInput

// ErrChecklistTemplateNotFound re-exports store.ErrChecklistTemplateNotFound
// under the checklisttemplates namespace.
var ErrChecklistTemplateNotFound = store.ErrChecklistTemplateNotFound
