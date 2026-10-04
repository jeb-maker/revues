package apiv1

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/jeb-maker/revues/internal/store"
)

// GetSearch serves GET /api/v1/search.
func (s *Server) GetSearch(w http.ResponseWriter, r *http.Request, params GetSearchParams) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}
	if _, ok := orgFromRequest(r); !ok {
		writeAPIError(w, http.StatusForbidden, "org_required", "Organisation active requise.")
		return
	}

	q := strings.TrimSpace(params.Q)
	if q == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Paramètre q requis.")
		return
	}

	limit := store.DefaultSearchLimitPerKind
	if params.Limit != nil {
		limit = *params.Limit
	}

	orgRole, orgMember := s.orgMembership(r, user.ID)
	orgAdmin := orgMember && (orgRole == store.OrgRoleOwner || orgRole == store.OrgRoleAdmin)
	includeTemplates := orgMember
	// Assignation / Mes tâches désactivés produit — ne pas exposer kind=task (réactiver avec FEATURE_ASSIGN_TASKS).
	const includeTasks = false

	hits, totals, err := s.Store.GlobalSearch(r.Context(), store.GlobalSearchOpts{
		UserID:           user.ID,
		Admin:            orgAdmin,
		Query:            q,
		LimitPerKind:     limit,
		IncludeTemplates: includeTemplates,
		IncludeTasks:     includeTasks,
	})
	if err != nil {
		slog.Error("global search", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	out := make([]SearchResult, 0, len(hits))
	for _, hit := range hits {
		item := SearchResult{
			Kind:  SearchResultKind(hit.Kind),
			Id:    hit.ID,
			Title: hit.Title,
			Href:  hit.Href,
		}
		if hit.Subtitle != "" {
			sub := hit.Subtitle
			item.Subtitle = &sub
		}
		out = append(out, item)
	}
	writeJSON(w, http.StatusOK, SearchResponse{
		Results:     out,
		TotalByKind: totals,
	})
}
