package templates

import (
	"fmt"
	"html/template"
	"io/fs"

	"github.com/jeb-maker/revues/internal/store"
	webassets "github.com/jeb-maker/revues/web"
)

// PageData is shared view data for HTML pages.
type PageData struct {
	Title      string
	User       *store.User
	CSRFToken  string
	LoginError string
}

// AdminUsersData is view data for the whitelist admin screen.
type AdminUsersData struct {
	PageData
	Emails  []store.AllowedEmail
	Message string
	Error   string
}

// ProjectsListData is view data for the project index.
type ProjectsListData struct {
	PageData
	Projects  []store.Project
	CanCreate bool
	Message   string
	Error     string
}

// ProjectFormData is view data for create/edit project forms.
type ProjectFormData struct {
	PageData
	Project    *store.Project
	FormAction string
	Error      string
}

// ProjectShowData is view data for project detail.
type ProjectShowData struct {
	PageData
	Project          *store.Project
	Members          []store.ProjectMember
	Runs             []store.ChecklistRun
	MemberRole       string
	CanManage        bool
	CanManageMembers bool
	CanLaunch        bool
	Message          string
	Error            string
}

// TemplateEditorRow is one editable checklist point in the form.
type TemplateEditorRow struct {
	Section  string
	Label    string
	HelpText string
	Required bool
}

// ChecklistTemplatesListData is view data for template index on a project.
type ChecklistTemplatesListData struct {
	PageData
	Project    *store.Project
	Templates  []store.ChecklistTemplateSummary
	MemberRole string
	CanManage  bool
	Message    string
	Error      string
}

// ChecklistTemplateFormData is view data for create/edit template forms.
type ChecklistTemplateFormData struct {
	PageData
	Project    *store.Project
	Template   *store.ChecklistTemplate
	Version    *store.TemplateVersion
	Name       string
	Rows       []TemplateEditorRow
	FormAction string
	Error      string
}

// ChecklistTemplateShowData is view data for template detail.
type ChecklistTemplateShowData struct {
	PageData
	Project    *store.Project
	Template   *store.ChecklistTemplate
	Version    *store.TemplateVersion
	Items      []store.TemplateItem
	MemberRole string
	CanManage  bool
	Message    string
	Error      string
}

// RunWizardProjectsData is view data for run wizard step 1.
type RunWizardProjectsData struct {
	PageData
	Projects []store.Project
	Step     int
	Message  string
	Error    string
}

// RunWizardTemplatesData is view data for run wizard step 2.
type RunWizardTemplatesData struct {
	PageData
	Project    *store.Project
	Templates  []store.ChecklistTemplateSummary
	Step       int
	MemberRole string
	CanLaunch  bool
	Message    string
	Error      string
}

// RunWizardLaunchData is view data for run wizard step 3.
type RunWizardLaunchData struct {
	PageData
	Project    *store.Project
	Template   *store.ChecklistTemplate
	Version    *store.TemplateVersion
	ItemCount  int
	Title      string
	FormAction string
	Step       int
	MemberRole string
	CanLaunch  bool
	Error      string
}

// RunShowData is view data for run detail.
type RunShowData struct {
	PageData
	Project       *store.Project
	Run           *store.ChecklistRun
	Items         []store.RunItem
	NokItems      []store.RunItem
	Members       []store.ProjectMember
	TemplateName  string
	VersionNum    int
	MemberRole    string
	CanLaunch     bool
	CanCheck      bool
	CanAssign     bool
	CanComplete   bool
	ClosingNote   string
	Message       string
	ItemError     string
	AssignError   string
	CompleteError string
	Error         string
}

// MyTasksData is view data for assigned tasks list.
type MyTasksData struct {
	PageData
	Tasks           []store.AssignedRunItemSummary
	Projects        []store.Project
	FilterProjectID int64
	FilterStatus    string
}

// RunItemShowData is view data for run item detail with audit history.
type RunItemShowData struct {
	PageData
	Project    *store.Project
	Run        *store.ChecklistRun
	Item       *store.RunItem
	Events     []store.RunItemEvent
	MemberRole string
	CanCheck   bool
}

// Parse loads layout and page templates from the embedded filesystem.
func Parse() (*template.Template, error) {
	root, err := fs.Sub(webassets.Templates, "templates")
	if err != nil {
		return nil, fmt.Errorf("templates root: %w", err)
	}

	tpl := template.New("").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	})

	tpl, err = tpl.ParseFS(root,
		"layouts/base.html",
		"pages/*.html",
	)
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	return tpl, nil
}
