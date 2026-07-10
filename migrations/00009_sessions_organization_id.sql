-- +goose Up
PRAGMA foreign_keys=OFF;

CREATE TABLE sessions_new (
    id              INTEGER PRIMARY KEY,
    token_hash      TEXT NOT NULL UNIQUE,
    user_id         INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    expires_at      TEXT NOT NULL,
    created_at      TEXT NOT NULL
);

INSERT INTO sessions_new (id, token_hash, user_id, organization_id, expires_at, created_at)
SELECT
    s.id,
    s.token_hash,
    s.user_id,
    COALESCE(
        (SELECT om.organization_id FROM organization_members om WHERE om.user_id = s.user_id ORDER BY om.organization_id LIMIT 1),
        (SELECT id FROM organizations WHERE slug = 'default')
    ),
    s.expires_at,
    s.created_at
FROM sessions s;

DROP TABLE sessions;
ALTER TABLE sessions_new RENAME TO sessions;

CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);

PRAGMA foreign_keys=ON;

-- +goose Down
PRAGMA foreign_keys=OFF;

CREATE TABLE sessions_old (
    id              INTEGER PRIMARY KEY,
    token_hash      TEXT NOT NULL UNIQUE,
    user_id         INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at      TEXT NOT NULL,
    created_at      TEXT NOT NULL
);

INSERT INTO sessions_old (id, token_hash, user_id, expires_at, created_at)
SELECT id, token_hash, user_id, expires_at, created_at
FROM sessions;

DROP TABLE sessions;
ALTER TABLE sessions_old RENAME TO sessions;

CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);

PRAGMA foreign_keys=ON;
