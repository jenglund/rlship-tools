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

-- Drop functions
DROP FUNCTION IF EXISTS validate_activity_owner();
DROP FUNCTION IF EXISTS validate_list_owner();
DROP FUNCTION IF EXISTS increment_version();
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_unique_tribe_name;
DROP INDEX IF EXISTS idx_users_firebase_uid;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_tribes_name;
DROP INDEX IF EXISTS idx_tribe_members_tribe_id;
DROP INDEX IF EXISTS idx_tribe_members_user_id;
DROP INDEX IF EXISTS idx_lists_owner_user;
DROP INDEX IF EXISTS idx_lists_owner_tribe;
DROP INDEX IF EXISTS idx_lists_sync_id;
DROP INDEX IF EXISTS idx_list_items_list_id;
DROP INDEX IF EXISTS idx_activities_user_id;
DROP INDEX IF EXISTS idx_activity_photos_activity_id;
DROP INDEX IF EXISTS idx_activity_shares_activity_id;
DROP INDEX IF EXISTS idx_list_sharing_list_id;
DROP INDEX IF EXISTS idx_sync_conflicts_list_id;
DROP INDEX IF EXISTS idx_list_conflicts_list_id;

-- Drop tables
DROP TABLE IF EXISTS activity_owners;
DROP TABLE IF EXISTS activity_photos;
DROP TABLE IF EXISTS activity_shares;
DROP TABLE IF EXISTS list_sharing;
DROP TABLE IF EXISTS list_items;
DROP TABLE IF EXISTS list_owners;
DROP TABLE IF EXISTS sync_conflicts;
DROP TABLE IF EXISTS list_conflicts;
DROP TABLE IF EXISTS lists;
DROP TABLE IF EXISTS activities;
DROP TABLE IF EXISTS tribe_members;
DROP TABLE IF EXISTS tribes;
DROP TABLE IF EXISTS users;

-- Drop test database role
DROP ROLE IF EXISTS "user";

-- Drop enums
DROP TYPE IF EXISTS visibility_type;
DROP TYPE IF EXISTS tribe_type;
DROP TYPE IF EXISTS sync_status;
DROP TYPE IF EXISTS sync_source;
DROP TYPE IF EXISTS activity_type;
DROP TYPE IF EXISTS owner_type;
DROP TYPE IF EXISTS membership_type; 