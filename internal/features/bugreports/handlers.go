package bugreports

import (
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/middleware"
	viewtemplates "github.com/jeb-maker/revues/internal/web/templates"
)

const (
	formPath       = "/signaler"
	maxTitleLen    = 200
	maxDescLen     = 8000
	maxStepsLen    = 4000
	maxPageURLLen  = 2000
	severityLow    = "low"
	severityNormal = "normal"
	severityHigh   = "high"
)

// Deps holds shared dependencies for bug report handlers.
type Deps struct {
	Templates     *template.Template
	SessionSecret string
	ReportsDir    string
}

// BugReports handles the in-app bug report form.
type BugReports struct {
	Deps
}

// PageData builds shared view data with user and CSRF from the request context.
func (d *Deps) PageData(r *http.Request, title string) viewtemplates.PageData {
	data := viewtemplates.PageData{Title: title}
	if user, ok := middleware.UserFromContext(r.Context()); ok {
		data.User = user
		if token := middleware.SessionTokenFromContext(r); token != "" {
			data.CSRFToken = auth.CSRFToken(token, d.SessionSecret)
		}
	}
	viewtemplates.ApplyHeaderFromContext(r, &data)
	return data
}

// Form shows the bug report form with auto-captured context summary.
func (h *BugReports) Form(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	ctxSummary := buildContextSummary(r, user)
	pageURL := resolvePageURL(r.URL.Query().Get("from"), r.Referer(), "")
	if pageURL != "" {
		ctxSummary.PageURL = pageURL
	}

	pageData := viewtemplates.ApplyPageMeta(h.PageData(r, "Signaler un problème"), viewtemplates.BCBugReport())
	pageData.ReportsAutoOpen = true
	data := viewtemplates.BugReportData{
		PageData:  pageData,
		Context:   ctxSummary,
		PageURL:   ctxSummary.PageURL,
		Severity:  severityNormal,
		Message:   r.URL.Query().Get("msg"),
		ReturnURL: safeReturnPath(ctxSummary.PageURL),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "bug_report", data); err != nil {
		slog.Error("render bug report form", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Create validates the form, persists the report, and redirects with confirmation.
func (h *BugReports) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	steps := strings.TrimSpace(r.FormValue("steps"))
	severity := normalizeSeverity(r.FormValue("severity"))
	pageURL := resolvePageURL(r.FormValue("page_url"), r.Referer(), r.FormValue("from"))

	ctxSummary := buildContextSummary(r, user)
	ctxSummary.PageURL = pageURL

	if title == "" || description == "" {
		h.renderFormError(w, r, ctxSummary, title, description, steps, severity,
			"Le titre et la description sont obligatoires.")
		return
	}
	if len(title) > maxTitleLen {
		h.renderFormError(w, r, ctxSummary, title, description, steps, severity,
			"Le titre est trop long (200 caractères max).")
		return
	}
	if len(description) > maxDescLen {
		h.renderFormError(w, r, ctxSummary, title, description, steps, severity,
			"La description est trop longue.")
		return
	}
	if len(steps) > maxStepsLen {
		h.renderFormError(w, r, ctxSummary, title, description, steps, severity,
			"Les étapes de reproduction sont trop longues.")
		return
	}

	now := time.Now().UTC()
	report := Report{
		ID:              uuid.NewString(),
		CreatedAt:       now.Format(time.RFC3339),
		Title:           title,
		Description:     description,
		Steps:           steps,
		Severity:        severity,
		Source:          sourceForm,
		PageURL:         pageURL,
		UserID:          user.ID,
		UserLogin:       user.Login,
		UserEmail:       user.Email,
		UserDisplayName: user.DisplayName,
		UserRole:        user.Role,
		OrgID:           ctxSummary.OrgID,
		OrgName:         ctxSummary.OrgName,
		OrgRole:         ctxSummary.OrgRole,
		UIRunLabel:      ctxSummary.UIRunLabel,
		SimpleUI:        ctxSummary.SimpleUI,
		UICaps:          ctxSummary.UICapsMap(),
		UserAgent:       r.UserAgent(),
		RequestID:       chimw.GetReqID(r.Context()),
	}

	path, err := AppendReport(h.ReportsDir, report)
	if err != nil {
		slog.Error("persist bug report", "err", err, "request_id", report.RequestID)
		h.renderFormError(w, r, ctxSummary, title, description, steps, severity,
			"Impossible d'enregistrer le signalement. Réessayez plus tard.")
		return
	}

	slog.Info("bug report submitted",
		"report_id", report.ID,
		"path", path,
		"user_id", report.UserID,
		"severity", report.Severity,
		"page_url", report.PageURL,
		"org_id", report.OrgID,
		"simple_ui", report.SimpleUI,
		"request_id", report.RequestID,
	)

	redirect := formPath + "?msg=" + url.QueryEscape("Signalement enregistré. Merci !")
	if ret := safeReturnPath(pageURL); ret != "" && ret != formPath {
		redirect += "&from=" + url.QueryEscape(ret)
	}
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

func (h *BugReports) renderFormError(
	w http.ResponseWriter,
	r *http.Request,
	ctx viewtemplates.BugReportContext,
	title, description, steps, severity, errMsg string,
) {
	data := viewtemplates.BugReportData{
		PageData:    viewtemplates.ApplyPageMeta(h.PageData(r, "Signaler un problème"), viewtemplates.BCBugReport()),
		Context:     ctx,
		PageURL:     ctx.PageURL,
		TitleValue:  title,
		Description: description,
		Steps:       steps,
		Severity:    severity,
		Error:       errMsg,
		ReturnURL:   safeReturnPath(ctx.PageURL),
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	if err := h.Templates.ExecuteTemplate(w, "bug_report", data); err != nil {
		slog.Error("render bug report form error", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func buildContextSummary(r *http.Request, user *store.User) viewtemplates.BugReportContext {
	ctx := viewtemplates.BugReportContext{
		UserID:          user.ID,
		UserLogin:       user.Login,
		UserEmail:       user.Email,
		UserDisplayName: user.DisplayName,
		UserRole:        user.Role,
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
		UserAgent:       r.UserAgent(),
		RequestID:       chimw.GetReqID(r.Context()),
	}

	hd, ok := middleware.HeaderDataFromContext(r.Context())
	if ok {
		ctx.SimpleUI = hd.SimpleUI
		ctx.ShowAssign = hd.ShowAssign
		ctx.ShowMyTasks = hd.ShowMyTasks
		ctx.ShowSubjectColumn = hd.ShowSubjectColumn
		ctx.ShowCollab = hd.ShowCollab
		if hd.ActiveOrg != nil {
			ctx.OrgID = hd.ActiveOrg.ID
			ctx.OrgName = hd.ActiveOrg.Name
			ctx.UIRunLabel = hd.ActiveOrg.UIRunLabel
			for _, m := range hd.UserOrganizations {
				if m.Organization.ID == hd.ActiveOrg.ID {
					ctx.OrgRole = m.Role
					break
				}
			}
		}
	} else if org, orgOK := middleware.OrganizationFromContext(r.Context()); orgOK {
		ctx.OrgID = org.ID
		ctx.OrgName = org.Name
		ctx.UIRunLabel = org.UIRunLabel
	}

	return ctx
}

func normalizeSeverity(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case severityLow, severityHigh:
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return severityNormal
	}
}

// resolvePageURL prefers an explicit client/from value, then Referer.
// Only relative app paths are kept (anti-tampering / open-redirect hygiene).
func resolvePageURL(candidates ...string) string {
	for _, c := range candidates {
		if p := safeReturnPath(c); p != "" {
			return p
		}
	}
	return ""
}

func safeReturnPath(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > maxPageURLLen {
		return ""
	}
	// Absolute URLs: keep path+query if parseable.
	if strings.Contains(raw, "://") {
		u, err := url.Parse(raw)
		if err != nil || u.Path == "" {
			return ""
		}
		raw = u.RequestURI()
	}
	if !strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "//") {
		return ""
	}
	if strings.ContainsAny(raw, "\r\n") {
		return ""
	}
	// Never bounce back to the report form itself as "return".
	if pathOnly := strings.SplitN(raw, "?", 2)[0]; pathOnly == formPath {
		return formPath
	}
	return raw
}

// ReportsDirFromAttachments derives data/bug-reports next to attachments.
func ReportsDirFromAttachments(attachmentsDir string) string {
	if attachmentsDir == "" {
		return "data/bug-reports"
	}
	return filepath.Join(filepath.Dir(attachmentsDir), "bug-reports")
}
