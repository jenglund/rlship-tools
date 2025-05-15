-- Drop indexes
DROP INDEX IF EXISTS idx_tribe_members_type;
DROP INDEX IF EXISTS idx_tribe_members_expires;
DROP INDEX IF EXISTS idx_tribe_members_invited_by;

-- Drop columns from tribe_members table
ALTER TABLE tribe_members
    DROP COLUMN IF EXISTS type,
    DROP COLUMN IF EXISTS expires_at,
    DROP COLUMN IF EXISTS invited_by;

-- Drop enum
DROP TYPE IF EXISTS membership_type; 