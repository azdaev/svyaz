-- +goose Up
ALTER TABLE projects ADD COLUMN slug TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX idx_projects_slug ON projects(slug) WHERE slug != '';

-- +goose Down
DROP INDEX IF EXISTS idx_projects_slug;
ALTER TABLE projects DROP COLUMN slug;
