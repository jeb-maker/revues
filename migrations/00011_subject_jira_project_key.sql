-- +goose Up
ALTER TABLE subjects ADD COLUMN jira_project_key TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE subjects DROP COLUMN jira_project_key;
