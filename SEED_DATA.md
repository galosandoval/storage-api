# Seed Data Reference

## Default Users

The seed migration (`migrations/20260108000000_seed_sandoval_family.sql`) creates:

### Household
- **ID**: `00000000-0000-0000-0000-000000000001`
- **Name**: Sandoval Family

### Users

| Name | Email | Role | External Sub | User ID |
|------|-------|------|--------------|---------|
| Galo | galosan92@gmail.com | admin | `user_galo` | `00000000-0000-0000-0000-000000000101` |
| Laura | laural.030590@gmail.com | member | `user_laura` | `00000000-0000-0000-0000-000000000102` |

## Running the Seed Migration

```bash
# Make sure PostgreSQL is running
docker-compose up -d

# Run migrations (includes seed)
./migrate.sh
```

## Updating External Sub Values

When you integrate OAuth (Auth0, Google, etc.), you'll need to update the `external_sub` values with real OAuth provider IDs.

### Get Real External Sub from OAuth Provider

After a user logs in via OAuth, you'll receive a `sub` claim (subject identifier). For example:
- **Auth0**: `auth0|1234567890abcdef`
- **Google**: `google-oauth2|1234567890`
- **Clerk**: `user_2abc123def456`

### Update User with Real External Sub

```sql
-- Update Galo's external_sub
UPDATE users
SET external_sub = 'auth0|YOUR_ACTUAL_SUB_HERE'
WHERE email = 'galosan92@gmail.com';

-- Update Laura's external_sub
UPDATE users
SET external_sub = 'auth0|HER_ACTUAL_SUB_HERE'
WHERE email = 'laural.030590@gmail.com';
```

Or via `psql`:

```bash
# Connect to database
docker exec -it storage-postgres psql -U storageapp -d storage_db

# Run update queries
UPDATE users SET external_sub = 'auth0|YOUR_SUB' WHERE email = 'galosan92@gmail.com';
UPDATE users SET external_sub = 'auth0|HER_SUB' WHERE email = 'laural.030590@gmail.com';

# Verify
SELECT email, external_sub, role FROM users;

# Exit
\q
```

## Testing with Seed Data

You can test the API using the placeholder `external_sub` values:

```bash
# Test getting user by external_sub (if you have an endpoint for it)
curl http://localhost:8080/v1/users/user_galo
```

## Re-seeding (Development Only)

If you need to reset and re-seed:

```bash
# Rollback all migrations
goose -dir migrations postgres "$DATABASE_URL" reset

# Reapply all migrations (including seed)
./migrate.sh
```

⚠️ **Warning**: This will delete ALL data in the database!

## Production Notes

- Do NOT use these fixed UUIDs in production
- Replace placeholder `external_sub` values before deploying
- Consider using a separate seed script for production vs development
- Add more robust role-based access control as needed

## Adding More Users

To add additional family members, you can either:

### Option 1: Add to seed migration

Edit `migrations/20260108000000_seed_sandoval_family.sql` and add more INSERT statements.

### Option 2: Create new migration

```bash
# Create new migration
goose -dir migrations create add_user_john sql

# Edit the generated file to add the user
```

### Option 3: Insert via SQL

```sql
INSERT INTO users (household_id, external_sub, email, role)
VALUES (
  '00000000-0000-0000-0000-000000000001',
  'user_john',
  'john@example.com',
  'member'
);
```
