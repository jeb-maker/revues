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

// RunUILabels holds org-configurable French labels for checklist runs (instances).
type RunUILabels struct {
	Nav         string // desktop nav / H1 / breadcrumbs (e.g. "Listes en cours")
	NavShort    string // mobile nav ≤36rem (e.g. "En cours"); equals Nav when unused
	Singular    string // lowercase phrase form: "revue", "liste"
	Plural      string // lowercase: "revues", "listes"
	Article     string // "une" | "un" for CTA
	NoneArticle string // "Aucune" | "Aucun" for empty states
}

// UILabels groups injectable UI label presets for the layout.
type UILabels struct {
	Subject SubjectUILabels
	Run     RunUILabels
}

// SubjectLabelPreset is one selectable org preset for subject wording.
type SubjectLabelPreset struct {
	Value    string
	Singular string
	Plural   string
}

// RunLabelPreset is one selectable org preset for run wording.
type RunLabelPreset struct {
	Value string
	Nav   string
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

// RunLabelPresets returns the allowed presets for the admin run-label select.
func RunLabelPresets() []RunLabelPreset {
	return []RunLabelPreset{
		{Value: store.UIRunLabelRevues, Nav: "Revues"},
		{Value: store.UIRunLabelListesEnCours, Nav: "Listes en cours"},
		{Value: store.UIRunLabelAudits, Nav: "Audits"},
		{Value: store.UIRunLabelChecklists, Nav: "Checklists"},
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

// RunLabelsForPreset resolves a stored run preset key to UI labels.
func RunLabelsForPreset(preset string) RunUILabels {
	normalized, err := store.NormalizeUIRunLabel(preset)
	if err != nil {
		normalized = store.UIRunLabelRevues
	}
	switch normalized {
	case store.UIRunLabelListesEnCours:
		return RunUILabels{
			Nav:         "Listes en cours",
			NavShort:    "En cours",
			Singular:    "liste",
			Plural:      "listes",
			Article:     "une",
			NoneArticle: "Aucune",
		}
	case store.UIRunLabelAudits:
		return RunUILabels{
			Nav:         "Audits",
			NavShort:    "Audits",
			Singular:    "audit",
			Plural:      "audits",
			Article:     "un",
			NoneArticle: "Aucun",
		}
	case store.UIRunLabelChecklists:
		return RunUILabels{
			Nav:         "Checklists",
			NavShort:    "Checklists",
			Singular:    "checklist",
			Plural:      "checklists",
			Article:     "une",
			NoneArticle: "Aucune",
		}
	default:
		return RunUILabels{
			Nav:         "Revues",
			NavShort:    "Revues",
			Singular:    "revue",
			Plural:      "revues",
			Article:     "une",
			NoneArticle: "Aucune",
		}
	}
}

// DefaultUILabels returns the v1 default subject + run label presets.
func DefaultUILabels() UILabels {
	return UILabels{
		Subject: SubjectLabelsForPreset(store.UISubjectLabelSujet),
		Run:     RunLabelsForPreset(store.UIRunLabelRevues),
	}
}

// LabelsFromOrganization resolves injectable labels from the active org presets.
func LabelsFromOrganization(org *store.Organization) UILabels {
	if org == nil {
		return DefaultUILabels()
	}
	return UILabels{
		Subject: SubjectLabelsForPreset(org.UISubjectLabel),
		Run:     RunLabelsForPreset(org.UIRunLabel),
	}
}

// EnsureLabels sets default UI labels when none were resolved from org settings.
func EnsureLabels(data *PageData) {
	if data.Labels.Subject.Singular == "" {
		data.Labels.Subject = DefaultUILabels().Subject
	}
	if data.Labels.Run.Nav == "" {
		data.Labels.Run = DefaultUILabels().Run
	}
}

// LaunchRunCTA returns the primary launch button label (e.g. "Lancer une revue").
func LaunchRunCTA(run RunUILabels) string {
	return fmt.Sprintf("Lancer %s %s", run.Article, run.Singular)
}

// LaunchActionTitle returns the tooltip for the launch-review header button.
func LaunchActionTitle(subject SubjectUILabels, run RunUILabels) string {
	return fmt.Sprintf("Lancer %s %s sur ce %s", run.Article, run.Singular, lowerFirst(subject.Singular))
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
