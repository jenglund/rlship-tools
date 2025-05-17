-- Restore foreign key constraint on lists table
ALTER TABLE lists
    ADD CONSTRAINT lists_owner_id_fkey 
    FOREIGN KEY (owner_id) 
    REFERENCES users(id);

-- Remove default values from enum fields
ALTER TABLE tribes 
    ALTER COLUMN type DROP DEFAULT,
    ALTER COLUMN visibility DROP DEFAULT;

ALTER TABLE activities 
    ALTER COLUMN visibility DROP DEFAULT;

ALTER TABLE lists 
    ALTER COLUMN visibility DROP DEFAULT,
    ALTER COLUMN sync_status DROP DEFAULT,
    ALTER COLUMN owner_type DROP DEFAULT;

ALTER TABLE tribe_members 
    ALTER COLUMN membership_type DROP DEFAULT;

ALTER TABLE list_sharing 
    ALTER COLUMN sharing_type DROP DEFAULT,
    ALTER COLUMN shared_with_type DROP DEFAULT;

-- Drop list_owners table and all its dependencies
DROP TABLE IF EXISTS list_owners CASCADE; 