-- Drop triggers
DROP TRIGGER IF EXISTS increment_tribes_version ON tribes;
DROP TRIGGER IF EXISTS increment_tribe_members_version ON tribe_members;
DROP TRIGGER IF EXISTS increment_users_version ON users;
DROP TRIGGER IF EXISTS increment_lists_version ON lists;
DROP TRIGGER IF EXISTS increment_list_items_version ON list_items;
DROP TRIGGER IF EXISTS increment_list_sharing_version ON list_sharing;
DROP TRIGGER IF EXISTS increment_activities_version ON activities;
DROP TRIGGER IF EXISTS increment_activity_sharing_version ON activity_sharing;

-- Drop function
DROP FUNCTION IF EXISTS increment_version();

-- Drop version columns
ALTER TABLE tribes DROP COLUMN IF EXISTS version;
ALTER TABLE tribe_members DROP COLUMN IF EXISTS version;
ALTER TABLE users DROP COLUMN IF EXISTS version;
ALTER TABLE lists DROP COLUMN IF EXISTS version;
ALTER TABLE list_items DROP COLUMN IF EXISTS version;
ALTER TABLE list_sharing DROP COLUMN IF EXISTS version;
ALTER TABLE activities DROP COLUMN IF EXISTS version;
ALTER TABLE activity_sharing DROP COLUMN IF EXISTS version; 