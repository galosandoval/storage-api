-- +goose Up
ALTER TABLE storage_items ADD COLUMN web_path TEXT;

-- +goose Down
ALTER TABLE storage_items DROP COLUMN IF EXISTS web_path;
