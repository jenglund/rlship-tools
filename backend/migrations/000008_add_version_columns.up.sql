-- Add version columns to tables that need concurrency control
ALTER TABLE tribes ADD COLUMN version INTEGER NOT NULL DEFAULT 1;
ALTER TABLE tribe_members ADD COLUMN version INTEGER NOT NULL DEFAULT 1;
ALTER TABLE users ADD COLUMN version INTEGER NOT NULL DEFAULT 1;
ALTER TABLE lists ADD COLUMN version INTEGER NOT NULL DEFAULT 1;
ALTER TABLE list_items ADD COLUMN version INTEGER NOT NULL DEFAULT 1;
ALTER TABLE list_sharing ADD COLUMN version INTEGER NOT NULL DEFAULT 1;
ALTER TABLE activities ADD COLUMN version INTEGER NOT NULL DEFAULT 1;
ALTER TABLE activity_sharing ADD COLUMN version INTEGER NOT NULL DEFAULT 1;

-- Create function to increment version on update
CREATE OR REPLACE FUNCTION increment_version()
RETURNS TRIGGER AS $$
BEGIN
    -- Only increment version if this is not a soft delete operation
    IF (TG_OP = 'UPDATE' AND NEW.deleted_at IS NULL) THEN
        NEW.version = OLD.version + 1;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers to automatically increment version on update
CREATE TRIGGER increment_tribes_version
    BEFORE UPDATE ON tribes
    FOR EACH ROW
    EXECUTE FUNCTION increment_version();

CREATE TRIGGER increment_tribe_members_version
    BEFORE UPDATE ON tribe_members
    FOR EACH ROW
    EXECUTE FUNCTION increment_version();

CREATE TRIGGER increment_users_version
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION increment_version();

CREATE TRIGGER increment_lists_version
    BEFORE UPDATE ON lists
    FOR EACH ROW
    EXECUTE FUNCTION increment_version();

CREATE TRIGGER increment_list_items_version
    BEFORE UPDATE ON list_items
    FOR EACH ROW
    EXECUTE FUNCTION increment_version();

CREATE TRIGGER increment_list_sharing_version
    BEFORE UPDATE ON list_sharing
    FOR EACH ROW
    EXECUTE FUNCTION increment_version();

CREATE TRIGGER increment_activities_version
    BEFORE UPDATE ON activities
    FOR EACH ROW
    EXECUTE FUNCTION increment_version();

CREATE TRIGGER increment_activity_sharing_version
    BEFORE UPDATE ON activity_sharing
    FOR EACH ROW
    EXECUTE FUNCTION increment_version(); 