-- Remove NOT NULL constraints
ALTER TABLE activities ALTER COLUMN type DROP NOT NULL;
ALTER TABLE activities ALTER COLUMN visibility DROP NOT NULL;

ALTER TABLE lists ALTER COLUMN type DROP NOT NULL;
ALTER TABLE lists ALTER COLUMN visibility DROP NOT NULL;
ALTER TABLE lists ALTER COLUMN sync_status DROP NOT NULL;
ALTER TABLE lists ALTER COLUMN owner_type DROP NOT NULL;

ALTER TABLE tribes ALTER COLUMN type DROP NOT NULL;
ALTER TABLE tribes ALTER COLUMN visibility DROP NOT NULL;

ALTER TABLE tribe_members ALTER COLUMN membership_type DROP NOT NULL;

ALTER TABLE users ALTER COLUMN provider DROP NOT NULL;

-- Convert columns back to VARCHAR
ALTER TABLE activities ALTER COLUMN type TYPE VARCHAR(20) USING type::text;
ALTER TABLE activities ALTER COLUMN visibility TYPE VARCHAR(10) USING visibility::text;

-- First remove the default value for sync_status
ALTER TABLE lists ALTER COLUMN sync_status DROP DEFAULT;

-- Convert list columns back to VARCHAR
ALTER TABLE lists ALTER COLUMN type TYPE VARCHAR(20) USING type::text;
ALTER TABLE lists ALTER COLUMN visibility TYPE VARCHAR(10) USING visibility::text;
ALTER TABLE lists ALTER COLUMN sync_status TYPE VARCHAR(10) USING sync_status::text;
ALTER TABLE lists ALTER COLUMN owner_type TYPE VARCHAR(10) USING CASE 
    WHEN owner_type::text = 'user' THEN 'user'
    WHEN owner_type::text = 'tribe' THEN 'tribe'
    ELSE 'user'
END;

-- Add back the default value after conversion
ALTER TABLE lists ALTER COLUMN sync_status SET DEFAULT 'none';

ALTER TABLE tribes ALTER COLUMN type TYPE VARCHAR(20) USING type::text;
ALTER TABLE tribes ALTER COLUMN visibility TYPE VARCHAR(10) USING visibility::text;

ALTER TABLE tribe_members ALTER COLUMN membership_type TYPE VARCHAR(10) USING membership_type::text;

ALTER TABLE users ALTER COLUMN provider TYPE VARCHAR(10) USING provider::text;

-- Drop enum types
DROP TYPE IF EXISTS visibility_type;
DROP TYPE IF EXISTS owner_type;
DROP TYPE IF EXISTS list_type;
DROP TYPE IF EXISTS list_sync_status;
DROP TYPE IF EXISTS tribe_type;
DROP TYPE IF EXISTS membership_type;
DROP TYPE IF EXISTS auth_provider;
DROP TYPE IF EXISTS activity_type; 