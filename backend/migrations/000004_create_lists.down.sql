-- Drop triggers
DROP TRIGGER IF EXISTS update_list_items_updated_at ON list_items;
DROP TRIGGER IF EXISTS update_lists_updated_at ON lists;

-- Drop tables (in correct order due to foreign key constraints)
DROP TABLE IF EXISTS list_conflicts;
DROP TABLE IF EXISTS list_items;
DROP TABLE IF EXISTS lists; 