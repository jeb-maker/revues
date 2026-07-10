-- +goose Up
CREATE TABLE organizations (
    id              INTEGER PRIMARY KEY,
    name            TEXT NOT NULL,
    slug            TEXT NOT NULL UNIQUE,
    created_at      TEXT NOT NULL,
    created_by      INTEGER REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE organization_members (
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id         INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role            TEXT NOT NULL
                    CHECK (role IN ('owner', 'admin', 'member')),
    created_at      TEXT NOT NULL,
    PRIMARY KEY (organization_id, user_id)
);

CREATE INDEX idx_organization_members_user ON organization_members(user_id);

INSERT INTO organizations (name, slug, created_at, created_by)
VALUES ('Default', 'default', strftime('%Y-%m-%dT%H:%M:%SZ', 'now'), NULL);

INSERT INTO organization_members (organization_id, user_id, role, created_at)
SELECT
    (SELECT id FROM organizations WHERE slug = 'default'),
    u.id,
    CASE
        WHEN u.role = 'admin' AND u.id = (SELECT MIN(id) FROM users WHERE role = 'admin')
            THEN 'owner'
        ELSE 'member'
    END,
    strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
FROM users u;

-- +goose Down
DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS organizations;
