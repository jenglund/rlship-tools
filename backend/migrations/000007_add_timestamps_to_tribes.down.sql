-- Remove updated_at and deleted_at columns from tribes table
ALTER TABLE tribes
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS deleted_at; 