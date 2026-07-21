package templates

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"net/url"
	"strconv"
	"strings"
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

// PageAction is a header action link (primary CTA or secondary).
type PageAction struct {
	Label    string
	Title    string // tooltip and aria-label; defaults to Label when empty
	URL      string
	Primary  bool
	IconOnly bool
}

// PrimaryAction returns a primary header button action.
func PrimaryAction(label, url string) PageAction {
	return PageAction{Label: label, Title: label, URL: url, Primary: true}
}

// CreateAction returns a compact "+" create button with an accessible title.
func CreateAction(title, url string) PageAction {
	return PageAction{Label: "+", Title: title, URL: url, Primary: true, IconOnly: true}
}

// LaunchAction returns the standard launch-revue header button.
func LaunchAction(url string, subject SubjectUILabels, run RunUILabels) PageAction {
	return PageAction{Label: LaunchRunCTA(run), Title: LaunchActionTitle(subject, run), URL: url, Primary: true}
}

// SecondaryAction returns a secondary header action.
func SecondaryAction(label, url string) PageAction {
	return PageAction{Label: label, Title: label, URL: url, Primary: false}
}

// FormatItemStatus returns a French label for a checklist item status.
func FormatItemStatus(s string) string {
	switch s {
	case store.RunItemStatusPending:
		return "En attente"
	case store.RunItemStatusOK:
		return "OK"
	case store.RunItemStatusNOK:
		return "Non validé"
	case store.RunItemStatusNA:
		return "N/A"
	default:
		return s
	}
}

// FormatRunStatus returns a French label for a review run status.
func FormatRunStatus(s string) string {
	switch s {
	case store.RunStatusDraft:
		return "Brouillon"
	case store.RunStatusInProgress:
		return "En cours"
	case store.RunStatusDone:
		return "Terminée"
	case store.RunStatusArchived:
		return "Archivée"
	default:
		return s
	}
}

// FormatRole returns a French label for organization or project roles.
//
//nolint:misspell // French UI labels
func FormatRole(s string) string {
	switch s {
	case "reader":
		return "Lecteur"
	case "editor":
		return "Éditeur"
	case "admin":
		return "Administrateur"
	case "lead":
		return "Responsable"
	case "contributor":
		return "Contributeur"
	case "viewer":
		return "Observateur"
	case "owner":
		return "Propriétaire"
	case "member":
		return "Membre"
	default:
		return s
	}
}

// FormatAccessSource returns a French badge label for a ResolveSubjectAccess source.
// teamNames maps team id → name for "team:{id}" sources.
func FormatAccessSource(source string, teamNames map[int64]string) string {
	switch {
	case source == store.AccessSourceDirect:
		return "direct"
	case source == store.AccessSourceOrgAdmin:
		return "admin organisation"
	case source == store.AccessSourceGlobalAdmin:
		return "admin global"
	case source == store.AccessSourceOrgMemberLegacy:
		return "membre organisation"
	case strings.HasPrefix(source, "team:"):
		id, err := strconv.ParseInt(strings.TrimPrefix(source, "team:"), 10, 64)
		if err == nil {
			if name, ok := teamNames[id]; ok && name != "" {
				return "via équipe " + name
			}
		}
		return "via équipe"
	default:
		return source
	}
}

// TeamAssignPreview formats the pre-add team assignment message.
func TeamAssignPreview(teamName string, memberCount int, role string) string {
	roleLabel := FormatRole(role)
	if memberCount == 1 {
		return fmt.Sprintf("Équipe %s : 1 membre aura le rôle %s", teamName, roleLabel)
	}
	return fmt.Sprintf("Équipe %s : %d membres auront le rôle %s", teamName, memberCount, roleLabel)
}

// RunItemTableColspan returns the column count for the run items table empty row.
func RunItemTableColspan(runStatus string, canCheck, canAssign, showAssign bool) int {
	n := 4 // Point, Statut, Commentaire, PJ
	if showAssign {
		n++
	}
	if runStatus == store.RunStatusInProgress && (canCheck || canAssign) {
		n++ // Actions
	}
	return n
}

// PageData is shared view data for HTML pages.
type PageData struct {
	Title               string
	User                *store.User
	CSRFToken           string
	LoginError          string
	DevAuth             bool
	DevAuthUsers        []store.User
	ActiveTab           string
	AdminSection        string
	CanManageOrgUsers   bool
	ShowOrganisationNav bool
	SimpleUI            bool
	SimpleSubjectID     int64
	ShowAssign          bool
	ShowMyTasks         bool
	ShowSubjectColumn   bool
	ShowCollab          bool
	RequestID           string
	ReportsAutoOpen     bool // open @jeb-maker/reports widget on load (/signaler)
	Breadcrumbs         []Breadcrumb
	PageActions         []PageAction
	Labels              UILabels
	ActiveOrganization  *store.Organization
	UserOrganizations   []store.OrganizationMembership
	PendingInvitations  []store.OrganizationInvitation
}

// ReportsMetadata returns trusted session context for the @jeb-maker/reports widget.
// The server re-applies identity on POST /signaler/api; this is for diagnostics only.
func (d PageData) ReportsMetadata() map[string]any {
	if d.User == nil {
		return nil
	}
	meta := map[string]any{
		"app":        "revues",
		"user_id":    d.User.ID,
		"user_login": d.User.Login,
		"user_role":  d.User.Role,
		"simple_ui":  d.SimpleUI,
		"request_id": d.RequestID,
		"ui_caps": map[string]any{
			"simple_ui":           d.SimpleUI,
			"show_assign":         d.ShowAssign,
			"show_my_tasks":       d.ShowMyTasks,
			"show_subject_column": d.ShowSubjectColumn,
			"show_collab":         d.ShowCollab,
		},
	}
	if d.ActiveOrganization != nil {
		meta["org_id"] = d.ActiveOrganization.ID
		meta["org_name"] = d.ActiveOrganization.Name
		meta["ui_run_label"] = d.ActiveOrganization.UIRunLabel
	}
	return meta
}

// AdminUsersData is view data for the org-scoped whitelist admin screen.
type AdminUsersData struct {
	PageData
	OrganizationName string
	Emails           []store.AllowedEmail
	Message          string
	Error            string
}

// AdminTeamsData is view data for the org teams list / create screen.
type AdminTeamsData struct {
	PageData
	OrganizationName string
	Teams            []store.OrganizationTeam
	Message          string
	Error            string
}

// AdminTeamDetailData is view data for a single team and its members.
type AdminTeamDetailData struct {
	PageData
	OrganizationName string
	Team             store.OrganizationTeam
	Members          []store.TeamMember
	Candidates       []store.OrganizationMemberUser
	Message          string
	Error            string
}

// AdminOrgHubData is view data for the organisation admin landing page.
type AdminOrgHubData struct {
	PageData
	OrganizationName string
}

// AdminSubjectLabelsData is view data for the org subject + run label preset screen.
type AdminSubjectLabelsData struct {
	PageData
	Presets    []SubjectLabelPreset
	Current    string
	RunPresets []RunLabelPreset
	CurrentRun string
	Message    string
	Error      string
}

// AdminLeadPoliciesData is view data for org lead-delegation policies.
type AdminLeadPoliciesData struct {
	PageData
	Policies store.OrgLeadPolicies
	Message  string
	Error    string
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

// BugReportContext is auto-captured request/session context shown on the form.
type BugReportContext struct {
	PageURL           string
	UserID            int64
	UserLogin         string
	UserEmail         string
	UserDisplayName   string
	UserRole          string
	OrgID             int64
	OrgName           string
	OrgRole           string
	UIRunLabel        string
	SimpleUI          bool
	ShowAssign        bool
	ShowMyTasks       bool
	ShowSubjectColumn bool
	ShowCollab        bool
	Timestamp         string
	UserAgent         string
	RequestID         string
}

// UICapsMap returns progressive-disclosure flags for persistence.
func (c BugReportContext) UICapsMap() map[string]any {
	return map[string]any{
		"simple_ui":           c.SimpleUI,
		"show_assign":         c.ShowAssign,
		"show_my_tasks":       c.ShowMyTasks,
		"show_subject_column": c.ShowSubjectColumn,
		"show_collab":         c.ShowCollab,
	}
}

// BugReportData is view data for the in-app bug report form.
type BugReportData struct {
	PageData
	Context     BugReportContext
	PageURL     string
	TitleValue  string
	Description string
	Steps       string
	Severity    string
	ReturnURL   string
	Message     string
	Error       string
}

// RunsListData is view data for the runs index page.
type RunsListData struct {
	PageData
	Runs              []store.RunListSummary
	FilterQuery       string
	FilterStatus      string
	HasActiveFilters  bool
	HasSubjects       bool
	CanCreate         bool
	CanLaunch         bool
	CanManageOrgUsers bool
	Pagination        Pagination
	Message           string
	Error             string
}

// Pagination holds list page navigation.
type Pagination struct {
	Page       int
	PageSize   int
	Total      int
	TotalPages int
	HasPrev    bool
	HasNext    bool
	PrevURL    string
	NextURL    string
}

// NewPagination builds pagination links for a filtered list.
func NewPagination(page, pageSize, total int, urlForPage func(page int) string) Pagination {
	if pageSize <= 0 {
		pageSize = 25
	}
	if page < 1 {
		page = 1
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	if totalPages > 0 && page > totalPages {
		page = totalPages
	}
	p := Pagination{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		HasPrev:    page > 1 && totalPages > 0,
		HasNext:    totalPages > 0 && page < totalPages,
	}
	if p.HasPrev {
		p.PrevURL = urlForPage(page - 1)
	}
	if p.HasNext {
		p.NextURL = urlForPage(page + 1)
	}
	return p
}

// SubjectsListData is view data for the subjects dashboard.
type SubjectsListData struct {
	PageData
	Subjects          []store.Subject
	FilterQuery       string
	HasActiveFilters  bool
	CanCreate         bool
	CanManageOrgUsers bool
	Message           string
	Error             string
}

// SubjectFormData is view data for create/edit subject forms.
type SubjectFormData struct {
	PageData
	Subject          *store.Subject
	Domains          string
	Tags             string
	FormAction       string
	CanSetVisibility bool
	Error            string
}

// OrgNewData is view data for the organization creation form.
type OrgNewData struct {
	PageData
	Name  string
	Slug  string
	Error string
}

// OrgSelectData is view data for the organization selection screen.
type OrgSelectData struct {
	PageData
	Organizations []store.OrganizationMembership
	DefaultOrgID  int64
	Error         string
}

// SubjectTeamPreviewData is the HTMX fragment for team assignment preview.
type SubjectTeamPreviewData struct {
	Empty       bool
	TeamName    string
	MemberCount int
	Role        string // lead | contributor | viewer
	RoleLabel   string
}

// SubjectShowData is view data for subject detail.
type SubjectShowData struct {
	PageData
	Subject             *store.Subject
	Domains             []string
	Tags                []string
	Members             []store.SubjectMember
	DirectMembers       []store.DirectSubjectMember
	Teams               []store.TeamSubjectRole
	AvailableTeams      []store.OrganizationTeam
	AccessSources       []string
	Runs                []store.RunWithProgress
	NokItems            []store.SubjectNokItemSummary
	MemberRole          string
	CanManage           bool
	CanManageMembers    bool
	CanAssignTeams      bool
	CanManageOrgUsers   bool
	TeamsPolicyDenied   bool
	MembersPolicyDenied bool
	CanLaunch           bool
	EditPath            string
	AddMemberEmail      string
	AddMemberRole       string
	AddTeamID           int64
	AddTeamRole         string
	Message             string
	Error               string
}

// TemplatesIndexData is view data for the global templates tab.
type TemplatesIndexData struct {
	PageData
	Templates        []store.TemplateIndexRow
	FilterQuery      string
	HasActiveFilters bool
	CanManage        bool
	NotionConfigured bool
	Message          string
}

// TemplateEditorRow is one editable checklist point in the form.
type TemplateEditorRow struct {
	RowIndex int
	Label    string
	HelpText string
	Required bool
}

// TemplateEditorSection is a group of checklist points under one section title.
type TemplateEditorSection struct {
	SectionIndex int
	Title        string
	Items        []TemplateEditorRow
}

// TemplateItemSection groups stored template items for read-only display.
type TemplateItemSection struct {
	Title string
	Items []store.TemplateItem
}

type ChecklistTemplateNotionImportData struct {
	PageData
	CanManage        bool
	NotionConfigured bool
	Step             string
	FormAction       string
	DatabaseRef      string
	DatabaseID       string
	DatabaseTitle    string
	TemplateName     string
	Tags             string
	Properties       []NotionPropertyOption
	Mapping          notion.ColumnMapping
	PreviewItems     []TemplateEditorRow
	PreviewCount     int
	Error            string
}
type NotionPropertyOption struct{ Name, Type string }

// ChecklistTemplatesListData is view data for template index on a subject.
type ChecklistTemplatesListData struct {
	PageData
	Subject                    *store.Subject
	Templates                  []store.ChecklistTemplateSummary
	MemberRole                 string
	CanManage                  bool
	ForRun                     bool
	SelectedTemplateID         int64
	SelectedTemplateName       string
	SelectedTemplateCompatible bool
	FilterQuery                string
	HasActiveFilters           bool
	Message                    string
	Error                      string
}

// ChecklistTemplateFormData is view data for create/edit template forms.
type ChecklistTemplateFormData struct {
	PageData
	Template         *store.ChecklistTemplate
	Version          *store.TemplateVersion
	Name             string
	Tags             string
	TagsList         []string
	Sections         []TemplateEditorSection
	SectionsEnabled  bool
	MaxItemLabelLen  int
	MaxItemHelpLen   int
	NameError        string
	ItemsError       string
	FormAction       string
	NotionConfigured bool
	Error            string
}

// ChecklistTemplateShowData is view data for template detail.
type ChecklistTemplateShowData struct {
	PageData
	Template     *store.ChecklistTemplate
	Version      *store.TemplateVersion
	Tags         []string
	ItemSections []TemplateItemSection
	ItemCount    int
	CanManage    bool
	CanLaunch    bool
	Message      string
	Error        string
}

// RunWizardSubjectsData is view data for run wizard step 1.
type RunWizardSubjectsData struct {
	PageData
	Subjects           []store.Subject
	SelectedTemplateID int64
	FilterQuery        string
	CanCreate          bool
	Message            string
	Error              string
}

// RunWizardLaunchData is view data for run wizard step 3.
type RunWizardLaunchData struct {
	PageData
	Subject    *store.Subject
	Template   *store.ChecklistTemplate
	Version    *store.TemplateVersion
	ItemCount  int
	Title      string
	DueDate    string
	FormAction string
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

// RunCompleteStatusData is view data for the run complete-section status fragment.
type RunCompleteStatusData struct {
	Run      *store.ChecklistRun
	NokItems []store.RunItem
	Progress RunProgressData
}

// RunItemRowData is view data for a single run item table row fragment.
type RunItemRowData struct {
	RunID          int64
	RunStatus      string
	Item           store.RunItem
	Members        []store.SubjectMember
	CSRFToken      string
	CanCheck       bool
	CanAssign      bool
	ShowAssign     bool // false in SimpleUI (solo / particulier)
	CanLinkJira    bool
	JiraConfigured bool
	JiraLink       store.IntegrationLink
	Attachment     *store.Attachment
	ItemError      string
	AssignError    string
}

// RunItemSectionData is view data for one collapsible section of run items.
type RunItemSectionData struct {
	Title      string
	Items      []store.RunItem
	Total      int
	OKCount    int
	NonOKCount int
	AllOKOrNA  bool
}

// RunShowData is view data for run detail.
type RunShowData struct {
	PageData
	Subject           *store.Subject
	Run               *store.ChecklistRun
	RunDisplayLabel   string
	Items             []store.RunItem
	ItemSections      []RunItemSectionData
	NokItems          []store.RunItem
	Sections          []string
	FilterSection     string
	FilterStatus      string
	JiraLinks         map[int64]store.IntegrationLink
	Attachments       map[int64]*store.Attachment
	Members           []store.SubjectMember
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
	CanExportEvidence bool
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
	Tasks            []store.AssignedRunItemSummary
	FilterQuery      string
	FilterStatus     string
	HasActiveFilters bool
}

// RunItemShowData is view data for run item detail with audit history.
type RunItemShowData struct {
	PageData
	Subject               *store.Subject
	Run                   *store.ChecklistRun
	RunDisplayLabel       string
	Item                  *store.RunItem
	Events                []store.RunItemEvent
	JiraLink              *store.IntegrationLink
	MemberRole            string
	CanCheck              bool
	CanLinkJira           bool
	JiraConfigured        bool
	CanManageIntegrations bool
	JiraIssueInput        string
	Message               string
	LinkError             string
	CreateError           string
	ShowJiraCreate        bool
	JiraCreateTitle       string
	JiraCreateDesc        string
	Attachment            *store.Attachment
	CanUpload             bool
	UploadError           string
}

// Parse loads layout and page templates from the embedded filesystem.
// assetVersion is appended to static URLs for cache busting (may be empty in tests).
func Parse(assetVersion string) (*template.Template, error) {
	root, err := fs.Sub(webassets.Templates, "templates")
	if err != nil {
		return nil, fmt.Errorf("templates root: %w", err)
	}

	tpl := template.New("").Funcs(template.FuncMap{
		"assetURL": func(path string) string {
			if assetVersion == "" {
				return path
			}
			return path + "?v=" + assetVersion
		},
		"toJSON": func(v any) (template.JS, error) {
			b, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return template.JS(b), nil
		},
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
		"breadcrumbCurrent": func(crumbs []Breadcrumb) string {
			return BreadcrumbCurrent(crumbs)
		},
		"breadcrumbAncestors": func(crumbs []Breadcrumb) []Breadcrumb {
			return BreadcrumbAncestors(crumbs)
		},
		"mul": func(a, b int) int { return a * b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"formatItemStatus":    FormatItemStatus,
		"formatRunStatus":     FormatRunStatus,
		"formatRole":          FormatRole,
		"formatAccessSource":  FormatAccessSource,
		"teamAssignPreview":   TeamAssignPreview,
		"lowerFirst":          LowerFirst,
		"launchActionTitle":   LaunchActionTitle,
		"launchRunCTA":        LaunchRunCTA,
		"runItemTableColspan": RunItemTableColspan,
		"formatDueDate":       formatDueDate,
		"formatDateTime":      formatDateTime,
		"dueDateInput":        dueDateInput,
		"runsListURL": func(status, q string, page int) string {
			return RunsListURL(status, q, page)
		},
		"listURL": func(path, q string) string {
			return listURL(path, url.Values{"q": {strings.TrimSpace(q)}})
		},
		"runWizardPath": func(templateID int64) string {
			return RunWizardPath(templateID)
		},
		"subjectModelesListPath": func(subjectID int64, forRun bool, templateID int64) string {
			return SubjectModelesListPath(subjectID, forRun, templateID)
		},
		"subjectTemplatesForRunPath": func(subjectID, templateID int64) string {
			return SubjectTemplatesForRunPath(subjectID, templateID)
		},
		"myTasksListURL": func(status, q string) string {
			return listURL("/mes-taches", url.Values{
				"status": {status},
				"q":      {strings.TrimSpace(q)},
			})
		},
		"attachmentIsImage": func(att *store.Attachment) bool {
			if att == nil {
				return false
			}
			return attachments.IsImageMime(att.MimeType)
		},
		"runItemRow": func(run *store.ChecklistRun, item store.RunItem, members []store.SubjectMember, csrf string, canCheck, canAssign, showAssign, canLinkJira, jiraConfigured bool, jiraLink store.IntegrationLink, attachment *store.Attachment, itemErr, assignErr string) RunItemRowData {
			return RunItemRowData{
				RunID:          run.ID,
				RunStatus:      run.Status,
				Item:           item,
				Members:        members,
				CSRFToken:      csrf,
				CanCheck:       canCheck,
				CanAssign:      canAssign && showAssign,
				ShowAssign:     showAssign,
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
	return formatDateTime(due.String)
}

func formatDateTime(value string) string {
	if value == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t, err = time.Parse("2006-01-02", value)
		if err != nil {
			return value
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

func listURL(path string, values url.Values) string {
	clean := url.Values{}
	for key, vals := range values {
		for _, val := range vals {
			val = strings.TrimSpace(val)
			if val != "" {
				clean.Add(key, val)
			}
		}
	}
	if enc := clean.Encode(); enc != "" {
		return path + "?" + enc
	}
	return path
}

// RunsListURL builds the /revues list URL with optional filters and page.
func RunsListURL(status, q string, page int) string {
	values := url.Values{
		"status": {status},
		"q":      {strings.TrimSpace(q)},
	}
	if page > 1 {
		values.Set("page", strconv.Itoa(page))
	}
	return listURL("/revues", values)
}
