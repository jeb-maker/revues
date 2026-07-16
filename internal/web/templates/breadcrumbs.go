package templates

import "strconv"

const (
	PathRevues         = "/revues"
	PathSubjects       = "/subjects"
	PathAdminOrg       = "/admin"
	PathAdminSubjects  = "/admin/subjects"
	PathProjects       = "/subjects" // deprecated alias
	PathTasks          = "/mes-taches"
	PathTemplates      = "/modeles"
	PathAdmin          = "/admin/integrations"
	PathRevuesNouvelle = "/revues/nouvelle"
	PathRunsNew        = "/revues/nouvelle" // deprecated alias
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

func templatePath(subjectID, templateID int64) string {
	return subjectModelesPath(subjectID) + "/" + strconv.FormatInt(templateID, 10)
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

// SubjectModelesClearPath drops search filters while keeping launch context.
func SubjectModelesClearPath(subjectID int64, forRun bool, templateID int64) string {
	if forRun {
		return SubjectTemplatesForRunPath(subjectID, templateID)
	}
	if templateID > 0 {
		return subjectModelesPath(subjectID) + "?template=" + strconv.FormatInt(templateID, 10)
	}
	return subjectModelesPath(subjectID)
}

// ProjectTemplatesForRunPath is a deprecated alias for SubjectTemplatesForRunPath.
func ProjectTemplatesForRunPath(projectID int64) string {
	return SubjectTemplatesForRunPath(projectID)
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
func BCRevues() []Breadcrumb {
	return []Breadcrumb{current("Revues")}
}

// BCSubjects is the subjects index breadcrumb.
func BCSubjects(labels SubjectUILabels) []Breadcrumb {
	return []Breadcrumb{current(labels.Plural)}
}

// BCProjects is a deprecated alias for BCSubjects.
func BCProjects() []Breadcrumb {
	return BCSubjects(DefaultUILabels().Subject)
}

// BCTasks is the my tasks index breadcrumb.
func BCTasks() []Breadcrumb {
	return []Breadcrumb{current("Mes tâches")}
}

// BCTemplatesIndex is the global templates index breadcrumb.
func BCTemplatesIndex() []Breadcrumb {
	return []Breadcrumb{current("Modèles")}
}

// BCAdmin is the admin section root breadcrumb.
func BCAdmin() []Breadcrumb {
	return []Breadcrumb{current("Admin")}
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

// BCSubjectNew is the create subject form breadcrumb.
func BCSubjectNew(labels SubjectUILabels) []Breadcrumb {
	return []Breadcrumb{crumb(labels.Plural, PathSubjects), current("Nouveau")}
}

// BCProjectNew is a deprecated alias for BCSubjectNew.
func BCProjectNew() []Breadcrumb {
	return BCSubjectNew(DefaultUILabels().Subject)
}

// BCSubjectShow is a subject detail breadcrumb; the subject name is the page H1.
func BCSubjectShow(name string, labels SubjectUILabels) []Breadcrumb {
	return []Breadcrumb{crumb(labels.Plural, PathSubjects), current(name)}
}

// BCProjectShow is a deprecated alias for BCSubjectShow.
func BCProjectShow(name string) []Breadcrumb {
	return BCSubjectShow(name, DefaultUILabels().Subject)
}

// BCSubjectEdit is the edit subject form breadcrumb.
func BCSubjectEdit(name string, id int64, labels SubjectUILabels) []Breadcrumb {
	return []Breadcrumb{crumb(labels.Plural, PathSubjects), crumb(name, subjectPath(id)), current("Modifier")}
}

// BCProjectEdit is a deprecated alias for BCSubjectEdit.
func BCProjectEdit(name string, id int64) []Breadcrumb {
	return BCSubjectEdit(name, id, DefaultUILabels().Subject)
}

// BCRunShow is a run detail breadcrumb.
func BCRunShow(title string) []Breadcrumb {
	return []Breadcrumb{crumb("Revues", PathRevues), current(title)}
}

// BCRunItemShow is a run item detail breadcrumb.
func BCRunItemShow(runTitle string, runID int64, itemLabel string) []Breadcrumb {
	return []Breadcrumb{crumb("Revues", PathRevues), crumb(runTitle, runPath(runID)), current(itemLabel)}
}

// BCRunWizardSubjects is run wizard step 1.
func BCRunWizardSubjects() []Breadcrumb {
	return []Breadcrumb{crumb("Revues", PathRevues), current("Lancer une revue")}
}

// BCRunWizardProjects is a deprecated alias for BCRunWizardSubjects.
func BCRunWizardProjects() []Breadcrumb {
	return BCRunWizardSubjects()
}

// BCRunWizardTemplates is run wizard step 2 (subject already chosen).
func BCRunWizardTemplates(subjectName string, subjectID int64) []Breadcrumb {
	return []Breadcrumb{
		crumb("Revues", PathRevues),
		crumb(subjectName, subjectPath(subjectID)),
		current("Choisir un modèle"),
	}
}

// BCRunWizardLaunch is run wizard step 3 (confirm title and launch).
func BCRunWizardLaunch(subjectName string, subjectID int64, templateName string, version, itemCount int) []Breadcrumb {
	return []Breadcrumb{
		crumb("Revues", PathRevues),
		crumb(subjectName, subjectPath(subjectID)),
		crumb("Choisir un modèle", SubjectTemplatesForRunPath(subjectID)),
		current(runLaunchTemplateLabel(templateName, version, itemCount)),
	}
}

func runLaunchTemplateLabel(name string, version, itemCount int) string {
	suffix := " points de contrôle"
	if itemCount == 1 {
		suffix = " point de contrôle"
	}
	return name + " · v" + strconv.Itoa(version) + " · " + strconv.Itoa(itemCount) + suffix
}

// BCTemplatesNewWizard is the global new template wizard breadcrumb.
func BCTemplatesNewWizard() []Breadcrumb {
	return []Breadcrumb{crumb("Modèles", PathTemplates), current("Nouveau")}
}

// BCTemplateGlobalShow is the global template detail breadcrumb.
func BCTemplateGlobalShow(name string, templateID int64) []Breadcrumb {
	return []Breadcrumb{
		crumb("Modèles", PathTemplates),
		current(name),
	}
}

// BCTemplateGlobalEdit is the global template edit form breadcrumb.
func BCTemplateGlobalEdit(name string, templateID int64) []Breadcrumb {
	return []Breadcrumb{
		crumb("Modèles", PathTemplates),
		crumb(name, "/modeles/"+strconv.FormatInt(templateID, 10)),
		current("Modifier"),
	}
}

// BCTemplateNotionImportGlobal is the global Notion import wizard breadcrumb.
func BCTemplateNotionImportGlobal() []Breadcrumb {
	return []Breadcrumb{crumb("Modèles", PathTemplates), current("Importer depuis Notion")}
}

// BCSubjectTemplatesList is a subject's template list breadcrumb.
func BCSubjectTemplatesList(subjectName string, subjectID int64, labels SubjectUILabels) []Breadcrumb {
	return []Breadcrumb{crumb(labels.Plural, PathSubjects), crumb(subjectName, subjectPath(subjectID)), current("Modèles")}
}

// BCProjectTemplatesList is a deprecated alias for BCSubjectTemplatesList.
func BCProjectTemplatesList(projectName string, projectID int64) []Breadcrumb {
	return BCSubjectTemplatesList(projectName, projectID, DefaultUILabels().Subject)
}

// BCTemplateNew is the create template form breadcrumb.
func BCTemplateNew(subjectName string, subjectID int64, labels SubjectUILabels) []Breadcrumb {
	return []Breadcrumb{
		crumb(labels.Plural, PathSubjects),
		crumb(subjectName, subjectPath(subjectID)),
		crumb("Modèles", subjectModelesPath(subjectID)),
		current("Nouveau"),
	}
}

// BCTemplateShow is a template detail breadcrumb.
func BCTemplateShow(subjectName string, subjectID int64, templateName string, labels SubjectUILabels) []Breadcrumb {
	return []Breadcrumb{
		crumb(labels.Plural, PathSubjects),
		crumb(subjectName, subjectPath(subjectID)),
		crumb("Modèles", subjectModelesPath(subjectID)),
		current(templateName),
	}
}

// BCTemplateEdit is the edit template form breadcrumb.
func BCTemplateEdit(subjectName string, subjectID int64, templateName string, templateID int64, labels SubjectUILabels) []Breadcrumb {
	return []Breadcrumb{
		crumb(labels.Plural, PathSubjects),
		crumb(subjectName, subjectPath(subjectID)),
		crumb("Modèles", subjectModelesPath(subjectID)),
		crumb(templateName, templatePath(subjectID, templateID)),
		current("Modifier"),
	}
}

// BCTemplateNotionImport is the Notion import wizard breadcrumb.
func BCTemplateNotionImport(subjectName string, subjectID int64, labels SubjectUILabels) []Breadcrumb {
	return []Breadcrumb{
		crumb(labels.Plural, PathSubjects),
		crumb(subjectName, subjectPath(subjectID)),
		crumb("Modèles", subjectModelesPath(subjectID)),
		current("Importer depuis Notion"),
	}
}

// BCAdminOrgHub is the organisation admin landing page breadcrumb.
func BCAdminOrgHub() []Breadcrumb {
	return []Breadcrumb{current("Organisation")}
}

// BCAdminSubjectLabels is the org subject label preset breadcrumb.
func BCAdminSubjectLabels(labels SubjectUILabels) []Breadcrumb {
	return []Breadcrumb{crumb("Organisation", PathAdminOrg), current("Libellé " + LowerFirst(labels.Singular))}
}

// BCAdminSubjects is the org admin subjects list breadcrumb.
func BCAdminSubjects(labels SubjectUILabels) []Breadcrumb {
	return []Breadcrumb{crumb("Organisation", PathAdminOrg), current(labels.Plural)}
}

// BCAdminSubjectNew is the create subject form breadcrumb under Organisation.
func BCAdminSubjectNew(labels SubjectUILabels) []Breadcrumb {
	return []Breadcrumb{crumb("Organisation", PathAdminOrg), crumb(labels.Plural, PathAdminSubjects), current("Nouveau")}
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

// BCAdminIntegrations is the admin integrations overview breadcrumb.
func BCAdminIntegrations() []Breadcrumb {
	return []Breadcrumb{crumb("Admin", PathAdmin), current("Intégrations")}
}

// BCAdminSMTP is the admin SMTP settings breadcrumb.
func BCAdminSMTP() []Breadcrumb {
	return []Breadcrumb{crumb("Admin", PathAdmin), current("SMTP")}
}

// BCAdminWebhooks is the admin webhooks settings breadcrumb.
func BCAdminWebhooks() []Breadcrumb {
	return []Breadcrumb{crumb("Admin", PathAdmin), current("Webhooks")}
}

// BCAdminJira is the admin Jira settings breadcrumb.
func BCAdminJira() []Breadcrumb {
	return []Breadcrumb{crumb("Admin", PathAdmin), current("Jira")}
}

// BCAdminNotion is the admin Notion settings breadcrumb.
func BCAdminNotion() []Breadcrumb {
	return []Breadcrumb{crumb("Admin", PathAdmin), current("Notion")}
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
