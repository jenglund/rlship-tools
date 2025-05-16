-- Create enum types
CREATE TYPE visibility_type AS ENUM ('private', 'shared', 'public');
CREATE TYPE owner_type AS ENUM ('user', 'tribe');
CREATE TYPE list_type AS ENUM ('general', 'location', 'activity', 'interest', 'google_map');
CREATE TYPE list_sync_status AS ENUM ('none', 'synced', 'pending', 'conflict');
CREATE TYPE tribe_type AS ENUM ('couple', 'polycule', 'friends', 'family', 'roommates', 'coworkers', 'custom');
CREATE TYPE membership_type AS ENUM ('full', 'limited', 'guest');
CREATE TYPE auth_provider AS ENUM ('google', 'email', 'phone');
CREATE TYPE activity_type AS ENUM ('location', 'interest', 'list', 'activity');

-- Convert existing columns to use enum types
ALTER TABLE activities ALTER COLUMN type TYPE activity_type USING type::activity_type;
ALTER TABLE activities ALTER COLUMN visibility TYPE visibility_type USING visibility::visibility_type;

-- First remove the default value for sync_status
ALTER TABLE lists ALTER COLUMN sync_status DROP DEFAULT;

-- Convert list columns to enum types
ALTER TABLE lists ALTER COLUMN type TYPE list_type USING type::list_type;
ALTER TABLE lists ALTER COLUMN visibility TYPE visibility_type USING visibility::visibility_type;
ALTER TABLE lists ALTER COLUMN sync_status TYPE list_sync_status USING sync_status::list_sync_status;

-- Drop the CHECK constraint on owner_type before conversion
ALTER TABLE lists DROP CONSTRAINT IF EXISTS lists_owner_type_check;

-- Ensure owner_type values are valid before conversion
UPDATE lists SET owner_type = 'user' WHERE owner_type NOT IN ('user', 'tribe');
ALTER TABLE lists ALTER COLUMN owner_type TYPE owner_type USING owner_type::owner_type;

-- Add back the default value after conversion
ALTER TABLE lists ALTER COLUMN sync_status SET DEFAULT 'none'::list_sync_status;

ALTER TABLE tribes ALTER COLUMN type TYPE tribe_type USING type::tribe_type;
ALTER TABLE tribes ALTER COLUMN visibility TYPE visibility_type USING visibility::visibility_type;

ALTER TABLE tribe_members ALTER COLUMN membership_type TYPE membership_type USING membership_type::membership_type;

ALTER TABLE users ALTER COLUMN provider TYPE auth_provider USING provider::auth_provider;

-- Add NOT NULL constraints
ALTER TABLE activities ALTER COLUMN type SET NOT NULL;
ALTER TABLE activities ALTER COLUMN visibility SET NOT NULL;

ALTER TABLE lists ALTER COLUMN type SET NOT NULL;
ALTER TABLE lists ALTER COLUMN visibility SET NOT NULL;
ALTER TABLE lists ALTER COLUMN sync_status SET NOT NULL;
ALTER TABLE lists ALTER COLUMN owner_type SET NOT NULL;

ALTER TABLE tribes ALTER COLUMN type SET NOT NULL;
ALTER TABLE tribes ALTER COLUMN visibility SET NOT NULL;

ALTER TABLE tribe_members ALTER COLUMN membership_type SET NOT NULL;

ALTER TABLE users ALTER COLUMN provider SET NOT NULL; 