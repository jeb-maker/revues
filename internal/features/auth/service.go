package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"unicode"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/features/organizations"
	"github.com/jeb-maker/revues/internal/store"
)

// Service holds auth business logic shared by the JSON API and OAuth redirects.
type Service struct {
	Store    AuthStore
	Sessions *auth.SessionManager
	GitHub   *auth.GitHubOAuth
	Config   config.Config
}

// LoginResult is returned after a successful local or OAuth login.
type LoginResult struct {
	User       *store.User
	SessionRaw string
	CSRFToken  string
	Redirect   string
}

// PostLoginRoute decides how to seed the session organization and where to send the SPA.
// Delegates to organizations.PostLoginRoute (WP-010) for /org/new|/org/select|/ parity.
func PostLoginRoute(ctx context.Context, st interface {
	CountUserOrganizations(ctx context.Context, userID int64) (int, error)
	ListUserOrganizations(ctx context.Context, userID int64) ([]store.OrganizationMembership, error)
}, userID int64) (sessionOrgID int64, redirect string, err error) {
	return organizations.PostLoginRoute(ctx, st, userID)
}

// PasswordLogin authenticates email + password and creates a session.
func (s *Service) PasswordLogin(ctx context.Context, email, password string) (*LoginResult, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	user, passwordHash, err := s.Store.UserCredentialsByEmail(ctx, email)
	if err != nil || passwordHash == "" || !auth.VerifyPassword(passwordHash, password) {
		if errors.Is(err, store.ErrUserNotFound) || passwordHash == "" {
			_, _ = auth.HashPassword("timing-padding-password")
		}
		return nil, ErrInvalidCredentials
	}

	// Re-apply whitelist / bootstrap role on every login (same rules as GitHub).
	role, roleErr := s.Store.ResolveLoginRoleStrict(ctx, email, s.Config.BootstrapAdminEmail, s.Config.LoginRequireWhitelist)
	if roleErr != nil {
		if errors.Is(roleErr, store.ErrEmailNotAllowed) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("resolve login role: %w", roleErr)
	}
	if role != user.Role {
		if setErr := s.Store.SetUserRole(ctx, user.ID, role); setErr != nil {
			return nil, fmt.Errorf("sync login role: %w", setErr)
		}
		user.Role = role
	}

	return s.finishLocalLogin(ctx, user)
}

// Register creates a local account and session when the email is allowed.
// Display name is derived from the email local-part (no client input).
func (s *Service) Register(ctx context.Context, email, password, confirm string) (*LoginResult, error) {
	email = strings.TrimSpace(email)

	if _, err := mail.ParseAddress(email); err != nil || !strings.Contains(email, "@") {
		return nil, fmt.Errorf("%w: invalid email", ErrValidation)
	}
	email = strings.ToLower(email)

	if password != confirm {
		return nil, fmt.Errorf("%w: password mismatch", ErrValidation)
	}
	if err := auth.ValidatePasswordStrength(password); err != nil {
		return nil, fmt.Errorf("%w: weak password", ErrValidation)
	}

	role, err := s.Store.ResolveLoginRoleStrict(ctx, email, s.Config.BootstrapAdminEmail, s.Config.LoginRequireWhitelist)
	if err != nil {
		if errors.Is(err, store.ErrEmailNotAllowed) {
			return nil, ErrEmailNotAllowed
		}
		return nil, fmt.Errorf("resolve login role: %w", err)
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	login := localLoginFromEmail(email)
	displayName := displayNameFromEmail(email)
	user, err := s.Store.CreateLocalUser(ctx, email, login, displayName, role, passwordHash)
	if err != nil {
		if errors.Is(err, store.ErrEmailTaken) {
			return nil, ErrEmailNotAllowed // generic — do not reveal email taken
		}
		return nil, fmt.Errorf("create local user: %w", err)
	}

	if err = s.Store.EnsureBootstrapOrgOwner(ctx, user.ID, email, s.Config.BootstrapAdminEmail); err != nil {
		return nil, fmt.Errorf("bootstrap org owner: %w", err)
	}

	return s.finishLocalLogin(ctx, user)
}

func (s *Service) finishLocalLogin(ctx context.Context, user *store.User) (*LoginResult, error) {
	if err := s.Store.TouchLastLogin(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("touch last login: %w", err)
	}

	sessionOrgID, redirect, err := PostLoginRoute(ctx, s.Store, user.ID)
	if err != nil {
		return nil, err
	}

	sessionToken, csrf, err := s.Sessions.CreateLoginSession(ctx, user.ID, sessionOrgID)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &LoginResult{
		User:       user,
		SessionRaw: sessionToken,
		CSRFToken:  csrf,
		Redirect:   redirect,
	}, nil
}

// CompleteGitHubLogin upserts the GitHub user and opens a session.
func (s *Service) CompleteGitHubLogin(ctx context.Context, profile *auth.GitHubProfile) (*LoginResult, error) {
	role, err := s.Store.ResolveLoginRoleStrict(ctx, profile.Email, s.Config.BootstrapAdminEmail, s.Config.LoginRequireWhitelist)
	if err != nil {
		if errors.Is(err, store.ErrEmailNotAllowed) {
			return nil, ErrEmailNotAllowed
		}
		return nil, fmt.Errorf("resolve login role: %w", err)
	}

	displayName := profile.DisplayName
	if displayName == "" {
		displayName = profile.Login
	}

	user, err := s.Store.UpsertGitHubUser(
		ctx,
		profile.ID,
		profile.Login,
		profile.Email,
		displayName,
		profile.AvatarURL,
		role,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert github user: %w", err)
	}

	if err = s.Store.EnsureBootstrapOrgOwner(ctx, user.ID, profile.Email, s.Config.BootstrapAdminEmail); err != nil {
		return nil, fmt.Errorf("bootstrap org owner: %w", err)
	}

	sessionOrgID, redirect, err := PostLoginRoute(ctx, s.Store, user.ID)
	if err != nil {
		return nil, err
	}

	sessionToken, csrf, err := s.Sessions.CreateLoginSession(ctx, user.ID, sessionOrgID)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &LoginResult{
		User:       user,
		SessionRaw: sessionToken,
		CSRFToken:  csrf,
		Redirect:   redirect,
	}, nil
}

// GitHubOAuthConfigured reports whether OAuth client credentials are set.
func (s *Service) GitHubOAuthConfigured() bool {
	return s.GitHub != nil && s.GitHub.ClientID != "" && s.GitHub.ClientSecret != ""
}

func localLoginFromEmail(email string) string {
	local := email
	if at := strings.IndexByte(email, '@'); at > 0 {
		local = email[:at]
	}
	var b strings.Builder
	for _, r := range strings.ToLower(local) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if out == "" {
		return "user"
	}
	if len(out) > 64 {
		return out[:64]
	}
	return out
}

// displayNameFromEmail derives a display name from the email local-part.
func displayNameFromEmail(email string) string {
	local := email
	if at := strings.IndexByte(email, '@'); at > 0 {
		local = email[:at]
	}
	local = strings.TrimSpace(local)
	if local == "" {
		return "Utilisateur"
	}
	if len(local) > 80 {
		return local[:80]
	}
	return local
}

var (
	// ErrInvalidCredentials is returned for bad email/password pairs.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrValidation is returned for malformed register/login payloads.
	ErrValidation = errors.New("validation failed")
	// ErrEmailNotAllowed is returned when whitelist rejects an email (generic to clients).
	ErrEmailNotAllowed = errors.New("email not allowed")
)
