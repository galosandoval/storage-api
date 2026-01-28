-- +goose Up
ALTER TABLE users
  ADD COLUMN first_name TEXT,
  ADD COLUMN last_name TEXT,
  ADD COLUMN image_url TEXT;

-- +goose Down
ALTER TABLE users
  DROP COLUMN first_name,
  DROP COLUMN last_name,
  DROP COLUMN image_url;
