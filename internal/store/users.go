package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Store wraps database access for Revues.
type Store struct {
	db *sql.DB
}

// ErrUserNotFound is returned when a user lookup fails.
var ErrUserNotFound = errors.New("user not found")

// ErrEmailTaken is returned when creating a local user with an existing email.
var ErrEmailTaken = errors.New("email already registered")

// New returns a Store backed by db.
func New(db *sql.DB) *Store {
	return &Store{db: db}
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

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO users (github_id, login, email, display_name, avatar_url, role, password_hash, created_at, last_login_at)
		VALUES (?, ?, ?, ?, ?, ?, '', ?, ?)
		ON CONFLICT(github_id) DO UPDATE SET
			login = excluded.login,
			email = excluded.email,
			display_name = excluded.display_name,
			avatar_url = excluded.avatar_url,
			role = excluded.role,
			last_login_at = excluded.last_login_at
	`, githubID, login, email, displayName, avatarURL, role, now, now)
	if err != nil {
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
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO users (github_id, login, email, display_name, avatar_url, role, password_hash, created_at, last_login_at)
		VALUES (NULL, ?, ?, ?, '', ?, ?, ?, ?)
	`, login, email, displayName, role, passwordHash, now, now)
	if err != nil {
		if isUniqueConstraint(err) {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("create local user: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("create local user id: %w", err)
	}
	return s.UserByID(ctx, id)
}

// UserCredentialsByEmail loads a user and password hash by email (case-insensitive).
func (s *Store) UserCredentialsByEmail(ctx context.Context, email string) (*User, string, error) {
	var user User
	var githubID sql.NullInt64
	var passwordHash string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, github_id, login, email, display_name, avatar_url, role, password_hash
		FROM users WHERE lower(email) = lower(?)
	`, email).Scan(
		&user.ID, &githubID, &user.Login, &user.Email, &user.DisplayName, &user.AvatarURL, &user.Role, &passwordHash,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, "", ErrUserNotFound
	}
	if err != nil {
		return nil, "", fmt.Errorf("user credentials by email: %w", err)
	}
	if githubID.Valid {
		user.GitHubID = githubID.Int64
	}
	return &user, passwordHash, nil
}

// TouchLastLogin updates last_login_at for the user.
func (s *Store) TouchLastLogin(ctx context.Context, userID int64) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `UPDATE users SET last_login_at = ? WHERE id = ?`, now, userID)
	if err != nil {
		return fmt.Errorf("touch last login: %w", err)
	}
	return nil
}

// UserByEmail loads a user by email address.
func (s *Store) UserByEmail(ctx context.Context, email string) (*User, error) {
	user, err := s.scanUserRow(s.db.QueryRowContext(ctx, `
		SELECT id, github_id, login, email, display_name, avatar_url, role
		FROM users WHERE lower(email) = lower(?)
	`, email))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user by email: %w", err)
	}
	return user, nil
}

// UserByID loads a user by primary key.
func (s *Store) UserByID(ctx context.Context, id int64) (*User, error) {
	user, err := s.scanUserRow(s.db.QueryRowContext(ctx, `
		SELECT id, github_id, login, email, display_name, avatar_url, role
		FROM users WHERE id = ?
	`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user by id: %w", err)
	}
	return user, nil
}

// ListUsers returns all users ordered by login (dev switcher / admin tooling).
func (s *Store) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, github_id, login, email, display_name, avatar_url, role
		FROM users
		ORDER BY lower(login) ASC, id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		user, scanErr := s.scanUser(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		users = append(users, *user)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("list users rows: %w", err)
	}
	return users, nil
}

func (s *Store) userByGitHubID(ctx context.Context, githubID int64) (*User, error) {
	user, err := s.scanUserRow(s.db.QueryRowContext(ctx, `
		SELECT id, github_id, login, email, display_name, avatar_url, role
		FROM users WHERE github_id = ?
	`, githubID))
	if err != nil {
		return nil, fmt.Errorf("load user after upsert: %w", err)
	}
	return user, nil
}

type scannable interface {
	Scan(dest ...any) error
}

func (s *Store) scanUserRow(row scannable) (*User, error) {
	return s.scanUser(row)
}

func (s *Store) scanUser(row scannable) (*User, error) {
	var user User
	var githubID sql.NullInt64
	if err := row.Scan(&user.ID, &githubID, &user.Login, &user.Email, &user.DisplayName, &user.AvatarURL, &user.Role); err != nil {
		return nil, err
	}
	if githubID.Valid {
		user.GitHubID = githubID.Int64
	}
	return &user, nil
}

func isUniqueConstraint(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique constraint") || strings.Contains(msg, "constraint failed")
}
