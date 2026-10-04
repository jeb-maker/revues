package apiv1

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/jeb-maker/revues/internal/auth"
	authfeature "github.com/jeb-maker/revues/internal/features/auth"
	"github.com/jeb-maker/revues/internal/store"
	appmiddleware "github.com/jeb-maker/revues/internal/web/middleware"
)

const maxAuthBodyBytes = 1 << 20 // 1 MiB

// GetBootstrap serves GET /api/v1/bootstrap.
func (s *Server) GetBootstrap(w http.ResponseWriter, r *http.Request) {
	oauthEnabled := s.Auth != nil && s.Auth.GitHubOAuthConfigured()

	if user, ok := appmiddleware.UserFromContext(r.Context()); ok {
		token := appmiddleware.SessionTokenFromContext(r)
		csrf := auth.CSRFToken(token, s.Config.SessionSecret)
		u := mapUser(user)
		redirect := "/runs"
		if s.Auth != nil {
			_, route, err := authfeature.PostLoginRoute(r.Context(), s.Auth.Store, user.ID)
			if err == nil && route != "" {
				redirect = route
			}
			// Prefer runs hub when session already has an active org.
			if token != "" && s.Orgs != nil {
				if list, listErr := s.Orgs.List(r.Context(), user, token, r); listErr == nil && list.ActiveOrganizationID > 0 {
					redirect = "/runs"
				}
			}
		}
		writeJSON(w, http.StatusOK, BootstrapResponse{
			Authenticated:      true,
			CsrfToken:          csrf,
			GithubOauthEnabled: oauthEnabled,
			User:               &u,
			Redirect:           &redirect,
		})
		return
	}

	_, csrf, err := s.Sessions.EnsureGuestToken(w, r)
	if err != nil {
		slog.Error("bootstrap guest csrf", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	writeJSON(w, http.StatusOK, BootstrapResponse{
		Authenticated:      false,
		CsrfToken:          csrf,
		GithubOauthEnabled: oauthEnabled,
	})
}

// GetMe serves GET /api/v1/me.
func (s *Server) GetMe(w http.ResponseWriter, r *http.Request) {
	user, ok := appmiddleware.UserFromContext(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthenticated", "Authentification requise.")
		return
	}
	token := appmiddleware.SessionTokenFromContext(r)
	writeJSON(w, http.StatusOK, MeResponse{
		User:      mapUser(user),
		CsrfToken: auth.CSRFToken(token, s.Config.SessionSecret),
	})
}

// PostAuthLogin serves POST /api/v1/auth/login.
func (s *Server) PostAuthLogin(w http.ResponseWriter, r *http.Request) {
	if _, ok := appmiddleware.UserFromContext(r.Context()); ok {
		writeAPIError(w, http.StatusBadRequest, "already_authenticated", "Déjà connecté.")
		return
	}

	var raw struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSONBody(r, &raw); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
		return
	}

	email := strings.TrimSpace(raw.Email)
	password := raw.Password
	switch {
	case email == "" && password == "":
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Indiquez votre email et votre mot de passe.")
		return
	case email == "":
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Indiquez votre adresse email.")
		return
	case password == "":
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Indiquez votre mot de passe.")
		return
	case !strings.Contains(email, "@"):
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Adresse email invalide.")
		return
	}

	result, err := s.Auth.PasswordLogin(r.Context(), email, password)
	if err != nil {
		if errors.Is(err, authfeature.ErrInvalidCredentials) {
			writeAPIError(w, http.StatusUnauthorized, "invalid_credentials", "Email ou mot de passe incorrect.")
			return
		}
		slog.Error("password login", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	s.Sessions.ClearGuestCookie(w)
	s.Sessions.SetSessionCookie(w, result.SessionRaw)
	writeJSON(w, http.StatusOK, AuthSuccessResponse{
		User:      mapUser(result.User),
		CsrfToken: result.CSRFToken,
		Redirect:  result.Redirect,
	})
}

// PostAuthRegister serves POST /api/v1/auth/register.
func (s *Server) PostAuthRegister(w http.ResponseWriter, r *http.Request) {
	if _, ok := appmiddleware.UserFromContext(r.Context()); ok {
		writeAPIError(w, http.StatusBadRequest, "already_authenticated", "Déjà connecté.")
		return
	}

	var req RegisterRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_failed", "Requête invalide.")
		return
	}

	result, err := s.Auth.Register(r.Context(), string(req.Email), req.Password, req.PasswordConfirm)
	if err != nil {
		switch {
		case errors.Is(err, authfeature.ErrValidation):
			writeAPIError(w, http.StatusBadRequest, "validation_failed", validationMessage(err))
		case errors.Is(err, authfeature.ErrEmailNotAllowed):
			writeAPIError(w, http.StatusBadRequest, "registration_failed",
				"Impossible de créer ce compte. Contactez un administrateur si besoin.")
		default:
			slog.Error("register", "err", err)
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		}
		return
	}

	s.Sessions.ClearGuestCookie(w)
	s.Sessions.SetSessionCookie(w, result.SessionRaw)
	writeJSON(w, http.StatusCreated, AuthSuccessResponse{
		User:      mapUser(result.User),
		CsrfToken: result.CSRFToken,
		Redirect:  result.Redirect,
	})
}

// PostAuthLogout serves POST /api/v1/auth/logout.
func (s *Server) PostAuthLogout(w http.ResponseWriter, r *http.Request) {
	token := appmiddleware.SessionTokenFromContext(r)
	if token == "" {
		writeAPIError(w, http.StatusUnauthorized, "unauthenticated", "Authentification requise.")
		return
	}

	if err := s.Sessions.ClearSession(r.Context(), token); err != nil {
		slog.Error("clear session", "err", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "Erreur interne.")
		return
	}

	s.Sessions.ClearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func mapUser(u *store.User) User {
	out := User{
		Id:          u.ID,
		Login:       u.Login,
		Email:       openapi_types.Email(u.Email),
		DisplayName: u.DisplayName,
		Role:        u.Role,
	}
	if u.AvatarURL != "" {
		avatar := u.AvatarURL
		out.AvatarUrl = &avatar
	}
	return out
}

func decodeJSONBody(r *http.Request, dst any) error {
	defer func() { _ = r.Body.Close() }()
	dec := json.NewDecoder(io.LimitReader(r.Body, maxAuthBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}

func validationMessage(err error) string {
	msg := err.Error()
	switch {
	case errors.Is(err, authfeature.ErrValidation) && strings.Contains(msg, "invalid email"):
		return "Adresse email invalide."
	case errors.Is(err, authfeature.ErrValidation) && strings.Contains(msg, "password mismatch"):
		return "Les mots de passe ne correspondent pas."
	case errors.Is(err, authfeature.ErrValidation) && strings.Contains(msg, "weak password"):
		return "Le mot de passe doit contenir entre 12 et 72 caractères."
	default:
		if i := strings.Index(msg, ": "); i >= 0 && i+2 < len(msg) {
			tail := strings.TrimSpace(msg[i+2:])
			if tail != "" && !strings.Contains(tail, "\n") {
				return tail
			}
		}
		return "Requête invalide."
	}
}
