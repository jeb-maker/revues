package templates

import (
	"fmt"

	"github.com/jeb-maker/revues/internal/store"
)

// SubjectUILabels holds org-configurable French labels for the subject entity.
type SubjectUILabels struct {
	Singular string // e.g. "Sujet"
	Plural   string // e.g. "Sujets"
	Hint     string // short help text for forms
}

// UILabels groups injectable UI label presets for the layout.
type UILabels struct {
	Subject SubjectUILabels
}

// SubjectLabelPreset is one selectable org preset for subject wording.
type SubjectLabelPreset struct {
	Value    string
	Singular string
	Plural   string
}

// SubjectLabelPresets returns the allowed presets for the admin select.
func SubjectLabelPresets() []SubjectLabelPreset {
	return []SubjectLabelPreset{
		{Value: store.UISubjectLabelSujet, Singular: "Sujet", Plural: "Sujets"},
		{Value: store.UISubjectLabelCible, Singular: "Cible", Plural: "Cibles"},
		{Value: store.UISubjectLabelEntite, Singular: "Entité", Plural: "Entités"},
		{Value: store.UISubjectLabelAsset, Singular: "Actif", Plural: "Actifs"},
	}
}

// SubjectLabelsForPreset resolves a stored preset key to UI labels.
func SubjectLabelsForPreset(preset string) SubjectUILabels {
	normalized, err := store.NormalizeUISubjectLabel(preset)
	if err != nil {
		normalized = store.UISubjectLabelSujet
	}
	switch normalized {
	case store.UISubjectLabelCible:
		return SubjectUILabels{
			Singular: "Cible",
			Plural:   "Cibles",
			Hint:     "Ex. application, service, composant.",
		}
	case store.UISubjectLabelEntite:
		return SubjectUILabels{
			Singular: "Entité",
			Plural:   "Entités",
			Hint:     "Ex. équipe, business unit, département.",
		}
	case store.UISubjectLabelAsset:
		return SubjectUILabels{
			Singular: "Actif",
			Plural:   "Actifs",
			Hint:     "Ex. serveur, VM, équipement.",
		}
	default:
		return SubjectUILabels{
			Singular: "Sujet",
			Plural:   "Sujets",
			Hint:     "Ex. site, application, matériel.",
		}
	}
}

// DefaultUILabels returns the v1 default subject label preset (sujet).
func DefaultUILabels() UILabels {
	return UILabels{Subject: SubjectLabelsForPreset(store.UISubjectLabelSujet)}
}

// LabelsFromOrganization resolves injectable labels from the active org preset.
func LabelsFromOrganization(org *store.Organization) UILabels {
	if org == nil {
		return DefaultUILabels()
	}
	return UILabels{Subject: SubjectLabelsForPreset(org.UISubjectLabel)}
}

// EnsureLabels sets default UI labels when none were resolved from org settings.
func EnsureLabels(data *PageData) {
	if data.Labels.Subject.Singular == "" {
		data.Labels = DefaultUILabels()
	}
}

// LaunchActionTitle returns the tooltip for the launch-review header button.
func LaunchActionTitle(labels SubjectUILabels) string {
	return fmt.Sprintf("Lancer une revue sur ce %s", lowerFirst(labels.Singular))
}

func lowerFirst(s string) string {
	return LowerFirst(s)
}

// LowerFirst lowercases the first rune for French UI phrases.
func LowerFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	if len(r) == 1 {
		return string(r)
	}
	if r[0] >= 'A' && r[0] <= 'Z' {
		r[0] += 'a' - 'A'
	}
	return string(r)
}
