package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jeb-maker/revues/internal/store/sqlc"
)

// Store wraps database access for Revues.
// Auth identity (users/sessions) goes through sqlc; other domains still use
// hand-written SQL in this package until follow-up migration issues.
type Store struct {
	db *sql.DB
	q  *sqlc.Queries
}

// ErrUserNotFound is returned when a user lookup fails.
var ErrUserNotFound = errors.New("user not found")

// ErrEmailTaken is returned when creating a local user with an existing email.
var ErrEmailTaken = errors.New("email already registered")

// New returns a Store backed by db.
func New(db *sql.DB) *Store {
	return &Store{db: db, q: sqlc.New(db)}
}

// Queries exposes the sqlc-generated query surface (rewrite / API handlers).
// Prefer Store methods for domain error mapping; use Queries for new typed access.
func (s *Store) Queries() *sqlc.Queries {
	return s.q
}

// User is an authenticated account.
type User struct {
	ID          int64
	GitHubID    int64 // 0 when the account has no linked GitHub identity
	Login       string
	Email       string
	DisplayName string
	AvatarURL   string
	Role        string
}

// UpsertGitHubUser inserts or updates a user from GitHub profile data.
func (s *Store) UpsertGitHubUser(ctx context.Context, githubID int64, login, email, displayName, avatarURL, role string) (*User, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	ghID := githubID
	if err := s.q.UpsertGitHubUser(ctx, sqlc.UpsertGitHubUserParams{
		GithubID:    &ghID,
		Login:       login,
		Email:       email,
		DisplayName: displayName,
		AvatarUrl:   avatarURL,
		Role:        role,
		CreatedAt:   now,
		LastLoginAt: &now,
	}); err != nil {
		return nil, fmt.Errorf("upsert user: %w", err)
	}

	return s.userByGitHubID(ctx, githubID)
}

// CreateLocalUser inserts an email/password account (no GitHub id).
func (s *Store) CreateLocalUser(ctx context.Context, email, login, displayName, role, passwordHash string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	login = strings.TrimSpace(login)
	displayName = strings.TrimSpace(displayName)
	if email == "" || login == "" || passwordHash == "" {
		return nil, fmt.Errorf("create local user: missing required fields")
	}

	now := time.Now().UTC().Format(time.RFC3339)
	id, err := s.q.CreateLocalUser(ctx, sqlc.CreateLocalUserParams{
		Login:        login,
		Email:        email,
		DisplayName:  displayName,
		Role:         role,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		LastLoginAt:  &now,
	})
	if err != nil {
		if isUniqueConstraint(err) {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("create local user: %w", err)
	}
	return s.UserByID(ctx, id)
}

// UserCredentialsByEmail loads a user and password hash by email (case-insensitive).
func (s *Store) UserCredentialsByEmail(ctx context.Context, email string) (*User, string, error) {
	row, err := s.q.GetUserCredentialsByEmail(ctx, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, "", ErrUserNotFound
	}
	if err != nil {
		return nil, "", fmt.Errorf("user credentials by email: %w", err)
	}
	return userFromCredentials(row), row.PasswordHash, nil
}

// TouchLastLogin updates last_login_at for the user.
func (s *Store) TouchLastLogin(ctx context.Context, userID int64) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if err := s.q.TouchLastLogin(ctx, sqlc.TouchLastLoginParams{
		LastLoginAt: &now,
		ID:          userID,
	}); err != nil {
		return fmt.Errorf("touch last login: %w", err)
	}
	return nil
}

// UserByEmail loads a user by email address.
func (s *Store) UserByEmail(ctx context.Context, email string) (*User, error) {
	row, err := s.q.GetUserByEmail(ctx, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user by email: %w", err)
	}
	return userFromEmailRow(row), nil
}

// UserByID loads a user by primary key.
func (s *Store) UserByID(ctx context.Context, id int64) (*User, error) {
	row, err := s.q.GetUserByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user by id: %w", err)
	}
	return userFromIDRow(row), nil
}

// ListUsers returns all users ordered by login (dev switcher / admin tooling).
func (s *Store) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := s.q.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	users := make([]User, 0, len(rows))
	for _, row := range rows {
		users = append(users, *userFromListRow(row))
	}
	return users, nil
}

func (s *Store) userByGitHubID(ctx context.Context, githubID int64) (*User, error) {
	ghID := githubID
	row, err := s.q.GetUserByGitHubID(ctx, &ghID)
	if err != nil {
		return nil, fmt.Errorf("load user after upsert: %w", err)
	}
	return userFromGitHubRow(row), nil
}

func githubIDValue(id *int64) int64 {
	if id == nil {
		return 0
	}
	return *id
}

func userFromIDRow(row sqlc.GetUserByIDRow) *User {
	return &User{
		ID: row.ID, GitHubID: githubIDValue(row.GithubID), Login: row.Login,
		Email: row.Email, DisplayName: row.DisplayName, AvatarURL: row.AvatarUrl, Role: row.Role,
	}
}

func userFromEmailRow(row sqlc.GetUserByEmailRow) *User {
	return &User{
		ID: row.ID, GitHubID: githubIDValue(row.GithubID), Login: row.Login,
		Email: row.Email, DisplayName: row.DisplayName, AvatarURL: row.AvatarUrl, Role: row.Role,
	}
}

func userFromGitHubRow(row sqlc.GetUserByGitHubIDRow) *User {
	return &User{
		ID: row.ID, GitHubID: githubIDValue(row.GithubID), Login: row.Login,
		Email: row.Email, DisplayName: row.DisplayName, AvatarURL: row.AvatarUrl, Role: row.Role,
	}
}

func userFromListRow(row sqlc.ListUsersRow) *User {
	return &User{
		ID: row.ID, GitHubID: githubIDValue(row.GithubID), Login: row.Login,
		Email: row.Email, DisplayName: row.DisplayName, AvatarURL: row.AvatarUrl, Role: row.Role,
	}
}

func userFromCredentials(row sqlc.GetUserCredentialsByEmailRow) *User {
	return &User{
		ID: row.ID, GitHubID: githubIDValue(row.GithubID), Login: row.Login,
		Email: row.Email, DisplayName: row.DisplayName, AvatarURL: row.AvatarUrl, Role: row.Role,
	}
}

func isUniqueConstraint(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique constraint") || strings.Contains(msg, "constraint failed")
}
