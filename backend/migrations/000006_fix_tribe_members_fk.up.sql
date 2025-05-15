-- Fix foreign key constraint name
ALTER TABLE tribe_members
    DROP CONSTRAINT IF EXISTS relationship_members_user_id_fkey,
    ADD CONSTRAINT tribe_members_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE; 