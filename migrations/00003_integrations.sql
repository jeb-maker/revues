-- +goose Up
CREATE TABLE integrations (
    id              INTEGER PRIMARY KEY,
    type            TEXT NOT NULL
                    CHECK (type IN ('jira', 'webhook', 'notion', 'smtp')),
    enabled         INTEGER NOT NULL DEFAULT 0 CHECK (enabled IN (0, 1)),
    config_encrypted BLOB NOT NULL,
    created_at      TEXT NOT NULL,
    updated_at      TEXT NOT NULL
);

CREATE TABLE integration_links (
    id              INTEGER PRIMARY KEY,
    run_item_id     INTEGER NOT NULL REFERENCES run_items(id) ON DELETE CASCADE,
    integration_id  INTEGER NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    external_key    TEXT NOT NULL DEFAULT '',
    external_url    TEXT NOT NULL DEFAULT '',
    created_at      TEXT NOT NULL
);

CREATE INDEX idx_integration_links_item ON integration_links(run_item_id);

-- +goose Down
DROP TABLE IF EXISTS integration_links;
DROP TABLE IF EXISTS integrations;
