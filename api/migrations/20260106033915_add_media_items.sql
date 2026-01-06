-- +goose Up
CREATE TABLE media_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  household_id UUID NOT NULL REFERENCES households(id) ON DELETE CASCADE,

  path TEXT NOT NULL,
  type TEXT NOT NULL CHECK (type IN ('photo', 'video')),
  mime_type TEXT,
  size_bytes BIGINT,
  sha256 TEXT,

  taken_at TIMESTAMPTZ,
  width INT,
  height INT,
  duration_sec INT,

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  UNIQUE (household_id, path)
);

CREATE INDEX idx_media_items_household_id ON media_items(household_id);
CREATE INDEX idx_media_items_taken_at ON media_items(taken_at);

-- +goose Down
DROP TABLE IF EXISTS media_items;
