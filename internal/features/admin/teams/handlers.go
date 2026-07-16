package teams

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

// Deps holds dependencies for admin teams handlers.
type Deps struct {
	Templates     *template.Template
	Store         TeamStore
	SessionSecret string
}

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

// AdminTeams manages organization teams and their members.
type AdminTeams struct {
	Deps
}

// List shows teams of the active organization and a create form.
func (h *AdminTeams) List(w http.ResponseWriter, r *http.Request) {
	teams, err := h.Store.ListOrganizationTeams(r.Context())
	if err != nil {
		slog.Error("list organization teams", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := h.listData(r)
	data.Teams = teams
	data.Message = r.URL.Query().Get("msg")
	if org, ok := middleware.OrganizationFromContext(r.Context()); ok {
		data.OrganizationName = org.Name
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "admin_teams", data); err != nil {
		slog.Error("render admin teams", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Create inserts a new team in the active organization.
func (h *AdminTeams) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	slug := strings.TrimSpace(r.FormValue("slug"))
	description := strings.TrimSpace(r.FormValue("description"))
	if name == "" {
		h.renderListError(w, r, "Le nom de l'équipe est requis.")
		return
	}
	if slug == "" {
		slug = name
	}

	team, err := h.Store.CreateTeam(r.Context(), name, slug, description)
	if err != nil {
		if errors.Is(err, store.ErrInvalidOrganizationSlug) {
			h.renderListError(w, r, "Slug invalide (lettres, chiffres et tirets).")
			return
		}
		if errors.Is(err, ErrTeamSlugTaken) {
			h.renderListError(w, r, "Ce slug d'équipe est déjà utilisé.")
			return
		}
		slog.Error("create team", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/teams/"+strconv.FormatInt(team.ID, 10)+"?msg=%C3%89quipe+cr%C3%A9%C3%A9e", http.StatusSeeOther)
}

// Show displays a team and its members.
func (h *AdminTeams) Show(w http.ResponseWriter, r *http.Request) {
	teamID, err := parseTeamID(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	data, ok := h.loadDetail(w, r, teamID)
	if !ok {
		return
	}
	data.Message = r.URL.Query().Get("msg")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, "admin_team_detail", data); err != nil {
		slog.Error("render admin team detail", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// AddMember adds an organization member to a team.
func (h *AdminTeams) AddMember(w http.ResponseWriter, r *http.Request) {
	teamID, err := parseTeamID(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(strings.TrimSpace(r.FormValue("user_id")), 10, 64)
	if err != nil || userID <= 0 {
		h.renderDetailError(w, r, teamID, "Membre invalide.")
		return
	}

	org, ok := middleware.OrganizationFromContext(r.Context())
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	_, isMember, err := h.Store.OrganizationMemberRole(r.Context(), org.ID, userID)
	if err != nil {
		slog.Error("organization member role", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		h.renderDetailError(w, r, teamID, "L'utilisateur n'est pas membre de l'organisation.")
		return
	}

	if err := h.Store.AddTeamMember(r.Context(), teamID, userID); err != nil {
		if errors.Is(err, ErrTeamNotFound) {
			http.NotFound(w, r)
			return
		}
		slog.Error("add team member", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/teams/"+strconv.FormatInt(teamID, 10)+"?msg=Membre+ajout%C3%A9", http.StatusSeeOther)
}

// RemoveMember removes a user from a team.
func (h *AdminTeams) RemoveMember(w http.ResponseWriter, r *http.Request) {
	teamID, err := parseTeamID(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(strings.TrimSpace(r.FormValue("user_id")), 10, 64)
	if err != nil || userID <= 0 {
		h.renderDetailError(w, r, teamID, "Membre invalide.")
		return
	}

	if err := h.Store.RemoveTeamMember(r.Context(), teamID, userID); err != nil {
		if errors.Is(err, ErrTeamNotFound) {
			http.NotFound(w, r)
			return
		}
		if errors.Is(err, ErrTeamMemberNotFound) {
			h.renderDetailError(w, r, teamID, "Ce membre n'est pas dans l'équipe.")
			return
		}
		slog.Error("remove team member", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/teams/"+strconv.FormatInt(teamID, 10)+"?msg=Membre+retir%C3%A9", http.StatusSeeOther)
}

func (h *AdminTeams) listData(r *http.Request) templates.AdminTeamsData {
	data := templates.AdminTeamsData{
		PageData: templates.ApplyPageMeta(h.PageData(r, ""), templates.BCAdminTeams()),
	}
	data.ActiveTab = "org"
	data.AdminSection = "teams"
	return data
}

func (h *AdminTeams) detailData(r *http.Request, team *store.OrganizationTeam) templates.AdminTeamDetailData {
	data := templates.AdminTeamDetailData{
		PageData: templates.ApplyPageMeta(h.PageData(r, ""), templates.BCAdminTeam(team.Name)),
		Team:     *team,
	}
	data.ActiveTab = "org"
	data.AdminSection = "teams"
	return data
}

func (h *AdminTeams) renderListError(w http.ResponseWriter, r *http.Request, message string) {
	teams, err := h.Store.ListOrganizationTeams(r.Context())
	if err != nil {
		slog.Error("list organization teams", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := h.listData(r)
	data.Teams = teams
	data.Error = message
	if org, ok := middleware.OrganizationFromContext(r.Context()); ok {
		data.OrganizationName = org.Name
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	if err := h.Templates.ExecuteTemplate(w, "admin_teams", data); err != nil {
		slog.Error("render admin teams error", "err", err)
	}
}

func (h *AdminTeams) loadDetail(w http.ResponseWriter, r *http.Request, teamID int64) (templates.AdminTeamDetailData, bool) {
	team, err := h.Store.TeamByID(r.Context(), teamID)
	if err != nil {
		if errors.Is(err, ErrTeamNotFound) {
			http.NotFound(w, r)
			return templates.AdminTeamDetailData{}, false
		}
		slog.Error("team by id", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return templates.AdminTeamDetailData{}, false
	}

	members, err := h.Store.ListTeamMembers(r.Context(), teamID)
	if err != nil {
		slog.Error("list team members", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return templates.AdminTeamDetailData{}, false
	}

	orgMembers, err := h.Store.ListOrganizationMembers(r.Context())
	if err != nil {
		slog.Error("list organization members", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return templates.AdminTeamDetailData{}, false
	}

	onTeam := make(map[int64]struct{}, len(members))
	for _, m := range members {
		onTeam[m.UserID] = struct{}{}
	}
	var candidates []store.OrganizationMemberUser
	for _, m := range orgMembers {
		if _, ok := onTeam[m.UserID]; !ok {
			candidates = append(candidates, m)
		}
	}

	data := h.detailData(r, team)
	data.Members = members
	data.Candidates = candidates
	if org, ok := middleware.OrganizationFromContext(r.Context()); ok {
		data.OrganizationName = org.Name
	}
	return data, true
}

func (h *AdminTeams) renderDetailError(w http.ResponseWriter, r *http.Request, teamID int64, message string) {
	data, ok := h.loadDetail(w, r, teamID)
	if !ok {
		return
	}
	data.Error = message

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	if err := h.Templates.ExecuteTemplate(w, "admin_team_detail", data); err != nil {
		slog.Error("render admin team detail error", "err", err)
	}
}

func parseTeamID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
}
