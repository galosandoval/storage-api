-- +goose Up
-- Seed Sandoval Family household and users

-- Insert household
INSERT INTO households (id, name, created_at)
VALUES (
  '00000000-0000-0000-0000-000000000001',
  'Sandoval Family',
  now()
);

-- Insert Galo (admin)
INSERT INTO users (id, household_id, external_sub, email, role, created_at)
VALUES (
  '00000000-0000-0000-0000-000000000101',
  '00000000-0000-0000-0000-000000000001',
  'user_galo',
  'galosan92@gmail.com',
  'admin',
  now()
);

-- Insert Laura (member)
INSERT INTO users (id, household_id, external_sub, email, role, created_at)
VALUES (
  '00000000-0000-0000-0000-000000000102',
  '00000000-0000-0000-0000-000000000001',
  'user_laura',
  'laural.030590@gmail.com',
  'member',
  now()
);

-- +goose Down
-- Remove seed data
DELETE FROM users WHERE household_id = '00000000-0000-0000-0000-000000000001';
DELETE FROM households WHERE id = '00000000-0000-0000-0000-000000000001';
