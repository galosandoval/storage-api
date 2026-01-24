-- +goose Up
-- Add thumbnail_path column for storing generated thumbnail paths
ALTER TABLE storage_items ADD COLUMN IF NOT EXISTS thumbnail_path VARCHAR(512);

-- +goose Down
ALTER TABLE storage_items DROP COLUMN IF EXISTS thumbnail_path;
