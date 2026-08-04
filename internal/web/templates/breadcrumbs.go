package templates

import (
	"strconv"
	"strings"
)

const (
	PathRevues         = "/revues"
	PathSubjects       = "/subjects"
	PathAdminOrg       = "/admin"
	PathAdminSubjects  = "/admin/subjects"
	PathAdminTeams     = "/admin/teams"
	PathTemplates      = "/modeles"
	PathAdmin          = "/admin/integrations"
	PathRevuesNouvelle = "/revues/nouvelle"
)

// ApplyPageMeta sets breadcrumbs and derives the document title from the last crumb.
func ApplyPageMeta(data PageData, crumbs []Breadcrumb) PageData {
	data.Breadcrumbs = crumbs
	if len(crumbs) > 0 {
		data.Title = crumbs[len(crumbs)-1].Label
	}
	return data
}

func crumb(label, url string) Breadcrumb {
	return Breadcrumb{Label: label, URL: url}
}

func current(label string) Breadcrumb {
	return Breadcrumb{Label: label}
}

func subjectPath(id int64) string {
	return "/subjects/" + strconv.FormatInt(id, 10)
}

func subjectModelesPath(id int64) string {
	return subjectPath(id) + "/modeles"
}

func runPath(id int64) string {
	return "/runs/" + strconv.FormatInt(id, 10)
}

// RunWizardPath is step 1 of the run launch wizard, optionally with a preselected template.
func RunWizardPath(templateID int64) string {
	if templateID <= 0 {
		return PathRevuesNouvelle
	}
	return PathRevuesNouvelle + "?template=" + strconv.FormatInt(templateID, 10)
}

// SubjectTemplatesForRunPath is the subject model picker when launching a review.
func SubjectTemplatesForRunPath(subjectID int64, templateID ...int64) string {
	path := subjectModelesPath(subjectID) + "?for_run=1"
	if len(templateID) > 0 && templateID[0] > 0 {
		path += "&template=" + strconv.FormatInt(templateID[0], 10)
	}
	return path
}

// SubjectModelesListPath builds the subject template list URL with optional wizard params.
func SubjectModelesListPath(subjectID int64, forRun bool, templateID int64) string {
	path := subjectModelesPath(subjectID)
	if !forRun && templateID <= 0 {
		return path
	}
	sep := "?"
	if forRun {
		path += sep + "for_run=1"
		sep = "&"
	}
	if templateID > 0 {
		path += sep + "template=" + strconv.FormatInt(templateID, 10)
	}
	return path
}

// BCRevues is the active runs index breadcrumb.
func BCRevues(run RunUILabels) []Breadcrumb {
	nav := run.Nav
	if nav == "" {
		nav = DefaultUILabels().Run.Nav
	}
	return []Breadcrumb{current(nav)}
}

// BCSubjects is the subjects index breadcrumb.
func BCSubjects(labels SubjectUILabels) []Breadcrumb {
	return []Breadcrumb{current(labels.Plural)}
}

// BCTasks is the my tasks index breadcrumb.
func BCTasks() []Breadcrumb {
	return []Breadcrumb{current("Mes tâches")}
}

// BCBugReport is the in-app bug report form breadcrumb.
func BCBugReport() []Breadcrumb {
	return []Breadcrumb{current("Signaler un problème")}
}

// TemplatesSectionLabel is the UI name for the templates area.
// listUI (mono-sujet) → Listes ; multi-sujet → Modèles.
func TemplatesSectionLabel(listUI bool) string {
	if listUI {
		return "Listes"
	}
	return "Modèles"
}

// BCTemplatesIndex is the global templates index breadcrumb.
func BCTemplatesIndex(simpleUI bool) []Breadcrumb {
	return []Breadcrumb{current(TemplatesSectionLabel(simpleUI))}
}

// BCLogin is the login page breadcrumb.
func BCLogin() []Breadcrumb {
	return []Breadcrumb{current("Connexion")}
}

// BCOrgNew is the organization creation form breadcrumb.
func BCOrgNew() []Breadcrumb {
	return []Breadcrumb{current("Nouvelle organisation")}
}

// BCOrgSelect is the organization selection screen breadcrumb.
func BCOrgSelect() []Breadcrumb {
	return []Breadcrumb{current("Choisir une organisation")}
}

// BCHome is the public landing page breadcrumb.
func BCHome() []Breadcrumb {
	return []Breadcrumb{current("Revues")}
}

func newSubjectCrumbLabel(labels SubjectUILabels) string {
	return "Nouveau " + strings.ToLower(labels.Singular)
}

// BCSubjectNew is the create subject form breadcrumb.
func BCSubjectNew(labels SubjectUILabels) []Breadcrumb {
	return []Breadcrumb{crumb(labels.Plural, PathSubjects), current(newSubjectCrumbLabel(labels))}
}

// BCSubjectShow is a subject detail breadcrumb; the subject name is the page H1.
func BCSubjectShow(name string, labels SubjectUILabels) []Breadcrumb {
	return []Breadcrumb{crumb(labels.Plural, PathSubjects), current(name)}
}

// BCSubjectEdit is the edit subject form breadcrumb.
func BCSubjectEdit(name string, id int64, labels SubjectUILabels) []Breadcrumb {
	return []Breadcrumb{crumb(labels.Plural, PathSubjects), crumb(name, subjectPath(id)), current("Modifier")}
}

// BCRunShow is a run detail breadcrumb.
func BCRunShow(title string, run RunUILabels) []Breadcrumb {
	return []Breadcrumb{crumb(runNav(run), PathRevues), current(title)}
}

// BCRunItemShow is a run item detail breadcrumb.
func BCRunItemShow(runTitle string, runID int64, itemLabel string, run RunUILabels) []Breadcrumb {
	return []Breadcrumb{crumb(runNav(run), PathRevues), crumb(runTitle, runPath(runID)), current(itemLabel)}
}

// BCRunWizardSubjects is run wizard step 1.
func BCRunWizardSubjects(run RunUILabels) []Breadcrumb {
	return []Breadcrumb{crumb(runNav(run), PathRevues), current(LaunchRunCTA(run))}
}

// BCRunWizardTemplates is run wizard step 2 (subject already chosen).
func BCRunWizardTemplates(subjectName string, subjectID int64, run RunUILabels) []Breadcrumb {
	return []Breadcrumb{
		crumb(runNav(run), PathRevues),
		crumb(subjectName, subjectPath(subjectID)),
		current("Choisir un modèle"),
	}
}

func runNav(run RunUILabels) string {
	if run.Nav == "" {
		return DefaultUILabels().Run.Nav
	}
	return run.Nav
}

// BCTemplatesNewWizard is the global new template wizard breadcrumb.
func BCTemplatesNewWizard(simpleUI bool) []Breadcrumb {
	sec := TemplatesSectionLabel(simpleUI)
	newLabel := "Nouveau"
	if simpleUI {
		newLabel = "Nouvelle"
	}
	return []Breadcrumb{crumb(sec, PathTemplates), current(newLabel)}
}

// BCTemplateGlobalShow is the global template detail breadcrumb.
func BCTemplateGlobalShow(name string, templateID int64, simpleUI bool) []Breadcrumb {
	return []Breadcrumb{
		crumb(TemplatesSectionLabel(simpleUI), PathTemplates),
		current(name),
	}
}

// BCTemplateGlobalEdit is the global template edit form breadcrumb.
func BCTemplateGlobalEdit(name string, templateID int64, simpleUI bool) []Breadcrumb {
	return []Breadcrumb{
		crumb(TemplatesSectionLabel(simpleUI), PathTemplates),
		crumb(name, "/modeles/"+strconv.FormatInt(templateID, 10)),
		current("Modifier"),
	}
}

// BCTemplateNotionImportGlobal is the global Notion import wizard breadcrumb.
func BCTemplateNotionImportGlobal(simpleUI bool) []Breadcrumb {
	return []Breadcrumb{crumb(TemplatesSectionLabel(simpleUI), PathTemplates), current("Importer depuis Notion")}
}

// BCSubjectTemplatesList is a subject's template list breadcrumb.
func BCSubjectTemplatesList(subjectName string, subjectID int64, labels SubjectUILabels) []Breadcrumb {
	return []Breadcrumb{crumb(labels.Plural, PathSubjects), crumb(subjectName, subjectPath(subjectID)), current("Modèles")}
}

// BCAdminOrgHub is the organisation admin landing page breadcrumb.
func BCAdminOrgHub() []Breadcrumb {
	return []Breadcrumb{current("Organisation")}
}

// BCAdminSubjectLabels is the org subject label preset breadcrumb.
func BCAdminSubjectLabels(labels SubjectUILabels) []Breadcrumb {
	return []Breadcrumb{crumb("Organisation", PathAdminOrg), current("Libellé " + LowerFirst(labels.Singular))}
}

// BCAdminLeadPolicies is the org lead-delegation policies breadcrumb.
func BCAdminLeadPolicies() []Breadcrumb {
	return []Breadcrumb{crumb("Organisation", PathAdminOrg), current("Politiques")}
}

// BCAdminSubjects is the org admin subjects list breadcrumb.
func BCAdminSubjects(labels SubjectUILabels) []Breadcrumb {
	return []Breadcrumb{crumb("Organisation", PathAdminOrg), current(labels.Plural)}
}

// BCAdminSubjectNew is the create subject form breadcrumb under Organisation.
func BCAdminSubjectNew(labels SubjectUILabels) []Breadcrumb {
	return []Breadcrumb{crumb("Organisation", PathAdminOrg), crumb(labels.Plural, PathAdminSubjects), current(newSubjectCrumbLabel(labels))}
}

// BCAdminSubjectEdit is the edit subject form breadcrumb under Organisation.
func BCAdminSubjectEdit(name string, id int64, labels SubjectUILabels) []Breadcrumb {
	return []Breadcrumb{
		crumb("Organisation", PathAdminOrg),
		crumb(labels.Plural, PathAdminSubjects),
		crumb(name, subjectPath(id)),
		current("Modifier"),
	}
}

func BCAdminUsers() []Breadcrumb {
	return []Breadcrumb{crumb("Organisation", PathAdminOrg), current("Emails autorisés")}
}

// BCAdminTeams is the admin teams list breadcrumb.
func BCAdminTeams() []Breadcrumb {
	return []Breadcrumb{crumb("Organisation", PathAdminOrg), current("Équipes")}
}

// BCAdminTeam is the admin team detail breadcrumb.
func BCAdminTeam(name string) []Breadcrumb {
	return []Breadcrumb{
		crumb("Organisation", PathAdminOrg),
		crumb("Équipes", PathAdminTeams),
		current(name),
	}
}

// BCAdminIntegrations is the admin integrations overview breadcrumb.
func BCAdminIntegrations() []Breadcrumb {
	return []Breadcrumb{crumb("Organisation", PathAdminOrg), current("Intégrations")}
}

// BCAdminSMTP is the admin SMTP settings breadcrumb.
func BCAdminSMTP() []Breadcrumb {
	return []Breadcrumb{crumb("Organisation", PathAdminOrg), crumb("Intégrations", PathAdmin), current("SMTP")}
}

// BCAdminWebhooks is the admin webhooks settings breadcrumb.
func BCAdminWebhooks() []Breadcrumb {
	return []Breadcrumb{crumb("Organisation", PathAdminOrg), crumb("Intégrations", PathAdmin), current("Webhooks")}
}

// BCAdminJira is the admin Jira settings breadcrumb.
func BCAdminJira() []Breadcrumb {
	return []Breadcrumb{crumb("Organisation", PathAdminOrg), crumb("Intégrations", PathAdmin), current("Jira")}
}

// BCAdminNotion is the admin Notion settings breadcrumb.
func BCAdminNotion() []Breadcrumb {
	return []Breadcrumb{crumb("Organisation", PathAdminOrg), crumb("Intégrations", PathAdmin), current("Notion")}
}

// BreadcrumbCurrent returns the label of the last breadcrumb.
func BreadcrumbCurrent(crumbs []Breadcrumb) string {
	if len(crumbs) == 0 {
		return ""
	}
	return crumbs[len(crumbs)-1].Label
}

// BreadcrumbAncestors returns parent crumbs only (excludes the current page).
// Empty when there is nothing useful to show above the H1.
func BreadcrumbAncestors(crumbs []Breadcrumb) []Breadcrumb {
	if len(crumbs) < 2 {
		return nil
	}
	return crumbs[:len(crumbs)-1]
}
