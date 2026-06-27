package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Store wraps database access for Revues.
type Store struct {
	db *sql.DB
}

// ErrUserNotFound is returned when a user lookup fails.
var ErrUserNotFound = errors.New("user not found")

// New returns a Store backed by db.
func New(db *sql.DB) *Store {
	return &Store{db: db}
}

// User is an authenticated account.
type User struct {
	ID          int64
	GitHubID    int64
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
		INSERT INTO users (github_id, login, email, display_name, avatar_url, role, created_at, last_login_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
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

	var user User
	err = s.db.QueryRowContext(ctx, `
		SELECT id, github_id, login, email, display_name, avatar_url, role
		FROM users WHERE github_id = ?
	`, githubID).Scan(&user.ID, &user.GitHubID, &user.Login, &user.Email, &user.DisplayName, &user.AvatarURL, &user.Role)
	if err != nil {
		return nil, fmt.Errorf("load user after upsert: %w", err)
	}

	return &user, nil
}

// UserByEmail loads a user by email address.
func (s *Store) UserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := s.db.QueryRowContext(ctx, `
		SELECT id, github_id, login, email, display_name, avatar_url, role
		FROM users WHERE lower(email) = lower(?)
	`, email).Scan(&user.ID, &user.GitHubID, &user.Login, &user.Email, &user.DisplayName, &user.AvatarURL, &user.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user by email: %w", err)
	}
	return &user, nil
}

// UserByID loads a user by primary key.
func (s *Store) UserByID(ctx context.Context, id int64) (*User, error) {
	var user User
	err := s.db.QueryRowContext(ctx, `
		SELECT id, github_id, login, email, display_name, avatar_url, role
		FROM users WHERE id = ?
	`, id).Scan(&user.ID, &user.GitHubID, &user.Login, &user.Email, &user.DisplayName, &user.AvatarURL, &user.Role)
	if err != nil {
		return nil, fmt.Errorf("user by id: %w", err)
	}

	return &user, nil
}
