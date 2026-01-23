-- +goose Up
ALTER TABLE storage_items
  ADD COLUMN preview_path TEXT,
  ADD COLUMN original_filename TEXT,
  ADD COLUMN camera_make TEXT,
  ADD COLUMN camera_model TEXT,
  ADD COLUMN latitude DOUBLE PRECISION,
  ADD COLUMN longitude DOUBLE PRECISION,
  ADD COLUMN orientation INT,
  ADD COLUMN iso INT,
  ADD COLUMN f_number DOUBLE PRECISION,
  ADD COLUMN exposure_time TEXT,
  ADD COLUMN focal_length DOUBLE PRECISION;

-- +goose Down
ALTER TABLE storage_items
  DROP COLUMN IF EXISTS preview_path,
  DROP COLUMN IF EXISTS original_filename,
  DROP COLUMN IF EXISTS camera_make,
  DROP COLUMN IF EXISTS camera_model,
  DROP COLUMN IF EXISTS latitude,
  DROP COLUMN IF EXISTS longitude,
  DROP COLUMN IF EXISTS orientation,
  DROP COLUMN IF EXISTS iso,
  DROP COLUMN IF EXISTS f_number,
  DROP COLUMN IF EXISTS exposure_time,
  DROP COLUMN IF EXISTS focal_length;
