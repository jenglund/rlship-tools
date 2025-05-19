-- Drop triggers
DROP TRIGGER IF EXISTS update_activity_owners_updated_at ON activity_owners;
DROP TRIGGER IF EXISTS validate_activity_owner_trigger ON activity_owners;
DROP TRIGGER IF EXISTS validate_list_owner_trigger ON list_owners;
DROP TRIGGER IF EXISTS increment_list_sharing_version ON list_sharing;
DROP TRIGGER IF EXISTS update_list_sharing_updated_at ON list_sharing;
DROP TRIGGER IF EXISTS increment_list_items_version ON list_items;
DROP TRIGGER IF EXISTS update_list_items_updated_at ON list_items;
DROP TRIGGER IF EXISTS increment_lists_version ON lists;
DROP TRIGGER IF EXISTS update_lists_updated_at ON lists;
DROP TRIGGER IF EXISTS update_activity_shares_updated_at ON activity_shares;
DROP TRIGGER IF EXISTS increment_activity_photos_version ON activity_photos;
DROP TRIGGER IF EXISTS update_activity_photos_updated_at ON activity_photos;
DROP TRIGGER IF EXISTS increment_activities_version ON activities;
DROP TRIGGER IF EXISTS update_activities_updated_at ON activities;
DROP TRIGGER IF EXISTS increment_tribe_members_version ON tribe_members;
DROP TRIGGER IF EXISTS update_tribe_members_updated_at ON tribe_members;
DROP TRIGGER IF EXISTS increment_tribes_version ON tribes;
DROP TRIGGER IF EXISTS update_tribes_updated_at ON tribes;
DROP TRIGGER IF EXISTS increment_users_version ON users;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_list_conflicts_updated_at ON list_conflicts;

-- Drop tables
DROP TABLE IF EXISTS activity_owners CASCADE;
DROP TABLE IF EXISTS activity_photos CASCADE;
DROP TABLE IF EXISTS activity_shares CASCADE;
DROP TABLE IF EXISTS list_sharing CASCADE;
DROP TABLE IF EXISTS list_items CASCADE;
DROP TABLE IF EXISTS list_owners CASCADE;
DROP TABLE IF EXISTS sync_conflicts CASCADE;
DROP TABLE IF EXISTS list_conflicts CASCADE;
DROP TABLE IF EXISTS lists CASCADE;
DROP TABLE IF EXISTS activities CASCADE;
DROP TABLE IF EXISTS tribe_members CASCADE;
DROP TABLE IF EXISTS tribes CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Drop functions
DROP FUNCTION IF EXISTS validate_activity_owner() CASCADE;
DROP FUNCTION IF EXISTS validate_list_owner() CASCADE;
DROP FUNCTION IF EXISTS increment_version() CASCADE;
DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;

-- Drop enums
DROP TYPE IF EXISTS visibility_type CASCADE;
DROP TYPE IF EXISTS tribe_type CASCADE;
DROP TYPE IF EXISTS sync_status CASCADE;
DROP TYPE IF EXISTS sync_source CASCADE;
DROP TYPE IF EXISTS activity_type CASCADE;
DROP TYPE IF EXISTS owner_type CASCADE;
DROP TYPE IF EXISTS membership_type CASCADE;

-- Drop test database role
DROP ROLE IF EXISTS "user" CASCADE; 