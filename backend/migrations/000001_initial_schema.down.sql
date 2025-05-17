-- Drop triggers first
DROP TRIGGER IF EXISTS increment_activity_sharing_version ON activity_sharing;
DROP TRIGGER IF EXISTS increment_activities_version ON activities;
DROP TRIGGER IF EXISTS increment_list_sharing_version ON list_sharing;
DROP TRIGGER IF EXISTS increment_list_items_version ON list_items;
DROP TRIGGER IF EXISTS increment_lists_version ON lists;
DROP TRIGGER IF EXISTS increment_tribe_members_version ON tribe_members;
DROP TRIGGER IF EXISTS increment_tribes_version ON tribes;
DROP TRIGGER IF EXISTS increment_users_version ON users;

DROP TRIGGER IF EXISTS update_activity_photos_updated_at ON activity_photos;
DROP TRIGGER IF EXISTS update_activity_sharing_updated_at ON activity_sharing;
DROP TRIGGER IF EXISTS update_activities_updated_at ON activities;
DROP TRIGGER IF EXISTS update_list_sharing_updated_at ON list_sharing;
DROP TRIGGER IF EXISTS update_list_items_updated_at ON list_items;
DROP TRIGGER IF EXISTS update_lists_updated_at ON lists;
DROP TRIGGER IF EXISTS update_tribe_members_updated_at ON tribe_members;
DROP TRIGGER IF EXISTS update_tribes_updated_at ON tribes;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop tables in correct order (respecting foreign key constraints)
DROP TABLE IF EXISTS activity_photos;
DROP TABLE IF EXISTS activity_sharing;
DROP TABLE IF EXISTS activities;
DROP TABLE IF EXISTS list_conflicts;
DROP TABLE IF EXISTS list_sharing;
DROP TABLE IF EXISTS list_items;
DROP TABLE IF EXISTS lists;
DROP TABLE IF EXISTS tribe_members;
DROP TABLE IF EXISTS tribes;
DROP TABLE IF EXISTS users;

-- Drop functions
DROP FUNCTION IF EXISTS increment_version();
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop enum types
DROP TYPE IF EXISTS sync_status_type;
DROP TYPE IF EXISTS sharing_type;
DROP TYPE IF EXISTS membership_type;
DROP TYPE IF EXISTS tribe_type;
DROP TYPE IF EXISTS visibility_type;
DROP TYPE IF EXISTS owner_type; 