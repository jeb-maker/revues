-- +goose Up
DROP INDEX IF EXISTS idx_allowed_emails_org;
DROP TABLE IF EXISTS allowed_emails;

-- +goose Down
CREATE TABLE allowed_emails (
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email           TEXT NOT NULL,
    role            TEXT NOT NULL DEFAULT 'reader'
                    CHECK (role IN ('admin', 'editor', 'reader')),
    created_at      TEXT NOT NULL,
    PRIMARY KEY (organization_id, email)
);

CREATE INDEX idx_allowed_emails_org ON allowed_emails(organization_id);
