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
	MemberRole       string
	CanManage        bool
	CanManageMembers bool
	Message          string
	Error            string
}

// Parse loads layout and page templates from the embedded filesystem.
func Parse() (*template.Template, error) {
	root, err := fs.Sub(webassets.Templates, "templates")
	if err != nil {
		return nil, fmt.Errorf("templates root: %w", err)
	}

	tpl, err := template.ParseFS(root,
		"layouts/base.html",
		"pages/*.html",
	)
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	return tpl, nil
}
