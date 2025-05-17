-- Drop trigger
DROP TRIGGER IF EXISTS update_list_conflicts_updated_at ON list_conflicts;

-- Drop index
DROP INDEX IF EXISTS idx_list_conflicts_list_id;

-- Drop table
DROP TABLE IF EXISTS list_conflicts; 