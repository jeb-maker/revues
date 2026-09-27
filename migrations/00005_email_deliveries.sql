-- +goose Up
CREATE TABLE email_deliveries (
    id               INTEGER PRIMARY KEY,
    organization_id  INTEGER REFERENCES organizations(id) ON DELETE CASCADE,
    to_address       TEXT NOT NULL,
    subject          TEXT NOT NULL,
    body             TEXT NOT NULL,
    attempts         INTEGER NOT NULL DEFAULT 0,
    next_attempt_at  TEXT,
    expires_at       TEXT,
    state            TEXT NOT NULL DEFAULT 'pending'
                       CHECK (state IN ('pending', 'done', 'poison')),
    last_error       TEXT,
    created_at       TEXT NOT NULL
);

CREATE INDEX idx_email_deliveries_drain
    ON email_deliveries (state, next_attempt_at);

CREATE INDEX idx_email_deliveries_org
    ON email_deliveries (organization_id);

-- +goose Down
DROP INDEX IF EXISTS idx_email_deliveries_org;
DROP INDEX IF EXISTS idx_email_deliveries_drain;
DROP TABLE IF EXISTS email_deliveries;
