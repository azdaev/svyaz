-- +goose Up
ALTER TABLE projects ADD COLUMN type TEXT NOT NULL DEFAULT 'non-commercial';

-- +goose Down
ALTER TABLE projects DROP COLUMN type;
