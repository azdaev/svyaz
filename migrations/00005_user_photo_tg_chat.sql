-- +goose Up
ALTER TABLE users ADD COLUMN photo_url TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN tg_chat_id INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE users DROP COLUMN photo_url;
ALTER TABLE users DROP COLUMN tg_chat_id;
