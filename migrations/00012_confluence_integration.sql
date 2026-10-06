-- +goose NO TRANSACTION
-- +goose Up
-- Rebuild integrations CHECK to include confluence ; FK OFF hors transaction
-- car integration_links référence integrations.
PRAGMA foreign_keys = OFF;

CREATE TABLE integrations_new (
    id               INTEGER PRIMARY KEY,
    organization_id  INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    type             TEXT NOT NULL
                     CHECK (type IN ('jira', 'webhook', 'notion', 'smtp', 'confluence')),
    enabled          INTEGER NOT NULL DEFAULT 0 CHECK (enabled IN (0, 1)),
    config_encrypted BLOB NOT NULL,
    created_at       TEXT NOT NULL,
    updated_at       TEXT NOT NULL,
    UNIQUE (organization_id, type)
);

INSERT INTO integrations_new (
    id, organization_id, type, enabled, config_encrypted, created_at, updated_at
)
SELECT
    id, organization_id, type, enabled, config_encrypted, created_at, updated_at
FROM integrations;

DROP TABLE integrations;
ALTER TABLE integrations_new RENAME TO integrations;

CREATE INDEX idx_integrations_organization ON integrations(organization_id);

PRAGMA foreign_keys = ON;

ALTER TABLE checklist_runs ADD COLUMN confluence_url TEXT NOT NULL DEFAULT '';

-- +goose Down
PRAGMA foreign_keys = OFF;

CREATE TABLE integrations_old (
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

INSERT INTO integrations_old (
    id, organization_id, type, enabled, config_encrypted, created_at, updated_at
)
SELECT
    id, organization_id, type, enabled, config_encrypted, created_at, updated_at
FROM integrations
WHERE type != 'confluence';

DROP TABLE integrations;
ALTER TABLE integrations_old RENAME TO integrations;

CREATE INDEX idx_integrations_organization ON integrations(organization_id);

PRAGMA foreign_keys = ON;

ALTER TABLE checklist_runs DROP COLUMN confluence_url;
