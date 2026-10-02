-- name: CreateSession :exec
INSERT INTO sessions (token_hash, user_id, organization_id, expires_at, created_at)
VALUES (?, ?, ?, ?, ?);

-- name: GetSessionByTokenHash :one
SELECT user_id, organization_id
FROM sessions
WHERE token_hash = ? AND expires_at > ?;

-- name: UpdateSessionOrganization :execrows
UPDATE sessions SET organization_id = ?
WHERE token_hash = ? AND expires_at > ?;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE token_hash = ?;

-- name: DeleteUserSessions :exec
DELETE FROM sessions WHERE user_id = ?;
