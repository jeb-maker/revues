-- +goose Up
CREATE TABLE organization_invitations (
    id              INTEGER PRIMARY KEY,
    email           TEXT NOT NULL,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    project_id      INTEGER REFERENCES projects(id) ON DELETE CASCADE,
    project_role    TEXT CHECK (project_role IS NULL OR project_role IN ('lead', 'contributor', 'viewer')),
    org_role        TEXT NOT NULL DEFAULT 'member'
                    CHECK (org_role IN ('owner', 'admin', 'member')),
    created_at      TEXT NOT NULL
);

CREATE INDEX idx_organization_invitations_email ON organization_invitations(email);

CREATE UNIQUE INDEX idx_organization_invitations_unique
    ON organization_invitations(organization_id, email, IFNULL(project_id, 0));

-- +goose Down
DROP TABLE IF EXISTS organization_invitations;
