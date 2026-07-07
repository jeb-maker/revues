package projects

import (
	"github.com/jeb-maker/revues/internal/store"
)

// Store wraps a *store.Store to expose project-related queries from the
// projects feature package. It embeds *store.Store so existing project SQL
// methods (CreateProject, ProjectByID, ListProjects, MemberRole, etc.) are
// promoted and accessible through the projects feature namespace.
//
// The underlying SQL stays in internal/store for now because it is shared by
// other features (runs, integrations, notifications) and the webhooks
// RunLoader interface. A dedicated issue will migrate the SQL into this
// package once the depending features are themselves extracted.
type Store struct {
	*store.Store
}

// New returns a projects Store backed by the given store.Store.
func New(s *store.Store) *Store {
	return &Store{Store: s}
}

// Project re-exports store.Project under the projects namespace.
type Project = store.Project

// ProjectMember re-exports store.ProjectMember under the projects namespace.
type ProjectMember = store.ProjectMember

// ErrProjectNotFound re-exports store.ErrProjectNotFound under the projects namespace.
var ErrProjectNotFound = store.ErrProjectNotFound
