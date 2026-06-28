package templates

import (
	"database/sql"
	"fmt"
	"html/template"
	"io/fs"
	"time"

	"github.com/jeb-maker/revues/internal/store"
	webassets "github.com/jeb-maker/revues/web"
)

// PageData is shared view data for HTML pages.
type PageData struct {
	Title      string
	User       *store.User
	CSRFToken  string
	LoginError string
	ActiveTab  string
}

// AdminUsersData is view data for the whitelist admin screen.
type AdminUsersData struct {
	PageData
	Emails  []store.AllowedEmail
	Message string
	Error   string
}

type AdminNotionData struct {
	PageData
	WorkspaceName     string
	DefaultDatabaseID string
	HasAPIToken       bool
	Configured        bool
	CanEncrypt        bool
	Message           string
	Error             string
}

type AdminJiraData struct {
	PageData
	InstanceType string
	BaseURL      string
	Email        string
	ProjectKey   string
	IssueType    string
	HasAPIToken  bool
	HasPAT       bool
	Configured   bool
	CanEncrypt   bool
	Message      string
	Error        string
}

// AdminSMTPData is view data for the SMTP admin screen.
type AdminSMTPData struct {
	PageData
	Host          string
	Port          int
	TLS           bool
	Username      string
	From          string
	HasPassword   bool
	Configured    bool
	CanEncrypt    bool
	TestRecipient string
	Message       string
	Error         string
}

type AdminIntegrationRow struct {
	Name        string
	Description string
	Enabled     bool
	ConfigPath  string
}

type AdminIntegrationsData struct {
	PageData
	Integrations []AdminIntegrationRow
	Error        string
}

// AdminWebhooksData is view data for the webhooks admin screen.
type AdminWebhooksData struct {
	PageData
	URLsText        string
	HasSecret       bool
	ReviewCompleted bool
	ReviewItemNOK   bool
	Configured      bool
	CanEncrypt      bool
	Message         string
	Error           string
}

// ProjectsListData is view data for the project dashboard.
type ProjectsListData struct {
	PageData
	Projects   []store.Project
	ActiveRuns []store.ActiveRunSummary
	CanCreate  bool
	Message    string
	Error      string
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
	Runs             []store.RunWithProgress
	NokItems         []store.ProjectNokItemSummary
	MemberRole       string
	CanManage        bool
	CanManageMembers bool
	CanLaunch        bool
	Message          string
	Error            string
}

// TemplatesIndexData is view data for the global templates tab.
type TemplatesIndexData struct {
	PageData
	Templates []store.TemplateIndexRow
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
	DueDate    string
	FormAction string
	Step       int
	MemberRole string
	CanLaunch  bool
	Error      string
}

// RunProgressData is view data for the run progress bar fragment.
type RunProgressData struct {
	RunID   int64
	Done    int
	Total   int
	Percent int
}

// RunItemRowData is view data for a single run item table row fragment.
type RunItemRowData struct {
	RunID          int64
	RunStatus      string
	Item           store.RunItem
	Members        []store.ProjectMember
	CSRFToken      string
	CanCheck       bool
	CanAssign      bool
	CanLinkJira    bool
	JiraConfigured bool
	JiraLink       store.IntegrationLink
	ItemError      string
	AssignError    string
}

// RunShowData is view data for run detail.
type RunShowData struct {
	PageData
	Project        *store.Project
	Run            *store.ChecklistRun
	Items          []store.RunItem
	NokItems       []store.RunItem
	JiraLinks      map[int64]store.IntegrationLink
	Members        []store.ProjectMember
	TemplateName   string
	VersionNum     int
	MemberRole     string
	CanLaunch      bool
	CanCheck       bool
	CanAssign      bool
	CanLinkJira    bool
	JiraConfigured bool
	CanComplete    bool
	Progress       RunProgressData
	ClosingNote    string
	Message        string
	ItemError      string
	AssignError    string
	CompleteError  string
	Error          string
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
	Project         *store.Project
	Run             *store.ChecklistRun
	Item            *store.RunItem
	Events          []store.RunItemEvent
	JiraLink        *store.IntegrationLink
	MemberRole      string
	CanCheck        bool
	CanLinkJira     bool
	JiraConfigured  bool
	JiraIssueInput  string
	Message         string
	LinkError       string
	CreateError     string
	ShowJiraCreate  bool
	JiraCreateTitle string
	JiraCreateDesc  string
}

// Parse loads layout and page templates from the embedded filesystem.
func Parse() (*template.Template, error) {
	root, err := fs.Sub(webassets.Templates, "templates")
	if err != nil {
		return nil, fmt.Errorf("templates root: %w", err)
	}

	tpl := template.New("").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"mul": func(a, b int) int { return a * b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"formatDueDate": formatDueDate,
		"dueDateInput":  dueDateInput,
		"runItemRow": func(run *store.ChecklistRun, item store.RunItem, members []store.ProjectMember, csrf string, canCheck, canAssign, canLinkJira, jiraConfigured bool, jiraLink store.IntegrationLink, itemErr, assignErr string) RunItemRowData {
			return RunItemRowData{
				RunID:          run.ID,
				RunStatus:      run.Status,
				Item:           item,
				Members:        members,
				CSRFToken:      csrf,
				CanCheck:       canCheck,
				CanAssign:      canAssign,
				CanLinkJira:    canLinkJira,
				JiraConfigured: jiraConfigured,
				JiraLink:       jiraLink,
				ItemError:      itemErr,
				AssignError:    assignErr,
			}
		},
	})

	tpl, err = tpl.ParseFS(root,
		"layouts/base.html",
		"partials/*.html",
		"pages/*.html",
	)
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	return tpl, nil
}

func formatDueDate(due sql.NullString) string {
	if !due.Valid || due.String == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, due.String)
	if err != nil {
		t, err = time.Parse("2006-01-02", due.String)
		if err != nil {
			return due.String
		}
	}
	return t.UTC().Format("02/01/2006")
}

func dueDateInput(due sql.NullString) string {
	if !due.Valid || due.String == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, due.String)
	if err != nil {
		t, err = time.Parse("2006-01-02", due.String)
		if err != nil {
			return ""
		}
	}
	return t.UTC().Format("2006-01-02")
}
