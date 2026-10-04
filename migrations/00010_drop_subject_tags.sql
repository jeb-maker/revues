-- +goose Up
DROP TABLE IF EXISTS subject_tags;

-- +goose Down
CREATE TABLE subject_tags (
    subject_id INTEGER NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    tag        TEXT NOT NULL,
    PRIMARY KEY (subject_id, tag)
);
