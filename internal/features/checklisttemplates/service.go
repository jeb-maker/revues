package checklisttemplates

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/jeb-maker/revues/internal/features/subjects"
	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
)

// Sentinel errors for API mapping.
var (
	ErrNotFound   = store.ErrChecklistTemplateNotFound
	ErrValidation = errors.New("validation failed")
	ErrForbidden  = errors.New("forbidden")
	ErrImmutable  = store.ErrPublishedVersionImmutable
)

// Service orchestrates checklist template CRUD and versioning.
type Service struct {
	Store ChecklistTemplateStore
}

// Detail is a template with its latest (or selected) version, items and domains.
type Detail struct {
	Template *store.ChecklistTemplate
	Version  *store.TemplateVersion
	Items    []store.TemplateItem
	Domains  []string
}

// CanManageGlobal reports whether the user may create, edit or archive org
// checklist templates (any org member — catalog is org-scoped).
func CanManageGlobal(user *store.User, orgMember bool) bool {
	return subjects.CanCreateSubject(user, orgMember)
}

func (s *Service) orgMembership(ctx context.Context, userID int64) (orgAdmin bool, orgMember bool) {
	orgID, ok := orgctx.OrganizationID(ctx)
	if !ok || s.Store == nil {
		return false, false
	}
	role, member, err := s.Store.OrganizationMemberRole(ctx, orgID, userID)
	if err != nil || !member {
		return false, false
	}
	return role == store.OrgRoleOwner || role == store.OrgRoleAdmin, true
}

// ListIndex returns the global template catalog for the active organization.
func (s *Service) ListIndex(ctx context.Context, user *store.User, query string) ([]store.TemplateIndexRow, error) {
	if user == nil {
		return nil, ErrForbidden
	}
	orgAdmin, orgMember := s.orgMembership(ctx, user.ID)
	if !orgMember {
		return nil, ErrForbidden
	}
	rows, err := s.Store.ListTemplateIndex(ctx, user.ID, orgAdmin, strings.TrimSpace(query))
	if err != nil {
		return nil, fmt.Errorf("list template index: %w", err)
	}
	return rows, nil
}

// GetLatest loads a template with its latest published version and items.
func (s *Service) GetLatest(ctx context.Context, templateID int64) (*Detail, error) {
	tpl, err := s.Store.ChecklistTemplateByID(ctx, templateID)
	if err != nil {
		if errors.Is(err, store.ErrChecklistTemplateNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("load template: %w", err)
	}
	if tpl.ArchivedAt.Valid {
		return nil, ErrNotFound
	}

	version, err := s.Store.LatestTemplateVersion(ctx, tpl.ID)
	if err != nil {
		return nil, fmt.Errorf("latest version: %w", err)
	}
	items, err := s.Store.ListTemplateItems(ctx, version.ID)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	domains, err := s.Store.ListTemplateTags(ctx, tpl.ID)
	if err != nil {
		return nil, fmt.Errorf("list domains: %w", err)
	}
	return &Detail{Template: tpl, Version: version, Items: items, Domains: domains}, nil
}

// GetVersion loads a specific published version (immutable snapshot).
func (s *Service) GetVersion(ctx context.Context, templateID int64, versionNum int) (*Detail, error) {
	tpl, err := s.Store.ChecklistTemplateByID(ctx, templateID)
	if err != nil {
		if errors.Is(err, store.ErrChecklistTemplateNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("load template: %w", err)
	}
	if tpl.ArchivedAt.Valid {
		return nil, ErrNotFound
	}

	version, err := s.Store.TemplateVersionByNumber(ctx, templateID, versionNum)
	if err != nil {
		if errors.Is(err, store.ErrTemplateVersionNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("load version: %w", err)
	}
	items, err := s.Store.ListTemplateItems(ctx, version.ID)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	domains, err := s.Store.ListTemplateTags(ctx, tpl.ID)
	if err != nil {
		return nil, fmt.Errorf("list domains: %w", err)
	}
	return &Detail{Template: tpl, Version: version, Items: items, Domains: domains}, nil
}

// ListVersions returns all published versions for a template (newest first).
func (s *Service) ListVersions(ctx context.Context, templateID int64) ([]store.TemplateVersion, error) {
	tpl, err := s.Store.ChecklistTemplateByID(ctx, templateID)
	if err != nil {
		if errors.Is(err, store.ErrChecklistTemplateNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("load template: %w", err)
	}
	if tpl.ArchivedAt.Valid {
		return nil, ErrNotFound
	}
	versions, err := s.Store.ListTemplateVersions(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("list versions: %w", err)
	}
	return versions, nil
}

// CreateInput is the payload to create a template (version 1).
type CreateInput struct {
	Name    string
	Domains []string
	Items   []store.TemplateItemInput
}

// Create publishes a new template with version 1 (immutable snapshot).
func (s *Service) Create(ctx context.Context, user *store.User, in CreateInput) (*Detail, error) {
	if user == nil {
		return nil, ErrForbidden
	}
	_, orgMember := s.orgMembership(ctx, user.ID)
	if !CanManageGlobal(user, orgMember) {
		return nil, ErrForbidden
	}
	name, items, err := normalizeWrite(in.Name, in.Items)
	if err != nil {
		return nil, err
	}
	domains := store.NormalizeTags(in.Domains)

	tpl, version, err := s.Store.CreateChecklistTemplate(ctx, name, user.ID, domains, items)
	if err != nil {
		return nil, fmt.Errorf("create template: %w", err)
	}
	stored, err := s.Store.ListTemplateItems(ctx, version.ID)
	if err != nil {
		return nil, fmt.Errorf("list items after create: %w", err)
	}
	return &Detail{Template: tpl, Version: version, Items: stored, Domains: domains}, nil
}

// SaveInput updates metadata and either patches editorial fields on the latest
// version or publishes a new version when the item structure changes.
type SaveInput struct {
	Name    string
	Domains []string
	Items   []store.TemplateItemInput
}

// Save updates name/domains. Structural item changes publish a new version ;
// help_text-only (or metadata-only) changes stay on the latest version.
func (s *Service) Save(ctx context.Context, user *store.User, templateID int64, in SaveInput) (*Detail, error) {
	if user == nil {
		return nil, ErrForbidden
	}
	_, orgMember := s.orgMembership(ctx, user.ID)
	if !CanManageGlobal(user, orgMember) {
		return nil, ErrForbidden
	}
	name, items, err := normalizeWrite(in.Name, in.Items)
	if err != nil {
		return nil, err
	}
	domains := store.NormalizeTags(in.Domains)

	tpl, err := s.Store.ChecklistTemplateByID(ctx, templateID)
	if err != nil {
		if errors.Is(err, store.ErrChecklistTemplateNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("load template: %w", err)
	}
	if tpl.ArchivedAt.Valid {
		return nil, ErrNotFound
	}

	latest, err := s.Store.LatestTemplateVersion(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("latest version: %w", err)
	}
	currentItems, err := s.Store.ListTemplateItems(ctx, latest.ID)
	if err != nil {
		return nil, fmt.Errorf("list current items: %w", err)
	}

	if err = s.Store.UpdateChecklistTemplateName(ctx, templateID, name); err != nil {
		return nil, fmt.Errorf("update name: %w", err)
	}
	if err = s.Store.SetTemplateTags(ctx, templateID, domains); err != nil {
		return nil, fmt.Errorf("update domains: %w", err)
	}

	version := latest
	if IsStructuralItemChange(currentItems, items) {
		version, err = s.Store.CreateTemplateVersion(ctx, templateID, user.ID, items)
		if err != nil {
			return nil, fmt.Errorf("create version: %w", err)
		}
	} else if IsHelpTextChange(currentItems, items) {
		if err = s.Store.UpdateTemplateItemsHelpText(ctx, latest.ID, items); err != nil {
			return nil, fmt.Errorf("editorial help_text update: %w", err)
		}
	}

	tpl.Name = name
	stored, err := s.Store.ListTemplateItems(ctx, version.ID)
	if err != nil {
		return nil, fmt.Errorf("list items after save: %w", err)
	}
	return &Detail{Template: tpl, Version: version, Items: stored, Domains: domains}, nil
}

// IsStructuralItemChange reports whether next differs from current in count,
// order, label, section or required (help_text ignored).
func IsStructuralItemChange(current []store.TemplateItem, next []store.TemplateItemInput) bool {
	if len(current) != len(next) {
		return true
	}
	for i := range current {
		if current[i].Label != next[i].Label ||
			current[i].Section != next[i].Section ||
			current[i].Required != next[i].Required {
			return true
		}
	}
	return false
}

// IsHelpTextChange reports whether any help_text differs (same structure assumed).
func IsHelpTextChange(current []store.TemplateItem, next []store.TemplateItemInput) bool {
	if len(current) != len(next) {
		return true
	}
	for i := range current {
		if current[i].HelpText != next[i].HelpText {
			return true
		}
	}
	return false
}

// CreateVersion publishes a new version without changing name/domains.
func (s *Service) CreateVersion(ctx context.Context, user *store.User, templateID int64, items []store.TemplateItemInput) (*Detail, error) {
	if user == nil {
		return nil, ErrForbidden
	}
	_, orgMember := s.orgMembership(ctx, user.ID)
	if !CanManageGlobal(user, orgMember) {
		return nil, ErrForbidden
	}
	normalized, err := normalizeItems(items)
	if err != nil {
		return nil, err
	}

	tpl, err := s.Store.ChecklistTemplateByID(ctx, templateID)
	if err != nil {
		if errors.Is(err, store.ErrChecklistTemplateNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("load template: %w", err)
	}
	if tpl.ArchivedAt.Valid {
		return nil, ErrNotFound
	}

	version, err := s.Store.CreateTemplateVersion(ctx, templateID, user.ID, normalized)
	if err != nil {
		return nil, fmt.Errorf("create version: %w", err)
	}
	domains, err := s.Store.ListTemplateTags(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("list domains: %w", err)
	}
	stored, err := s.Store.ListTemplateItems(ctx, version.ID)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	return &Detail{Template: tpl, Version: version, Items: stored, Domains: domains}, nil
}

// Archive soft-deletes a template from the catalog.
func (s *Service) Archive(ctx context.Context, user *store.User, templateID int64) error {
	if user == nil {
		return ErrForbidden
	}
	_, orgMember := s.orgMembership(ctx, user.ID)
	if !CanManageGlobal(user, orgMember) {
		return ErrForbidden
	}
	if err := s.Store.ArchiveChecklistTemplate(ctx, templateID); err != nil {
		if errors.Is(err, store.ErrChecklistTemplateNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("archive template: %w", err)
	}
	return nil
}

func normalizeWrite(name string, items []store.TemplateItemInput) (string, []store.TemplateItemInput, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", nil, fmt.Errorf("%w: Le nom est obligatoire", ErrValidation)
	}
	if utf8.RuneCountInString(name) > MaxTemplateNameLen {
		return "", nil, fmt.Errorf("%w: Le nom ne peut pas dépasser %d caractères", ErrValidation, MaxTemplateNameLen)
	}
	normalized, err := normalizeItems(items)
	if err != nil {
		return "", nil, err
	}
	return name, normalized, nil
}

func normalizeItems(items []store.TemplateItemInput) ([]store.TemplateItemInput, error) {
	normalized := make([]store.TemplateItemInput, 0, len(items))
	for _, item := range items {
		label := strings.TrimSpace(item.Label)
		if label == "" {
			continue
		}
		normalized = append(normalized, store.TemplateItemInput{
			Section:  strings.TrimSpace(item.Section),
			Label:    label,
			HelpText: strings.TrimSpace(item.HelpText),
			Required: item.Required,
		})
	}
	if len(normalized) == 0 {
		return nil, fmt.Errorf("%w: Ajoutez au moins un point au modèle", ErrValidation)
	}
	if msg := validateTemplateItems(normalized); msg != "" {
		return nil, fmt.Errorf("%w: %s", ErrValidation, strings.TrimSuffix(msg, "."))
	}
	return normalized, nil
}
