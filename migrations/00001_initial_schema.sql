-- +goose Up
-- Schéma initial complet (aligné sur docs/schema/canonical.sql).

CREATE TABLE users (
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

CREATE TABLE organizations (
    id              INTEGER PRIMARY KEY,
    name            TEXT NOT NULL,
    slug            TEXT NOT NULL UNIQUE,
    ui_subject_label TEXT NOT NULL DEFAULT 'sujet'
                    CHECK (ui_subject_label IN ('sujet', 'cible', 'entite', 'asset')),
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

CREATE TABLE organization_invitations (
    id              INTEGER PRIMARY KEY,
    email           TEXT NOT NULL,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    org_role        TEXT NOT NULL DEFAULT 'member'
                    CHECK (org_role IN ('owner', 'admin', 'member')),
    created_at      TEXT NOT NULL
);

CREATE INDEX idx_organization_invitations_email ON organization_invitations(email);

CREATE UNIQUE INDEX idx_organization_invitations_unique
    ON organization_invitations(organization_id, email);

CREATE TABLE organization_teams (
    id              INTEGER PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    slug            TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    created_at      TEXT NOT NULL,
    UNIQUE (organization_id, slug)
);

CREATE INDEX idx_organization_teams_org ON organization_teams(organization_id);

CREATE TABLE team_members (
    team_id     INTEGER NOT NULL REFERENCES organization_teams(id) ON DELETE CASCADE,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TEXT NOT NULL,
    PRIMARY KEY (team_id, user_id)
);

CREATE INDEX idx_team_members_user ON team_members(user_id);

CREATE TABLE allowed_emails (
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email           TEXT NOT NULL,
    role            TEXT NOT NULL DEFAULT 'reader'
                    CHECK (role IN ('admin', 'editor', 'reader')),
    created_at      TEXT NOT NULL,
    PRIMARY KEY (organization_id, email)
);

CREATE INDEX idx_allowed_emails_org ON allowed_emails(organization_id);

CREATE TABLE sessions (
    id              INTEGER PRIMARY KEY,
    token_hash      TEXT NOT NULL UNIQUE,
    user_id         INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id INTEGER REFERENCES organizations(id) ON DELETE CASCADE,
    expires_at      TEXT NOT NULL,
    created_at      TEXT NOT NULL
);

CREATE INDEX idx_sessions_user ON sessions(user_id);

CREATE INDEX idx_sessions_expires ON sessions(expires_at);

CREATE TABLE subjects (
    id              INTEGER PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    archived_at     TEXT,
    created_at      TEXT NOT NULL,
    updated_at      TEXT NOT NULL
);

CREATE INDEX idx_subjects_organization ON subjects(organization_id);

CREATE TABLE subject_members (
    subject_id  INTEGER NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role        TEXT NOT NULL
                CHECK (role IN ('lead', 'contributor', 'viewer')),
    created_at  TEXT NOT NULL,
    PRIMARY KEY (subject_id, user_id)
);

CREATE INDEX idx_subject_members_user ON subject_members(user_id);

CREATE TABLE team_subject_roles (
    team_id     INTEGER NOT NULL REFERENCES organization_teams(id) ON DELETE CASCADE,
    subject_id  INTEGER NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    role        TEXT NOT NULL
                CHECK (role IN ('lead', 'contributor', 'viewer')),
    granted_by  INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at  TEXT NOT NULL,
    PRIMARY KEY (team_id, subject_id)
);

CREATE INDEX idx_team_subject_roles_subject ON team_subject_roles(subject_id);

-- Étiquettes descriptives (filtrer, retrouver — jamais accès)

CREATE TABLE subject_tags (
    subject_id INTEGER NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    tag        TEXT NOT NULL,
    PRIMARY KEY (subject_id, tag)
);

-- Domaines de matching modèles ↔ sujet

CREATE TABLE subject_domains (
    subject_id INTEGER NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    tag        TEXT NOT NULL,
    PRIMARY KEY (subject_id, tag)
);

CREATE TABLE checklist_templates (
    id              INTEGER PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    archived_at     TEXT,
    created_at      TEXT NOT NULL
);

CREATE INDEX idx_checklist_templates_organization ON checklist_templates(organization_id);

CREATE TABLE template_domains (
    template_id INTEGER NOT NULL REFERENCES checklist_templates(id) ON DELETE CASCADE,
    tag         TEXT NOT NULL,
    PRIMARY KEY (template_id, tag)
);

CREATE TABLE template_versions (
    id              INTEGER PRIMARY KEY,
    template_id     INTEGER NOT NULL REFERENCES checklist_templates(id) ON DELETE RESTRICT,
    version         INTEGER NOT NULL,
    published_at    TEXT NOT NULL,
    created_by      INTEGER REFERENCES users(id) ON DELETE SET NULL,
    UNIQUE (template_id, version)
);

CREATE INDEX idx_template_versions_template ON template_versions(template_id);

-- section : titre de section en texte (KISS v1, pas de table séparée)

CREATE TABLE template_items (
    id              INTEGER PRIMARY KEY,
    version_id      INTEGER NOT NULL REFERENCES template_versions(id) ON DELETE CASCADE,
    section         TEXT NOT NULL DEFAULT '',
    position        INTEGER NOT NULL,
    label           TEXT NOT NULL,
    help_text       TEXT NOT NULL DEFAULT '',
    required        INTEGER NOT NULL DEFAULT 1 CHECK (required IN (0, 1))
);

CREATE INDEX idx_template_items_version ON template_items(version_id, position);

CREATE TABLE checklist_runs (
    id                  INTEGER PRIMARY KEY,
    subject_id          INTEGER NOT NULL REFERENCES subjects(id) ON DELETE RESTRICT,
    template_version_id INTEGER NOT NULL REFERENCES template_versions(id) ON DELETE RESTRICT,
    status              TEXT NOT NULL DEFAULT 'draft'
                        CHECK (status IN ('draft', 'in_progress', 'done', 'archived')),
    due_date            TEXT,  -- ISO 8601, optionnel
    closing_note        TEXT NOT NULL DEFAULT '',
    created_by          INTEGER REFERENCES users(id) ON DELETE SET NULL,
    started_at          TEXT,
    completed_at        TEXT,
    notion_url          TEXT NOT NULL DEFAULT '',
    created_at          TEXT NOT NULL
);

CREATE INDEX idx_runs_subject ON checklist_runs(subject_id, status);

CREATE INDEX idx_runs_due ON checklist_runs(due_date);

-- Snapshot immuable (structure) ; champs limités mutables

CREATE TABLE run_items (
    id                  INTEGER PRIMARY KEY,
    run_id              INTEGER NOT NULL REFERENCES checklist_runs(id) ON DELETE CASCADE,
    source_item_id      INTEGER REFERENCES template_items(id) ON DELETE SET NULL,
    section             TEXT NOT NULL DEFAULT '',
    position            INTEGER NOT NULL,
    label               TEXT NOT NULL,
    help_text           TEXT NOT NULL DEFAULT '',
    required            INTEGER NOT NULL DEFAULT 1 CHECK (required IN (0, 1)),
    status              TEXT NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending', 'ok', 'nok', 'na')),
    comment             TEXT NOT NULL DEFAULT '',
    assigned_to         INTEGER REFERENCES users(id) ON DELETE SET NULL,
    checked_by          INTEGER REFERENCES users(id) ON DELETE SET NULL,
    checked_at          TEXT,
    updated_at          TEXT NOT NULL
);

CREATE INDEX idx_run_items_run ON run_items(run_id, position);

CREATE INDEX idx_run_items_assigned ON run_items(assigned_to, status);

CREATE TABLE run_item_events (
    id              INTEGER PRIMARY KEY,
    run_item_id     INTEGER NOT NULL REFERENCES run_items(id) ON DELETE CASCADE,
    user_id         INTEGER REFERENCES users(id) ON DELETE SET NULL,
    old_status      TEXT,
    new_status      TEXT NOT NULL,
    comment         TEXT NOT NULL DEFAULT '',
    created_at      TEXT NOT NULL
);

CREATE INDEX idx_run_item_events_item ON run_item_events(run_item_id, created_at);

CREATE TABLE integrations (
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

CREATE INDEX idx_integrations_organization ON integrations(organization_id);

CREATE TABLE integration_links (
    id              INTEGER PRIMARY KEY,
    run_item_id     INTEGER NOT NULL REFERENCES run_items(id) ON DELETE CASCADE,
    integration_id  INTEGER NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    external_key    TEXT NOT NULL DEFAULT '',
    external_url    TEXT NOT NULL DEFAULT '',
    created_at      TEXT NOT NULL
);

CREATE INDEX idx_integration_links_item ON integration_links(run_item_id);

CREATE TABLE settings (
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    key             TEXT NOT NULL,
    value_encrypted BLOB NOT NULL,
    updated_at      TEXT NOT NULL,
    PRIMARY KEY (organization_id, key)
);

CREATE TABLE attachments (
    id              INTEGER PRIMARY KEY,
    run_item_id     INTEGER NOT NULL UNIQUE REFERENCES run_items(id) ON DELETE CASCADE,
    filename        TEXT NOT NULL,
    mime_type       TEXT NOT NULL,
    size_bytes      INTEGER NOT NULL,
    storage_path    TEXT NOT NULL,
    created_at      TEXT NOT NULL
);

CREATE TABLE webhook_deliveries (
    id              INTEGER PRIMARY KEY,
    event_id        TEXT NOT NULL,
    event_type      TEXT NOT NULL,
    url             TEXT NOT NULL,
    status_code     INTEGER,
    success         INTEGER NOT NULL CHECK (success IN (0, 1)),
    created_at      TEXT NOT NULL
);

CREATE INDEX idx_webhook_deliveries_event ON webhook_deliveries(event_id);

-- Organisation par défaut (bootstrap dev/tests ; création org self-service au premier login).
INSERT INTO organizations (name, slug, created_at, created_by)
VALUES ('Default', 'default', strftime('%Y-%m-%dT%H:%M:%SZ', 'now'), NULL);

-- +goose Down
DROP TABLE IF EXISTS webhook_deliveries;
DROP TABLE IF EXISTS attachments;
DROP TABLE IF EXISTS settings;
DROP TABLE IF EXISTS integration_links;
DROP TABLE IF EXISTS integrations;
DROP TABLE IF EXISTS run_item_events;
DROP TABLE IF EXISTS run_items;
DROP TABLE IF EXISTS checklist_runs;
DROP TABLE IF EXISTS template_items;
DROP TABLE IF EXISTS template_versions;
DROP TABLE IF EXISTS template_domains;
DROP TABLE IF EXISTS checklist_templates;
DROP TABLE IF EXISTS subject_domains;
DROP TABLE IF EXISTS subject_tags;
DROP TABLE IF EXISTS team_subject_roles;
DROP TABLE IF EXISTS subject_members;
DROP TABLE IF EXISTS subjects;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS allowed_emails;
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS organization_teams;
DROP TABLE IF EXISTS organization_invitations;
DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS users;
