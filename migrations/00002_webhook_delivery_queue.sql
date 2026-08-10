-- +goose Up
ALTER TABLE webhook_deliveries ADD COLUMN payload BLOB;
ALTER TABLE webhook_deliveries ADD COLUMN attempts INTEGER NOT NULL DEFAULT 0;
ALTER TABLE webhook_deliveries ADD COLUMN next_attempt_at TEXT;
ALTER TABLE webhook_deliveries ADD COLUMN expires_at TEXT;
ALTER TABLE webhook_deliveries ADD COLUMN state TEXT NOT NULL DEFAULT 'done';
ALTER TABLE webhook_deliveries ADD COLUMN last_error TEXT;

CREATE INDEX idx_webhook_deliveries_drain
    ON webhook_deliveries (state, next_attempt_at);

-- +goose Down
DROP INDEX IF EXISTS idx_webhook_deliveries_drain;
ALTER TABLE webhook_deliveries DROP COLUMN last_error;
ALTER TABLE webhook_deliveries DROP COLUMN state;
ALTER TABLE webhook_deliveries DROP COLUMN expires_at;
ALTER TABLE webhook_deliveries DROP COLUMN next_attempt_at;
ALTER TABLE webhook_deliveries DROP COLUMN attempts;
ALTER TABLE webhook_deliveries DROP COLUMN payload;
