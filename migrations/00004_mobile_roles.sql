-- +goose Up
INSERT INTO roles (slug, name) VALUES ('ios', 'iOS-разработчик');
INSERT INTO roles (slug, name) VALUES ('android', 'Android-разработчик');
INSERT INTO roles (slug, name) VALUES ('flutter', 'Flutter-разработчик');

-- +goose Down
DELETE FROM roles WHERE slug IN ('ios', 'android', 'flutter');
