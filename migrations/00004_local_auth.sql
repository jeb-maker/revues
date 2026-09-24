-- +goose NO TRANSACTION
-- +goose Up
-- Comptes locaux (email + mot de passe) : github_id nullable, password_hash, email unique.

PRAGMA foreign_keys = OFF;

CREATE TABLE users_new (
    id              INTEGER PRIMARY KEY,
    github_id       INTEGER UNIQUE,
    login           TEXT NOT NULL,
    email           TEXT NOT NULL,
    display_name    TEXT NOT NULL DEFAULT '',
    avatar_url      TEXT NOT NULL DEFAULT '',
    role            TEXT NOT NULL DEFAULT 'reader'
                    CHECK (role IN ('admin', 'editor', 'reader')),
    password_hash   TEXT NOT NULL DEFAULT '',
    created_at      TEXT NOT NULL,
    last_login_at   TEXT
);

INSERT INTO users_new (
    id, github_id, login, email, display_name, avatar_url, role, password_hash, created_at, last_login_at
)
SELECT
    id, github_id, login, email, display_name, avatar_url, role, '', created_at, last_login_at
FROM users;

DROP TABLE users;
ALTER TABLE users_new RENAME TO users;

CREATE UNIQUE INDEX idx_users_email_lower ON users (lower(email));

PRAGMA foreign_keys = ON;

-- +goose Down
PRAGMA foreign_keys = OFF;

CREATE TABLE users_old (
    id              INTEGER PRIMARY KEY,
    github_id       INTEGER NOT NULL UNIQUE,
    login           TEXT NOT NULL,
    email           TEXT NOT NULL,
    display_name    TEXT NOT NULL DEFAULT '',
    avatar_url      TEXT NOT NULL DEFAULT '',
    role            TEXT NOT NULL DEFAULT 'reader'
                    CHECK (role IN ('admin', 'editor', 'reader')),
    created_at      TEXT NOT NULL,
    last_login_at   TEXT
);

INSERT INTO users_old (
    id, github_id, login, email, display_name, avatar_url, role, created_at, last_login_at
)
SELECT
    id,
    COALESCE(github_id, -id),
    login,
    email,
    display_name,
    avatar_url,
    role,
    created_at,
    last_login_at
FROM users
WHERE github_id IS NOT NULL OR password_hash = '';

DROP TABLE users;
ALTER TABLE users_old RENAME TO users;

PRAGMA foreign_keys = ON;
