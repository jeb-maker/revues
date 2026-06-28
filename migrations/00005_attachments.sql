-- +goose Up
CREATE TABLE attachments (
    id              INTEGER PRIMARY KEY,
    run_item_id     INTEGER NOT NULL UNIQUE REFERENCES run_items(id) ON DELETE CASCADE,
    filename        TEXT NOT NULL,
    mime_type       TEXT NOT NULL,
    size_bytes      INTEGER NOT NULL,
    storage_path    TEXT NOT NULL,
    created_at      TEXT NOT NULL
);
-- +goose Down
DROP TABLE IF EXISTS attachments;
