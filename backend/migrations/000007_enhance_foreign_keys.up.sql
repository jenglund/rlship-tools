-- Drop existing foreign key constraints
ALTER TABLE tribe_members DROP CONSTRAINT IF EXISTS tribe_members_tribe_id_fkey;
ALTER TABLE tribe_members DROP CONSTRAINT IF EXISTS tribe_members_user_id_fkey;

-- Recreate foreign key constraints with proper cascade rules
ALTER TABLE tribe_members
    ADD CONSTRAINT tribe_members_tribe_id_fkey
    FOREIGN KEY (tribe_id)
    REFERENCES tribes(id)
    ON DELETE CASCADE
    ON UPDATE CASCADE;

ALTER TABLE tribe_members
    ADD CONSTRAINT tribe_members_user_id_fkey
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE CASCADE
    ON UPDATE CASCADE;

-- Add trigger for soft delete propagation
CREATE OR REPLACE FUNCTION propagate_tribe_soft_delete()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL THEN
        -- Soft delete all tribe members when tribe is soft deleted
        UPDATE tribe_members
        SET deleted_at = NEW.deleted_at
        WHERE tribe_id = NEW.id AND deleted_at IS NULL;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for soft delete propagation
DROP TRIGGER IF EXISTS trigger_tribe_soft_delete ON tribes;
CREATE TRIGGER trigger_tribe_soft_delete
    AFTER UPDATE ON tribes
    FOR EACH ROW
    EXECUTE FUNCTION propagate_tribe_soft_delete();

-- Add trigger for user soft delete propagation
CREATE OR REPLACE FUNCTION propagate_user_soft_delete()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL THEN
        -- Soft delete all tribe memberships when user is soft deleted
        UPDATE tribe_members
        SET deleted_at = NEW.deleted_at
        WHERE user_id = NEW.id AND deleted_at IS NULL;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for user soft delete propagation
DROP TRIGGER IF EXISTS trigger_user_soft_delete ON users;
CREATE TRIGGER trigger_user_soft_delete
    AFTER UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION propagate_user_soft_delete();

-- Add constraint to ensure tribe_members deleted_at is not before its tribe's deleted_at
ALTER TABLE tribe_members
    ADD CONSTRAINT tribe_members_deleted_at_check
    CHECK (
        deleted_at IS NULL OR
        deleted_at >= (SELECT deleted_at FROM tribes WHERE id = tribe_id)
    );

-- Add constraint to ensure tribe_members deleted_at is not before its user's deleted_at
ALTER TABLE tribe_members
    ADD CONSTRAINT tribe_members_user_deleted_at_check
    CHECK (
        deleted_at IS NULL OR
        deleted_at >= (SELECT deleted_at FROM users WHERE id = user_id)
    ); 