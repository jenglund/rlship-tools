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

-- Add trigger to ensure tribe_members deleted_at is not before its tribe's or user's deleted_at
CREATE OR REPLACE FUNCTION validate_tribe_member_deleted_at()
RETURNS TRIGGER AS $$
DECLARE
    tribe_deleted_at TIMESTAMP WITH TIME ZONE;
    user_deleted_at TIMESTAMP WITH TIME ZONE;
BEGIN
    IF NEW.deleted_at IS NOT NULL THEN
        SELECT deleted_at INTO tribe_deleted_at FROM tribes WHERE id = NEW.tribe_id;
        SELECT deleted_at INTO user_deleted_at FROM users WHERE id = NEW.user_id;
        
        IF tribe_deleted_at IS NOT NULL AND NEW.deleted_at < tribe_deleted_at THEN
            RAISE EXCEPTION 'tribe_member.deleted_at cannot be before its tribe.deleted_at';
        END IF;
        
        IF user_deleted_at IS NOT NULL AND NEW.deleted_at < user_deleted_at THEN
            RAISE EXCEPTION 'tribe_member.deleted_at cannot be before its user.deleted_at';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for validating deleted_at
DROP TRIGGER IF EXISTS trigger_validate_tribe_member_deleted_at ON tribe_members;
CREATE TRIGGER trigger_validate_tribe_member_deleted_at
    BEFORE INSERT OR UPDATE ON tribe_members
    FOR EACH ROW
    EXECUTE FUNCTION validate_tribe_member_deleted_at(); 