-- +goose Up
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

CREATE TABLE sessions (
    id              INTEGER PRIMARY KEY,
    token_hash      TEXT NOT NULL UNIQUE,
    user_id         INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at      TEXT NOT NULL,
    created_at      TEXT NOT NULL
);

CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);

CREATE TABLE allowed_emails (
    email           TEXT PRIMARY KEY,
    role            TEXT NOT NULL DEFAULT 'reader'
                    CHECK (role IN ('admin', 'editor', 'reader')),
    created_at      TEXT NOT NULL
);

CREATE TABLE projects (
    id              INTEGER PRIMARY KEY,
    name            TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    archived_at     TEXT,
    created_at      TEXT NOT NULL,
    updated_at      TEXT NOT NULL
);

CREATE TABLE project_members (
    project_id      INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id         INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role            TEXT NOT NULL
                    CHECK (role IN ('lead', 'contributor', 'viewer')),
    created_at      TEXT NOT NULL,
    PRIMARY KEY (project_id, user_id)
);

CREATE INDEX idx_project_members_user ON project_members(user_id);

CREATE TABLE checklist_templates (
    id              INTEGER PRIMARY KEY,
    project_id      INTEGER REFERENCES projects(id) ON DELETE SET NULL,
    name            TEXT NOT NULL,
    archived_at     TEXT,
    created_at      TEXT NOT NULL
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
    project_id          INTEGER NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
    template_version_id INTEGER NOT NULL REFERENCES template_versions(id) ON DELETE RESTRICT,
    title               TEXT NOT NULL,
    status              TEXT NOT NULL DEFAULT 'draft'
                        CHECK (status IN ('draft', 'in_progress', 'done', 'archived')),
    due_date            TEXT,
    closing_note        TEXT NOT NULL DEFAULT '',
    created_by          INTEGER REFERENCES users(id) ON DELETE SET NULL,
    started_at          TEXT,
    completed_at        TEXT,
    created_at          TEXT NOT NULL
);

CREATE INDEX idx_runs_project ON checklist_runs(project_id, status);
CREATE INDEX idx_runs_due ON checklist_runs(due_date);

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

-- +goose Down
DROP TABLE IF EXISTS run_item_events;
DROP TABLE IF EXISTS run_items;
DROP TABLE IF EXISTS checklist_runs;
DROP TABLE IF EXISTS template_items;
DROP TABLE IF EXISTS template_versions;
DROP TABLE IF EXISTS checklist_templates;
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS allowed_emails;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
