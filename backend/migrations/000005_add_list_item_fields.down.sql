-- Remove columns from list_items table
ALTER TABLE list_items DROP COLUMN IF EXISTS available;
ALTER TABLE list_items DROP COLUMN IF EXISTS seasonal;
ALTER TABLE list_items DROP COLUMN IF EXISTS use_count;
ALTER TABLE list_items DROP COLUMN IF EXISTS start_date;
ALTER TABLE list_items DROP COLUMN IF EXISTS end_date; 