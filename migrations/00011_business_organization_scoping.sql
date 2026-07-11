-- +goose Up
PRAGMA foreign_keys=OFF;

-- projects
CREATE TABLE projects_new (
    id              INTEGER PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    archived_at     TEXT,
    created_at      TEXT NOT NULL,
    updated_at      TEXT NOT NULL
);

INSERT INTO projects_new (id, organization_id, name, description, archived_at, created_at, updated_at)
SELECT
    p.id,
    (SELECT id FROM organizations WHERE slug = 'default'),
    p.name,
    p.description,
    p.archived_at,
    p.created_at,
    p.updated_at
FROM projects p;

DROP TABLE projects;
ALTER TABLE projects_new RENAME TO projects;

CREATE INDEX idx_projects_organization ON projects(organization_id);

-- allowed_emails (org-scoped whitelist)
CREATE TABLE allowed_emails_new (
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email           TEXT NOT NULL,
    role            TEXT NOT NULL DEFAULT 'reader'
                    CHECK (role IN ('admin', 'editor', 'reader')),
    created_at      TEXT NOT NULL,
    PRIMARY KEY (organization_id, email)
);

INSERT INTO allowed_emails_new (organization_id, email, role, created_at)
SELECT
    (SELECT id FROM organizations WHERE slug = 'default'),
    ae.email,
    ae.role,
    ae.created_at
FROM allowed_emails ae;

DROP TABLE allowed_emails;
ALTER TABLE allowed_emails_new RENAME TO allowed_emails;

CREATE INDEX idx_allowed_emails_org ON allowed_emails(organization_id);

-- settings (org-scoped key/value)
CREATE TABLE settings_new (
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    key             TEXT NOT NULL,
    value_encrypted BLOB NOT NULL,
    updated_at      TEXT NOT NULL,
    PRIMARY KEY (organization_id, key)
);

INSERT INTO settings_new (organization_id, key, value_encrypted, updated_at)
SELECT
    (SELECT id FROM organizations WHERE slug = 'default'),
    s.key,
    s.value_encrypted,
    s.updated_at
FROM settings s;

DROP TABLE settings;
ALTER TABLE settings_new RENAME TO settings;

-- integrations (one row per type per organization)
CREATE TABLE integrations_new (
    id               INTEGER PRIMARY KEY,
    organization_id  INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    type             TEXT NOT NULL
                     CHECK (type IN ('jira', 'webhook', 'notion', 'smtp')),
    enabled          INTEGER NOT NULL DEFAULT 0 CHECK (enabled IN (0, 1)),
    config_encrypted BLOB NOT NULL,
    created_at       TEXT NOT NULL,
    updated_at       TEXT NOT NULL,
    UNIQUE (organization_id, type)
);

INSERT INTO integrations_new (id, organization_id, type, enabled, config_encrypted, created_at, updated_at)
SELECT
    i.id,
    (SELECT id FROM organizations WHERE slug = 'default'),
    i.type,
    i.enabled,
    i.config_encrypted,
    i.created_at,
    i.updated_at
FROM integrations i;

DROP TABLE integrations;
ALTER TABLE integrations_new RENAME TO integrations;

CREATE INDEX idx_integrations_organization ON integrations(organization_id);

PRAGMA foreign_keys=ON;

-- +goose Down
PRAGMA foreign_keys=OFF;

CREATE TABLE projects_old (
    id              INTEGER PRIMARY KEY,
    name            TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    archived_at     TEXT,
    created_at      TEXT NOT NULL,
    updated_at      TEXT NOT NULL
);

INSERT INTO projects_old (id, name, description, archived_at, created_at, updated_at)
SELECT id, name, description, archived_at, created_at, updated_at
FROM projects;

DROP TABLE projects;
ALTER TABLE projects_old RENAME TO projects;

CREATE TABLE allowed_emails_old (
    email           TEXT PRIMARY KEY,
    role            TEXT NOT NULL DEFAULT 'reader'
                    CHECK (role IN ('admin', 'editor', 'reader')),
    created_at      TEXT NOT NULL
);

INSERT INTO allowed_emails_old (email, role, created_at)
SELECT email, role, created_at
FROM allowed_emails
WHERE organization_id = (SELECT id FROM organizations WHERE slug = 'default');

DROP TABLE allowed_emails;
ALTER TABLE allowed_emails_old RENAME TO allowed_emails;

CREATE TABLE settings_old (
    key             TEXT PRIMARY KEY,
    value_encrypted BLOB NOT NULL,
    updated_at      TEXT NOT NULL
);

INSERT INTO settings_old (key, value_encrypted, updated_at)
SELECT key, value_encrypted, updated_at
FROM settings
WHERE organization_id = (SELECT id FROM organizations WHERE slug = 'default');

DROP TABLE settings;
ALTER TABLE settings_old RENAME TO settings;

CREATE TABLE integrations_old (
    id              INTEGER PRIMARY KEY,
    type            TEXT NOT NULL
                    CHECK (type IN ('jira', 'webhook', 'notion', 'smtp')),
    enabled         INTEGER NOT NULL DEFAULT 0 CHECK (enabled IN (0, 1)),
    config_encrypted BLOB NOT NULL,
    created_at      TEXT NOT NULL,
    updated_at      TEXT NOT NULL
);

INSERT INTO integrations_old (id, type, enabled, config_encrypted, created_at, updated_at)
SELECT id, type, enabled, config_encrypted, created_at, updated_at
FROM integrations
WHERE organization_id = (SELECT id FROM organizations WHERE slug = 'default');

DROP TABLE integrations;
ALTER TABLE integrations_old RENAME TO integrations;

PRAGMA foreign_keys=ON;
