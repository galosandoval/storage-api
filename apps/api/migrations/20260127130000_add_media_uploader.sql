-- +goose Up
ALTER TABLE storage_items
  ADD COLUMN uploader_id UUID REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN is_private BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX idx_storage_items_uploader_id ON storage_items(uploader_id);
CREATE INDEX idx_storage_items_is_private ON storage_items(is_private);

-- +goose Down
DROP INDEX IF EXISTS idx_storage_items_is_private;
DROP INDEX IF EXISTS idx_storage_items_uploader_id;
ALTER TABLE storage_items
  DROP COLUMN is_private,
  DROP COLUMN uploader_id;
