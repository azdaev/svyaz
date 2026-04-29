-- +goose Up
ALTER TABLE projects ADD COLUMN is_closed INTEGER NOT NULL DEFAULT 0;

-- +goose Down
-- SQLite does not support DROP COLUMN in older versions; no-op
