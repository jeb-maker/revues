package bugreports

import (
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/jeb-maker/revues/internal/store"
	"github.com/jeb-maker/revues/internal/web/middleware"
)

const (
	formPath       = "/signaler"
	maxTitleLen    = 200
	maxDescLen     = 8000
	maxPageURLLen  = 2000
	severityLow    = "low"
	severityNormal = "normal"
	severityHigh   = "high"
)

// Deps holds shared dependencies for bug report handlers (API remount in later WP).
type Deps struct {
	SessionSecret string
	ReportsDir    string
}

// BugReports handles in-app bug report submission.
type BugReports struct {
	Deps
}

// ContextSummary is auto-captured request/session context for persistence.
type ContextSummary struct {
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
	HasJira           bool
	HasNotion         bool
	HasWebhooks       bool
	HasEvidence       bool
	Timestamp         string
	UserAgent         string
	RequestID         string
}

// UICapsMap returns progressive-disclosure flags for persistence.
func (c ContextSummary) UICapsMap() map[string]any {
	return map[string]any{
		"simple_ui":           c.SimpleUI,
		"show_assign":         c.ShowAssign,
		"show_my_tasks":       c.ShowMyTasks,
		"show_subject_column": c.ShowSubjectColumn,
		"show_collab":         c.ShowCollab,
		"has_jira":            c.HasJira,
		"has_notion":          c.HasNotion,
		"has_webhooks":        c.HasWebhooks,
		"has_evidence":        c.HasEvidence,
	}
}

func buildContextSummary(r *http.Request, user *store.User) ContextSummary {
	ctx := ContextSummary{
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
		ctx.HasJira = hd.HasJira
		ctx.HasNotion = hd.HasNotion
		ctx.HasWebhooks = hd.HasWebhooks
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
