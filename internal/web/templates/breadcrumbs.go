package templates

import "strconv"

const (
	PathRevues    = "/revues"
	PathProjects  = "/projects"
	PathTasks     = "/mes-taches"
	PathTemplates = "/modeles"
	PathAdmin     = "/admin/integrations"
	PathRunsNew   = "/runs/new"
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

func projectPath(id int64) string {
	return "/projects/" + strconv.FormatInt(id, 10)
}

func projectTemplatesPath(id int64) string {
	return projectPath(id) + "/templates"
}

func templatePath(projectID, templateID int64) string {
	return projectTemplatesPath(projectID) + "/" + strconv.FormatInt(templateID, 10)
}

func runPath(id int64) string {
	return "/runs/" + strconv.FormatInt(id, 10)
}

// ProjectTemplatesForRunPath is the project template picker when launching a review.
func ProjectTemplatesForRunPath(projectID int64) string {
	return projectTemplatesPath(projectID) + "?for_run=1"
}

// BCRevues is the active runs index breadcrumb.
func BCRevues() []Breadcrumb {
	return []Breadcrumb{current("Revues")}
}

// BCProjects is the projects index breadcrumb.
func BCProjects() []Breadcrumb {
	return []Breadcrumb{current("Projets")}
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

// BCProjectNew is the create project form breadcrumb.
func BCProjectNew() []Breadcrumb {
	return []Breadcrumb{crumb("Projets", PathProjects), current("Nouveau")}
}

// BCProjectShow is a project detail breadcrumb.
func BCProjectShow(name string) []Breadcrumb {
	return []Breadcrumb{crumb("Projets", PathProjects), current(name)}
}

// BCProjectEdit is the edit project form breadcrumb.
func BCProjectEdit(name string, id int64) []Breadcrumb {
	return []Breadcrumb{crumb("Projets", PathProjects), crumb(name, projectPath(id)), current("Modifier")}
}

// BCRunShow is a run detail breadcrumb.
func BCRunShow(title string) []Breadcrumb {
	return []Breadcrumb{crumb("Revues", PathRevues), current(title)}
}

// BCRunItemShow is a run item detail breadcrumb.
func BCRunItemShow(runTitle string, runID int64, itemLabel string) []Breadcrumb {
	return []Breadcrumb{crumb("Revues", PathRevues), crumb(runTitle, runPath(runID)), current(itemLabel)}
}

// BCRunWizardProjects is run wizard step 1.
func BCRunWizardProjects() []Breadcrumb {
	return []Breadcrumb{crumb("Revues", PathRevues), current("Lancer une revue")}
}

// BCRunWizardTemplates is run wizard step 2 (project already chosen).
func BCRunWizardTemplates(projectName string, projectID int64) []Breadcrumb {
	return []Breadcrumb{
		crumb("Revues", PathRevues),
		crumb(projectName, projectPath(projectID)),
		current("Lancer"),
	}
}

// BCRunWizardLaunch is run wizard step 3 (confirm title and launch).
func BCRunWizardLaunch(projectName string, projectID int64, templateName string, version, itemCount int) []Breadcrumb {
	return []Breadcrumb{
		crumb("Revues", PathRevues),
		crumb(projectName, projectPath(projectID)),
		crumb("Lancer", ProjectTemplatesForRunPath(projectID)),
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

// BCProjectTemplatesList is a project's template list breadcrumb.
func BCProjectTemplatesList(projectName string, projectID int64) []Breadcrumb {
	return []Breadcrumb{crumb("Projets", PathProjects), crumb(projectName, projectPath(projectID)), current("Modèles")}
}

// BCTemplateNew is the create template form breadcrumb.
func BCTemplateNew(projectName string, projectID int64) []Breadcrumb {
	return []Breadcrumb{
		crumb("Projets", PathProjects),
		crumb(projectName, projectPath(projectID)),
		crumb("Modèles", projectTemplatesPath(projectID)),
		current("Nouveau"),
	}
}

// BCTemplateShow is a template detail breadcrumb.
func BCTemplateShow(projectName string, projectID int64, templateName string) []Breadcrumb {
	return []Breadcrumb{
		crumb("Projets", PathProjects),
		crumb(projectName, projectPath(projectID)),
		crumb("Modèles", projectTemplatesPath(projectID)),
		current(templateName),
	}
}

// BCTemplateEdit is the edit template form breadcrumb.
func BCTemplateEdit(projectName string, projectID int64, templateName string, templateID int64) []Breadcrumb {
	return []Breadcrumb{
		crumb("Projets", PathProjects),
		crumb(projectName, projectPath(projectID)),
		crumb("Modèles", projectTemplatesPath(projectID)),
		crumb(templateName, templatePath(projectID, templateID)),
		current("Modifier"),
	}
}

// BCTemplateNotionImport is the Notion import wizard breadcrumb.
func BCTemplateNotionImport(projectName string, projectID int64) []Breadcrumb {
	return []Breadcrumb{
		crumb("Projets", PathProjects),
		crumb(projectName, projectPath(projectID)),
		crumb("Modèles", projectTemplatesPath(projectID)),
		current("Importer depuis Notion"),
	}
}

// BCAdminUsers is the admin users page breadcrumb.
func BCAdminUsers() []Breadcrumb {
	return []Breadcrumb{crumb("Admin", PathAdmin), current("Emails autorisés")}
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
