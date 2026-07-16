package subjects

import (
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/middleware"
	"github.com/jeb-maker/revues/internal/web/templates"
)

// Deps holds shared dependencies for the subjects HTTP handlers.
type Deps struct {
	Templates     *template.Template
	Store         SubjectStore
	SessionSecret string
}

// PageData builds shared view data with user and CSRF from the request context.
func (d *Deps) PageData(r *http.Request, title string) templates.PageData {
	data := templates.PageData{Title: title}
	if user, ok := middleware.UserFromContext(r.Context()); ok {
		data.User = user
		if token := middleware.SessionTokenFromContext(r); token != "" {
			data.CSRFToken = auth.CSRFToken(token, d.SessionSecret)
		}
	}
	templates.ApplyHeaderFromContext(r, &data)
	return data
}

// PageDataTab is PageData with ActiveTab set.
func (d *Deps) PageDataTab(r *http.Request, title, activeTab string) templates.PageData {
	data := d.PageData(r, title)
	data.ActiveTab = activeTab
	return data
}

func adminSubjectsRoute(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/admin/subjects")
}

func subjectsListPath(r *http.Request) string {
	if adminSubjectsRoute(r) {
		return templates.PathAdminSubjects
	}
	return templates.PathSubjects
}

func (d *Deps) subjectsPageData(r *http.Request, title string) templates.PageData {
	if adminSubjectsRoute(r) {
		data := d.PageDataTab(r, title, "org")
		data.AdminSection = "subjects"
		return data
	}
	return d.PageDataTab(r, title, "subjects")
}

// Subjects handles subject CRUD and the run wizard subject step.
type Subjects struct {
	Deps
}

// List shows subjects visible to the current user.
func (h *Subjects) List(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	admin := auth.HasMinRole(user.Role, auth.RoleAdmin)
	filterQuery := strings.TrimSpace(r.URL.Query().Get("q"))
	items, err := h.Store.ListSubjects(r.Context(), user.ID, admin, filterQuery)
	if err != nil {
		slog.Error("list subjects", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	orgRole, orgMember, _ := h.Store.OrganizationMemberRole(r.Context(), 0, user.ID)
	if org, ok := middleware.OrganizationFromContext(r.Context()); ok {
		orgRole, orgMember, _ = h.Store.OrganizationMemberRole(r.Context(), org.ID, user.ID)
	}

	pd := h.subjectsPageData(r, templates.DefaultUILabels().Subject.Plural)
	data := templates.SubjectsListData{
		PageData:          pd,
		Subjects:          items,
		FilterQuery:       filterQuery,
		HasActiveFilters:  filterQuery != "",
		CanCreate:         CanCreateSubject(user),
		CanManageOrgUsers: CanManageOrgUsers(user, orgRole, orgMember),
		Message:           r.URL.Query().Get("msg"),
	}
	if adminSubjectsRoute(r) {
		data.Breadcrumbs = templates.BCAdminSubjects(data.Labels.Subject)
	} else {
		data.Breadcrumbs = templates.BCSubjects(data.Labels.Subject)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "subjects_list", data); err != nil {
		slog.Error("render subjects list", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// NewForm renders the create subject form.
func (h *Subjects) NewForm(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if !CanCreateSubject(user) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	pd := h.subjectsPageData(r, "Nouveau "+strings.ToLower(templates.DefaultUILabels().Subject.Singular))
	if adminSubjectsRoute(r) {
		pd.Breadcrumbs = templates.BCAdminSubjectNew(pd.Labels.Subject)
	} else {
		pd.Breadcrumbs = templates.BCSubjectNew(pd.Labels.Subject)
	}
	data := templates.SubjectFormData{
		PageData:   pd,
		FormAction: subjectsListPath(r),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "subject_form", data); err != nil {
		slog.Error("render subject form", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Create stores a new subject.
func (h *Subjects) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if !CanCreateSubject(user) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		h.renderFormError(w, r, nil, subjectsListPath(r), "Le nom est obligatoire.")
		return
	}
	description := strings.TrimSpace(r.FormValue("description"))
	domains := store.ParseTagsCSV(r.FormValue("domains"))
	tags := store.ParseTagsCSV(r.FormValue("tags"))

	subject, err := h.Store.CreateSubject(r.Context(), name, description, user.ID, domains)
	if err != nil {
		slog.Error("create subject", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := h.Store.SetSubjectTags(r.Context(), subject.ID, tags); err != nil {
		slog.Error("set subject tags", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	msg := templates.DefaultUILabels().Subject.Singular + "+cr%C3%A9%C3%A9"
	http.Redirect(w, r, "/subjects/"+strconv.FormatInt(subject.ID, 10)+"?msg="+msg, http.StatusSeeOther)
}

// Show displays subject details.
func (h *Subjects) Show(w http.ResponseWriter, r *http.Request) {
	subject, user, memberRole, orgMember, ok := h.loadSubject(w, r)
	if !ok {
		return
	}

	members, err := h.Store.ListSubjectMembers(r.Context(), subject.ID)
	if err != nil {
		slog.Error("list subject members", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	subjectRuns, err := h.Store.ListRunsWithProgressBySubject(r.Context(), subject.ID)
	if err != nil {
		slog.Error("list subject runs", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	nokItems, err := h.Store.ListSubjectNokItems(r.Context(), subject.ID)
	if err != nil {
		slog.Error("list subject nok items", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	domains, err := h.Store.ListSubjectDomains(r.Context(), subject.ID)
	if err != nil {
		slog.Error("list subject domains", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tags, err := h.Store.ListSubjectTags(r.Context(), subject.ID)
	if err != nil {
		slog.Error("list subject tags", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	callerOrgRole, callerOrgMember, _ := h.Store.OrganizationMemberRole(r.Context(), subject.OrganizationID, user.ID)
	data, err := h.buildSubjectShowData(r, subject, user, memberRole, callerOrgRole, callerOrgMember, orgMember, domains, tags, members, subjectRuns, nokItems, subjectShowExtras{
		message: r.URL.Query().Get("msg"),
	})
	if err != nil {
		slog.Error("build subject show data", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "subject_show", data); err != nil {
		slog.Error("render subject show", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// EditForm renders the edit subject form.
func (h *Subjects) EditForm(w http.ResponseWriter, r *http.Request) {
	subject, user, _, _, ok := h.loadSubject(w, r)
	if !ok {
		return
	}
	callerOrgRole, callerOrgMember, err := h.Store.OrganizationMemberRole(r.Context(), subject.OrganizationID, user.ID)
	if err != nil {
		slog.Error("caller org role", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !CanManageSubject(user, callerOrgRole, callerOrgMember) {
		http.NotFound(w, r)
		return
	}

	pd := h.subjectsPageData(r, "Modifier "+subject.Name)
	if adminSubjectsRoute(r) {
		pd.Breadcrumbs = templates.BCAdminSubjectEdit(subject.Name, subject.ID, pd.Labels.Subject)
	} else {
		pd.Breadcrumbs = templates.BCSubjectEdit(subject.Name, subject.ID, pd.Labels.Subject)
	}
	data := templates.SubjectFormData{
		PageData:   pd,
		Subject:    subject,
		FormAction: subjectFormAction(r, subject.ID),
	}
	if domains, err := h.Store.ListSubjectDomains(r.Context(), subject.ID); err != nil {
		slog.Error("list subject domains", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	} else {
		data.Domains = store.FormatTagsCSV(domains)
	}
	if tags, err := h.Store.ListSubjectTags(r.Context(), subject.ID); err != nil {
		slog.Error("list subject tags", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	} else {
		data.Tags = store.FormatTagsCSV(tags)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "subject_form", data); err != nil {
		slog.Error("render subject edit", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Update saves subject fields.
func (h *Subjects) Update(w http.ResponseWriter, r *http.Request) {
	subject, user, _, _, ok := h.loadSubject(w, r)
	if !ok {
		return
	}
	callerOrgRole, callerOrgMember, err := h.Store.OrganizationMemberRole(r.Context(), subject.OrganizationID, user.ID)
	if err != nil {
		slog.Error("caller org role", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !CanManageSubject(user, callerOrgRole, callerOrgMember) {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		h.renderFormError(w, r, subject, subjectFormAction(r, subject.ID), "Le nom est obligatoire.")
		return
	}
	description := strings.TrimSpace(r.FormValue("description"))
	domains := store.ParseTagsCSV(r.FormValue("domains"))
	tags := store.ParseTagsCSV(r.FormValue("tags"))

	if err := h.Store.UpdateSubject(r.Context(), subject.ID, name, description, domains); err != nil {
		slog.Error("update subject", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := h.Store.SetSubjectTags(r.Context(), subject.ID, tags); err != nil {
		slog.Error("set subject tags", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	msg := templates.DefaultUILabels().Subject.Singular + "+mis+%C3%A0+jour"
	redirectPath := "/subjects/" + strconv.FormatInt(subject.ID, 10)
	if adminSubjectsRoute(r) {
		redirectPath = subjectsListPath(r)
	}
	http.Redirect(w, r, redirectPath+"?msg="+msg, http.StatusSeeOther)
}

// Archive marks the subject archived.
func (h *Subjects) Archive(w http.ResponseWriter, r *http.Request) {
	subject, user, _, _, ok := h.loadSubject(w, r)
	if !ok {
		return
	}
	callerOrgRole, callerOrgMember, err := h.Store.OrganizationMemberRole(r.Context(), subject.OrganizationID, user.ID)
	if err != nil {
		slog.Error("caller org role", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !CanManageSubject(user, callerOrgRole, callerOrgMember) {
		http.NotFound(w, r)
		return
	}

	if err := h.Store.ArchiveSubject(r.Context(), subject.ID); err != nil {
		slog.Error("archive subject", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	msg := templates.DefaultUILabels().Subject.Singular + "+archiv%C3%A9"
	http.Redirect(w, r, subjectsListPath(r)+"?msg="+msg, http.StatusSeeOther)
}

// WizardNouvelle is run wizard step 1: choose or search subjects.
func (h *Subjects) WizardNouvelle(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	templateID := parseWizardTemplateID(r)
	filterQuery := strings.TrimSpace(r.URL.Query().Get("q"))
	allSubjects, err := h.Store.ListSubjects(r.Context(), user.ID, auth.HasMinRole(user.Role, auth.RoleAdmin), filterQuery)
	if err != nil {
		slog.Error("list subjects for run wizard", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var launchSubjects []Subject
	for _, subject := range allSubjects {
		_, orgMember, err := h.Store.OrganizationMemberRole(r.Context(), subject.OrganizationID, user.ID)
		if err != nil {
			slog.Error("org member role", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if CanLaunchRun(user, orgMember) {
			launchSubjects = append(launchSubjects, subject)
		}
	}

	if len(launchSubjects) == 1 && filterQuery == "" {
		http.Redirect(w, r, templates.SubjectTemplatesForRunPath(launchSubjects[0].ID, templateID), http.StatusSeeOther)
		return
	}

	pd := h.PageData(r, "Lancer une revue")
	pd.Breadcrumbs = templates.BCRunWizardSubjects()
	pd.ActiveTab = "runs"
	data := templates.RunWizardSubjectsData{
		PageData:           pd,
		Subjects:           launchSubjects,
		SelectedTemplateID: templateID,
		Message:            r.URL.Query().Get("msg"),
		FilterQuery:        filterQuery,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "run_wizard_subjects", data); err != nil {
		slog.Error("render run wizard step 1", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// WizardNouvelleCreate selects or inline-creates a subject for the run wizard.
func (h *Subjects) WizardNouvelleCreate(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	subjectIDRaw := strings.TrimSpace(r.FormValue("subject_id"))
	templateID := parseWizardTemplateID(r)
	if templateID <= 0 {
		if raw := strings.TrimSpace(r.FormValue("template")); raw != "" {
			if id, err := strconv.ParseInt(raw, 10, 64); err == nil && id > 0 {
				templateID = id
			}
		}
	}
	if subjectIDRaw != "" {
		subjectID, err := strconv.ParseInt(subjectIDRaw, 10, 64)
		if err != nil {
			h.renderWizardError(w, r, "Sujet invalide.")
			return
		}
		subject, _, _, orgMember, ok := h.loadSubjectByID(w, r, subjectID)
		if !ok {
			return
		}
		if !CanLaunchRun(user, orgMember) {
			http.NotFound(w, r)
			return
		}
		_ = subject
		http.Redirect(w, r, templates.SubjectTemplatesForRunPath(subjectID, templateID), http.StatusSeeOther)
		return
	}

	if !CanCreateSubject(user) {
		http.NotFound(w, r)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		h.renderWizardError(w, r, "Le nom est obligatoire.")
		return
	}

	subject, err := h.Store.CreateSubject(r.Context(), name, "", user.ID, store.ParseTagsCSV(r.FormValue("domains")))
	if err != nil {
		slog.Error("inline create subject", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, templates.SubjectTemplatesForRunPath(subject.ID, templateID), http.StatusSeeOther)
}

func parseWizardTemplateID(r *http.Request) int64 {
	raw := strings.TrimSpace(r.URL.Query().Get("template"))
	if raw == "" {
		return 0
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0
	}
	return id
}

func (h *Subjects) loadSubject(w http.ResponseWriter, r *http.Request) (*Subject, *store.User, string, bool, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return nil, nil, "", false, false
	}
	return h.loadSubjectByID(w, r, id)
}

func (h *Subjects) loadSubjectByID(w http.ResponseWriter, r *http.Request, id int64) (*Subject, *store.User, string, bool, bool) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return nil, nil, "", false, false
	}

	subject, err := h.Store.SubjectByID(r.Context(), id)
	if errors.Is(err, ErrSubjectNotFound) {
		http.NotFound(w, r)
		return nil, nil, "", false, false
	}
	if err != nil {
		slog.Error("load subject", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, "", false, false
	}

	_, orgMember, err := h.Store.OrganizationMemberRole(r.Context(), subject.OrganizationID, user.ID)
	if err != nil {
		slog.Error("org member role", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, nil, "", false, false
	}

	if !CanViewSubject(user, orgMember) {
		http.NotFound(w, r)
		return nil, nil, "", false, false
	}

	memberRole := "lead"
	if orgMember {
		memberRole = "lead"
	}

	return subject, user, memberRole, orgMember, true
}

func subjectFormAction(r *http.Request, id int64) string {
	if adminSubjectsRoute(r) {
		return templates.PathAdminSubjects + "/" + strconv.FormatInt(id, 10)
	}
	return templates.PathSubjects + "/" + strconv.FormatInt(id, 10)
}

func (h *Subjects) renderFormError(w http.ResponseWriter, r *http.Request, subject *Subject, action, message string) {
	title := "Nouveau " + strings.ToLower(templates.DefaultUILabels().Subject.Singular)
	if subject != nil {
		title = "Modifier " + subject.Name
	}
	pd := h.subjectsPageData(r, title)
	if subject != nil {
		if adminSubjectsRoute(r) {
			pd.Breadcrumbs = templates.BCAdminSubjectEdit(subject.Name, subject.ID, pd.Labels.Subject)
		} else {
			pd.Breadcrumbs = templates.BCSubjectEdit(subject.Name, subject.ID, pd.Labels.Subject)
		}
	} else if adminSubjectsRoute(r) {
		pd.Breadcrumbs = templates.BCAdminSubjectNew(pd.Labels.Subject)
	} else {
		pd.Breadcrumbs = templates.BCSubjectNew(pd.Labels.Subject)
	}
	data := templates.SubjectFormData{
		PageData:   pd,
		Subject:    subject,
		FormAction: action,
		Error:      message,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	if err := h.Templates.ExecuteTemplate(w, "subject_form", data); err != nil {
		slog.Error("render subject form error", "err", err)
	}
}

func (h *Subjects) renderWizardError(w http.ResponseWriter, r *http.Request, message string) {
	templateID := parseWizardTemplateID(r)
	pd := h.PageData(r, "Lancer une revue")
	pd.Breadcrumbs = templates.BCRunWizardSubjects()
	pd.ActiveTab = "runs"
	data := templates.RunWizardSubjectsData{
		PageData:           pd,
		SelectedTemplateID: templateID,
		Error:              message,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	if err := h.Templates.ExecuteTemplate(w, "run_wizard_subjects", data); err != nil {
		slog.Error("render run wizard error", "err", err)
	}
}

type subjectShowExtras struct {
	message string
}

func (h *Subjects) buildSubjectShowData(
	r *http.Request,
	subject *Subject,
	user *store.User,
	memberRole, callerOrgRole string,
	callerOrgMember, orgMember bool,
	domains, tags []string,
	members []SubjectMember,
	subjectRuns []RunWithProgress,
	nokItems []SubjectNokItemSummary,
	extras subjectShowExtras,
) (templates.SubjectShowData, error) {
	canManage := CanManageSubject(user, callerOrgRole, callerOrgMember)
	canLaunch := CanLaunchRun(user, orgMember)
	editPath := templates.PathSubjects + "/" + strconv.FormatInt(subject.ID, 10) + "/edit"
	if CanManageOrgUsers(user, callerOrgRole, callerOrgMember) {
		editPath = templates.PathAdminSubjects + "/" + strconv.FormatInt(subject.ID, 10) + "/edit"
	}

	pd := h.PageDataTab(r, subject.Name, "")
	pd.Breadcrumbs = templates.BCSubjectShow(subject.Name, pd.Labels.Subject)

	return templates.SubjectShowData{
		PageData:         pd,
		Subject:          subject,
		Domains:          domains,
		Tags:             tags,
		Members:          members,
		Runs:             subjectRuns,
		NokItems:         nokItems,
		MemberRole:       memberRole,
		CanManage:        canManage,
		CanManageMembers: false,
		CanLaunch:        canLaunch,
		EditPath:         editPath,
		Message:          extras.message,
	}, nil
}
