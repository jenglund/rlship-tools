-- Create test user role
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'user') THEN
        CREATE ROLE "user" WITH LOGIN PASSWORD 'test_password';
    END IF;
END
$$;

-- Grant necessary permissions to test user
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO "user";
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO "user";

-- Fix NOT NULL constraints with defaults
ALTER TABLE activities
    ALTER COLUMN user_id SET DEFAULT gen_random_uuid(),
    ALTER COLUMN metadata SET DEFAULT '{}'::jsonb,
    ALTER COLUMN visibility SET DEFAULT 'private';

ALTER TABLE lists
    ALTER COLUMN owner_id SET DEFAULT gen_random_uuid(),
    ALTER COLUMN metadata SET DEFAULT '{}'::jsonb,
    ALTER COLUMN visibility SET DEFAULT 'private';

ALTER TABLE tribes
    ALTER COLUMN metadata SET DEFAULT '{}'::jsonb,
    ALTER COLUMN type SET DEFAULT 'custom',
    ALTER COLUMN visibility SET DEFAULT 'private';

-- Add comments for test setup
COMMENT ON ROLE "user" IS 'Test user role created in migration 000003_fix_test_issues.up.sql'; 