-- +goose Up
ALTER TABLE project_roles ADD COLUMN count INTEGER NOT NULL DEFAULT 1;

-- +goose Down
ALTER TABLE project_roles DROP COLUMN count;
