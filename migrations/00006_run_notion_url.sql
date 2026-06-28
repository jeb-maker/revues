-- +goose Up
ALTER TABLE checklist_runs ADD COLUMN notion_url TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE checklist_runs DROP COLUMN notion_url;
