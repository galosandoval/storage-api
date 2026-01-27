-- +goose Up
-- Backfill uploader_id for existing media items to the first user in the household
UPDATE storage_items si
SET uploader_id = (
  SELECT u.id 
  FROM users u 
  WHERE u.household_id = si.household_id 
  ORDER BY u.created_at ASC 
  LIMIT 1
)
WHERE si.uploader_id IS NULL;

-- +goose Down
-- No rollback needed - this is a data backfill
