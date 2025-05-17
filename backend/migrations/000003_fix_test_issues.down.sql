-- Revoke permissions from test user
REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM "user";
REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public FROM "user";

-- Drop test user role
DROP ROLE IF EXISTS "user";

-- Remove default values
ALTER TABLE activities
    ALTER COLUMN user_id DROP DEFAULT,
    ALTER COLUMN metadata DROP DEFAULT,
    ALTER COLUMN visibility DROP DEFAULT;

ALTER TABLE lists
    ALTER COLUMN owner_id DROP DEFAULT,
    ALTER COLUMN metadata DROP DEFAULT,
    ALTER COLUMN visibility DROP DEFAULT;

ALTER TABLE tribes
    ALTER COLUMN metadata DROP DEFAULT,
    ALTER COLUMN type DROP DEFAULT,
    ALTER COLUMN visibility DROP DEFAULT; 