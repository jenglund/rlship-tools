-- Fix foreign key constraints
ALTER TABLE activities
    DROP CONSTRAINT IF EXISTS activities_user_id_fkey,
    ADD CONSTRAINT activities_user_id_fkey
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE;

-- Fix enum values
ALTER TABLE activities
    ALTER COLUMN visibility TYPE visibility_type
    USING visibility::visibility_type;

ALTER TABLE tribes
    ALTER COLUMN type TYPE tribe_type
    USING CASE
        WHEN type IS NULL OR type = '' THEN 'custom'::tribe_type
        WHEN type = 'couple' THEN 'couple'::tribe_type
        WHEN type = 'polycule' THEN 'polycule'::tribe_type
        WHEN type = 'friends' THEN 'friends'::tribe_type
        WHEN type = 'family' THEN 'family'::tribe_type
        WHEN type = 'roommates' THEN 'roommates'::tribe_type
        WHEN type = 'coworkers' THEN 'coworkers'::tribe_type
        ELSE 'custom'::tribe_type
    END,
    ALTER COLUMN visibility TYPE visibility_type
    USING CASE
        WHEN visibility IS NULL OR visibility = '' THEN 'private'::visibility_type
        ELSE visibility::visibility_type
    END;

-- Fix test database issues
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'testdb') THEN
        CREATE DATABASE testdb;
    END IF;
END
$$;

-- Grant permissions on testdb
DO $$
BEGIN
    EXECUTE format('GRANT ALL PRIVILEGES ON DATABASE testdb TO %I', current_user);
    IF EXISTS (SELECT FROM pg_roles WHERE rolname = 'user') THEN
        EXECUTE format('GRANT ALL PRIVILEGES ON DATABASE testdb TO %I', 'user');
    END IF;
END
$$; 