-- +goose NO TRANSACTION
-- +goose Up
-- Rebuild organizations (CHECK ui_subject_label) ; FK OFF hors transaction
-- sinon DROP échoue dès qu'il existe des runs (subjects ON DELETE RESTRICT).
PRAGMA foreign_keys = OFF;

CREATE TABLE organizations_new (
    id              INTEGER PRIMARY KEY,
    name            TEXT NOT NULL,
    slug            TEXT NOT NULL UNIQUE,
    ui_subject_label TEXT NOT NULL DEFAULT 'projet'
                    CHECK (ui_subject_label IN ('projet', 'sujet', 'cible', 'entite', 'asset')),
    ui_run_label TEXT NOT NULL DEFAULT 'revues'
                    CHECK (ui_run_label IN ('revues', 'listes_en_cours', 'audits', 'checklists')),
    leads_may_assign_teams     INTEGER NOT NULL DEFAULT 1 CHECK (leads_may_assign_teams IN (0, 1)),
    leads_may_invite_members   INTEGER NOT NULL DEFAULT 1 CHECK (leads_may_invite_members IN (0, 1)),
    leads_may_invite_externals INTEGER NOT NULL DEFAULT 0 CHECK (leads_may_invite_externals IN (0, 1)),
    created_at      TEXT NOT NULL,
    created_by      INTEGER REFERENCES users(id) ON DELETE SET NULL
);

INSERT INTO organizations_new (
    id, name, slug, ui_subject_label, ui_run_label,
    leads_may_assign_teams, leads_may_invite_members, leads_may_invite_externals,
    created_at, created_by
)
SELECT
    id, name, slug,
    CASE WHEN ui_subject_label = 'sujet' THEN 'projet' ELSE ui_subject_label END,
    ui_run_label,
    leads_may_assign_teams, leads_may_invite_members, leads_may_invite_externals,
    created_at, created_by
FROM organizations;

DROP TABLE organizations;
ALTER TABLE organizations_new RENAME TO organizations;

PRAGMA foreign_keys = ON;

-- +goose Down
PRAGMA foreign_keys = OFF;

CREATE TABLE organizations_old (
    id              INTEGER PRIMARY KEY,
    name            TEXT NOT NULL,
    slug            TEXT NOT NULL UNIQUE,
    ui_subject_label TEXT NOT NULL DEFAULT 'sujet'
                    CHECK (ui_subject_label IN ('sujet', 'cible', 'entite', 'asset')),
    ui_run_label TEXT NOT NULL DEFAULT 'revues'
                    CHECK (ui_run_label IN ('revues', 'listes_en_cours', 'audits', 'checklists')),
    leads_may_assign_teams     INTEGER NOT NULL DEFAULT 1 CHECK (leads_may_assign_teams IN (0, 1)),
    leads_may_invite_members   INTEGER NOT NULL DEFAULT 1 CHECK (leads_may_invite_members IN (0, 1)),
    leads_may_invite_externals INTEGER NOT NULL DEFAULT 0 CHECK (leads_may_invite_externals IN (0, 1)),
    created_at      TEXT NOT NULL,
    created_by      INTEGER REFERENCES users(id) ON DELETE SET NULL
);

INSERT INTO organizations_old (
    id, name, slug, ui_subject_label, ui_run_label,
    leads_may_assign_teams, leads_may_invite_members, leads_may_invite_externals,
    created_at, created_by
)
SELECT
    id, name, slug,
    CASE WHEN ui_subject_label = 'projet' THEN 'sujet' ELSE ui_subject_label END,
    ui_run_label,
    leads_may_assign_teams, leads_may_invite_members, leads_may_invite_externals,
    created_at, created_by
FROM organizations;

DROP TABLE organizations;
ALTER TABLE organizations_old RENAME TO organizations;

PRAGMA foreign_keys = ON;
