-- Drop triggers
DROP TRIGGER IF EXISTS update_list_sharing_updated_at ON list_sharing;
DROP TRIGGER IF EXISTS update_list_items_updated_at ON list_items;
DROP TRIGGER IF EXISTS update_lists_updated_at ON lists;
DROP TRIGGER IF EXISTS update_tribe_members_updated_at ON tribe_members;
DROP TRIGGER IF EXISTS update_tribes_updated_at ON tribes;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop tables in correct order (respecting foreign key constraints)
DROP TABLE IF EXISTS list_sharing;
DROP TABLE IF EXISTS list_conflicts;
DROP TABLE IF EXISTS list_items;
DROP TABLE IF EXISTS lists;
DROP TABLE IF EXISTS tribe_members;
DROP TABLE IF EXISTS tribes;
DROP TABLE IF EXISTS users;

-- Drop functions
DROP FUNCTION IF EXISTS update_updated_at_column; 