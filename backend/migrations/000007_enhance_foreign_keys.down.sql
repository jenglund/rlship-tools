-- Drop constraints
ALTER TABLE tribe_members DROP CONSTRAINT IF EXISTS tribe_members_deleted_at_check;
ALTER TABLE tribe_members DROP CONSTRAINT IF EXISTS tribe_members_user_deleted_at_check;

-- Drop triggers and functions
DROP TRIGGER IF EXISTS trigger_tribe_soft_delete ON tribes;
DROP TRIGGER IF EXISTS trigger_user_soft_delete ON users;
DROP FUNCTION IF EXISTS propagate_tribe_soft_delete();
DROP FUNCTION IF EXISTS propagate_user_soft_delete();

-- Drop and recreate original foreign key constraints
ALTER TABLE tribe_members DROP CONSTRAINT IF EXISTS tribe_members_tribe_id_fkey;
ALTER TABLE tribe_members DROP CONSTRAINT IF EXISTS tribe_members_user_id_fkey;

ALTER TABLE tribe_members
    ADD CONSTRAINT tribe_members_tribe_id_fkey
    FOREIGN KEY (tribe_id)
    REFERENCES tribes(id)
    ON DELETE CASCADE;

ALTER TABLE tribe_members
    ADD CONSTRAINT tribe_members_user_id_fkey
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE CASCADE; 