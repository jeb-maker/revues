package templates

import (
	"database/sql"
	"fmt"
	"html/template"
	"io/fs"
	"time"

	"github.com/jeb-maker/revues/internal/attachments"
	"github.com/jeb-maker/revues/internal/integrations/notion"
	"github.com/jeb-maker/revues/internal/store"
	webassets "github.com/jeb-maker/revues/web"
)

type Breadcrumb struct {
	URL   string
	Label string
}

type StepperStep struct {
	Label  string
	Status string
}

type StepperData struct {
	Steps []StepperStep
}

// PageData is shared view data for HTML pages.
type PageData struct {
	Title        string
	User         *store.User
	CSRFToken    string
	LoginError   string
	ActiveTab    string
	AdminSection string
	Breadcrumbs  []Breadcrumb
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

type ChecklistTemplateNotionImportData struct {
	PageData
	Project          *store.Project
	MemberRole       string
	CanManage        bool
	NotionConfigured bool
	Step             string
	FormAction       string
	DatabaseRef      string
	DatabaseID       string
	DatabaseTitle    string
	TemplateName     string
	Properties       []NotionPropertyOption
	Mapping          notion.ColumnMapping
	PreviewItems     []TemplateEditorRow
	PreviewCount     int
	Error            string
}
type NotionPropertyOption struct{ Name, Type string }

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

// TemplateNewWizardData is view data for the "choose project" step when creating a template.
type TemplateNewWizardData struct {
	PageData
	Projects []store.Project
}

// TemplateRowFragmentData is view data for a single template editor row fragment (HTMX).
type TemplateRowFragmentData struct {
	ProjectID int64
	Index     int
	Section   string
	Label     string
	HelpText  string
	Required  bool
	CSRFToken string
}

// RunWizardProjectsData is view data for run wizard step 1.
type RunWizardProjectsData struct {
	PageData
	Projects []store.Project
	Step     int
	Stepper  StepperData
	Message  string
	Error    string
}

// RunWizardTemplatesData is view data for run wizard step 2.
type RunWizardTemplatesData struct {
	PageData
	Project    *store.Project
	Templates  []store.ChecklistTemplateSummary
	Step       int
	Stepper    StepperData
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
	Stepper    StepperData
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
	Attachment     *store.Attachment
	ItemError      string
	AssignError    string
}

// RunShowData is view data for run detail.
type RunShowData struct {
	PageData
	Project           *store.Project
	Run               *store.ChecklistRun
	Items             []store.RunItem
	NokItems          []store.RunItem
	Sections          []string
	FilterSection     string
	FilterStatus      string
	JiraLinks         map[int64]store.IntegrationLink
	Attachments       map[int64]*store.Attachment
	Members           []store.ProjectMember
	TemplateName      string
	VersionNum        int
	MemberRole        string
	CanLaunch         bool
	CanCheck          bool
	CanAssign         bool
	CanLinkJira       bool
	JiraConfigured    bool
	CanComplete       bool
	NotionConfigured  bool
	CanExportNotion   bool
	Progress          RunProgressData
	ClosingNote       string
	Message           string
	ItemError         string
	AssignError       string
	CompleteError     string
	NotionExportError string
	Error             string
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
	Attachment      *store.Attachment
	CanUpload       bool
	UploadError     string
}

// Parse loads layout and page templates from the embedded filesystem.
func Parse() (*template.Template, error) {
	root, err := fs.Sub(webassets.Templates, "templates")
	if err != nil {
		return nil, fmt.Errorf("templates root: %w", err)
	}

	tpl := template.New("").Funcs(template.FuncMap{
		"icon": func(name string) template.HTML {
			switch name {
			case "plus":
				return `<svg class="icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M12 4v16m-8-8h16"/></svg>`
			case "x":
				return `<svg class="icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M6 6l12 12M6 18L18 6"/></svg>`
			case "arrow-left":
				return `<svg class="icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M19 12H5m6-6l-6 6 6 6"/></svg>`
			case "chevron-right":
				return `<svg class="icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M10 6l6 6-6 6"/></svg>`
			case "menu":
				return `<svg class="icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M4 6h16M4 12h16M4 18h16"/></svg>`
			case "download":
				return `<svg class="icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4m5-5l4 4 4-4m-4 4V3"/></svg>`
			case "external-link":
				return `<svg class="icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M7 17L17 7m0 0H9m8 0v8"/></svg>`
			case "check":
				return `<svg class="icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M5 13l4 4L19 7"/></svg>`
			case "alert":
				return `<svg class="icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M12 8v4m0 4h.01M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20z"/></svg>`
			case "settings":
				return `<svg class="icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09a1.65 1.65 0 0 0-1.08-1.51 1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09a1.65 1.65 0 0 0 1.51-1.08 1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1.08 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1.08z"/></svg>`
			case "users":
				return `<svg class="icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>`
			default:
				return ``
			}
		},
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"mul": func(a, b int) int { return a * b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"statusLabel": func(s string) string {
			switch s {
			case "pending":
				return "en_attente"
			case "ok":
				return "ok"
			case "nok":
				return "nok"
			case "na":
				return "non_applicable"
			default:
				return s
			}
		},
		"formatDueDate": formatDueDate,
		"dueDateInput":  dueDateInput,
		"attachmentIsImage": func(att *store.Attachment) bool {
			if att == nil {
				return false
			}
			return attachments.IsImageMime(att.MimeType)
		},
		"runItemRow": func(run *store.ChecklistRun, item store.RunItem, members []store.ProjectMember, csrf string, canCheck, canAssign, canLinkJira, jiraConfigured bool, jiraLink store.IntegrationLink, attachment *store.Attachment, itemErr, assignErr string) RunItemRowData {
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
				Attachment:     attachment,
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
