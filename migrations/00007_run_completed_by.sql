-- +goose Up
ALTER TABLE checklist_runs ADD COLUMN completed_by INTEGER REFERENCES users(id) ON DELETE SET NULL;

-- +goose Down
-- SQLite < 3.35: no DROP COLUMN in older goose targets — leave column on down.
-- +goose StatementBegin
SELECT 1;
-- +goose StatementEnd
