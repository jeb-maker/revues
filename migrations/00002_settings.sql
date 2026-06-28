-- +goose Up
CREATE TABLE settings (
    key             TEXT PRIMARY KEY,
    value_encrypted BLOB NOT NULL,
    updated_at      TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS settings;
