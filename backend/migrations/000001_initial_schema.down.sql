-- Drop triggers
DROP TRIGGER IF EXISTS validate_list_owner_insert ON list_owners;
DROP TRIGGER IF EXISTS validate_list_owner_update ON list_owners;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_tribes_updated_at ON tribes;
DROP TRIGGER IF EXISTS update_tribe_members_updated_at ON tribe_members;
DROP TRIGGER IF EXISTS update_lists_updated_at ON lists;
DROP TRIGGER IF EXISTS update_list_items_updated_at ON list_items;
DROP TRIGGER IF EXISTS update_list_shares_updated_at ON list_shares;
DROP TRIGGER IF EXISTS update_activities_updated_at ON activities;
DROP TRIGGER IF EXISTS update_sync_conflicts_updated_at ON sync_conflicts;

DROP TRIGGER IF EXISTS increment_users_version ON users;
DROP TRIGGER IF EXISTS increment_tribes_version ON tribes;
DROP TRIGGER IF EXISTS increment_tribe_members_version ON tribe_members;
DROP TRIGGER IF EXISTS increment_lists_version ON lists;
DROP TRIGGER IF EXISTS increment_list_items_version ON list_items;
DROP TRIGGER IF EXISTS increment_list_shares_version ON list_shares;
DROP TRIGGER IF EXISTS increment_activities_version ON activities;
DROP TRIGGER IF EXISTS increment_sync_conflicts_version ON sync_conflicts;

-- Drop indexes
DROP INDEX IF EXISTS idx_users_firebase_uid;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_tribes_type;
DROP INDEX IF EXISTS idx_tribes_visibility;
DROP INDEX IF EXISTS idx_lists_sync_status;
DROP INDEX IF EXISTS idx_lists_sync_source;
DROP INDEX IF EXISTS idx_list_items_list_id;
DROP INDEX IF EXISTS idx_list_items_last_chosen_at;
DROP INDEX IF EXISTS idx_activities_user_id;
DROP INDEX IF EXISTS idx_activities_tribe_id;
DROP INDEX IF EXISTS idx_activities_list_id;
DROP INDEX IF EXISTS idx_activities_item_id;
DROP INDEX IF EXISTS idx_activities_scheduled_for;
DROP INDEX IF EXISTS idx_activities_completed_at;
DROP INDEX IF EXISTS idx_sync_conflicts_list_id;

-- Drop tables
DROP TABLE IF EXISTS sync_conflicts;
DROP TABLE IF EXISTS activities;
DROP TABLE IF EXISTS list_shares;
DROP TABLE IF EXISTS list_items;
DROP TABLE IF EXISTS list_owners;
DROP TABLE IF EXISTS lists;
DROP TABLE IF EXISTS tribe_members;
DROP TABLE IF EXISTS tribes;
DROP TABLE IF EXISTS users;

-- Drop enum types
DROP TYPE IF EXISTS visibility_type;
DROP TYPE IF EXISTS tribe_type;
DROP TYPE IF EXISTS sync_status;
DROP TYPE IF EXISTS sync_source;

-- Drop test database role
DROP ROLE IF EXISTS "user";

-- Drop functions
DROP FUNCTION IF EXISTS update_updated_at_column;
DROP FUNCTION IF EXISTS increment_version;
DROP FUNCTION IF EXISTS validate_list_owner;

-- Drop enum types
DROP TYPE IF EXISTS sync_status_type;
DROP TYPE IF EXISTS sharing_type;
DROP TYPE IF EXISTS membership_type; 