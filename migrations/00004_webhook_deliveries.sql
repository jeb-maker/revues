-- +goose Up
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
-- +goose Down
DROP INDEX IF EXISTS idx_webhook_deliveries_event;
DROP TABLE IF EXISTS webhook_deliveries;
