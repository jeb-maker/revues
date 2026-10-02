-- name: UpsertGitHubUser :exec
INSERT INTO users (github_id, login, email, display_name, avatar_url, role, password_hash, created_at, last_login_at)
VALUES (?, ?, ?, ?, ?, ?, '', ?, ?)
ON CONFLICT(github_id) DO UPDATE SET
    login = excluded.login,
    email = excluded.email,
    display_name = excluded.display_name,
    avatar_url = excluded.avatar_url,
    role = excluded.role,
    last_login_at = excluded.last_login_at;

-- name: CreateLocalUser :execlastid
INSERT INTO users (github_id, login, email, display_name, avatar_url, role, password_hash, created_at, last_login_at)
VALUES (NULL, ?, ?, ?, '', ?, ?, ?, ?);

-- name: GetUserByID :one
SELECT id, github_id, login, email, display_name, avatar_url, role
FROM users
WHERE id = ?;

-- name: GetUserByEmail :one
SELECT id, github_id, login, email, display_name, avatar_url, role
FROM users
WHERE lower(email) = lower(?);

-- name: GetUserByGitHubID :one
SELECT id, github_id, login, email, display_name, avatar_url, role
FROM users
WHERE github_id = ?;

-- name: GetUserCredentialsByEmail :one
SELECT id, github_id, login, email, display_name, avatar_url, role, password_hash
FROM users
WHERE lower(email) = lower(?);

-- name: ListUsers :many
SELECT id, github_id, login, email, display_name, avatar_url, role
FROM users
ORDER BY lower(login) ASC, id ASC;

-- name: TouchLastLogin :exec
UPDATE users SET last_login_at = ? WHERE id = ?;
