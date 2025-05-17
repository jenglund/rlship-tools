-- Fix default values for enum fields first
ALTER TABLE tribes 
    ALTER COLUMN type SET DEFAULT 'custom',
    ALTER COLUMN visibility SET DEFAULT 'private';

ALTER TABLE activities 
    ALTER COLUMN visibility SET DEFAULT 'private';

ALTER TABLE lists 
    ALTER COLUMN visibility SET DEFAULT 'private',
    ALTER COLUMN sync_status SET DEFAULT 'none',
    ALTER COLUMN owner_type SET DEFAULT 'user';

ALTER TABLE tribe_members 
    ALTER COLUMN membership_type SET DEFAULT 'full';

ALTER TABLE list_sharing 
    ALTER COLUMN sharing_type SET DEFAULT 'view',
    ALTER COLUMN shared_with_type SET DEFAULT 'user';

-- Add NOT NULL constraints with defaults for metadata
UPDATE tribes SET metadata = '{}' WHERE metadata IS NULL;
UPDATE activities SET metadata = '{}' WHERE metadata IS NULL;
UPDATE list_items SET metadata = '{}' WHERE metadata IS NULL;
UPDATE activity_photos SET metadata = '{}' WHERE metadata IS NULL;

-- Create list_owners table
CREATE TABLE list_owners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id UUID NOT NULL,  -- We'll add the foreign key constraint later
    owner_id UUID NOT NULL,
    owner_type owner_type NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE (list_id, owner_id, owner_type)
);

-- Add indexes for list_owners
CREATE INDEX idx_list_owners_list ON list_owners(list_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_list_owners_owner ON list_owners(owner_id, owner_type) WHERE deleted_at IS NULL;

-- Add triggers for list_owners
CREATE TRIGGER update_list_owners_updated_at
    BEFORE UPDATE ON list_owners
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_list_owners_version
    BEFORE UPDATE ON list_owners
    FOR EACH ROW
    EXECUTE FUNCTION increment_version();

-- Migrate existing data to list_owners table
INSERT INTO list_owners (list_id, owner_id, owner_type, created_at, updated_at)
SELECT id, owner_id, owner_type, created_at, updated_at
FROM lists
ON CONFLICT DO NOTHING;

-- Now add the foreign key constraints
ALTER TABLE list_owners
    ADD CONSTRAINT list_owners_list_id_fkey
    FOREIGN KEY (list_id)
    REFERENCES lists(id)
    ON DELETE CASCADE;

-- Update lists table to reference list_owners
ALTER TABLE lists
    DROP CONSTRAINT lists_owner_id_fkey;

-- Add corresponding down migration
COMMENT ON TABLE list_owners IS 'Created in migration 000002_fix_constraints.up.sql'; 