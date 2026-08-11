-- +goose Up
ALTER TABLE webhook_deliveries ADD COLUMN organization_id INTEGER REFERENCES organizations(id) ON DELETE CASCADE;

UPDATE webhook_deliveries
SET organization_id = (SELECT id FROM organizations WHERE slug = 'default' LIMIT 1)
WHERE organization_id IS NULL;

CREATE INDEX idx_webhook_deliveries_org ON webhook_deliveries(organization_id);

-- +goose Down
DROP INDEX IF EXISTS idx_webhook_deliveries_org;
ALTER TABLE webhook_deliveries DROP COLUMN organization_id;
