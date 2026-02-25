-- +goose Up
ALTER TABLE responses ADD COLUMN role_id INTEGER REFERENCES roles(id);

-- +goose Down
ALTER TABLE responses DROP COLUMN role_id;
