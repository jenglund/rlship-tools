-- Drop triggers and functions
DROP TRIGGER IF EXISTS validate_activity_owner_trigger ON activity_owners;
DROP TRIGGER IF EXISTS update_activity_owners_updated_at ON activity_owners;
DROP FUNCTION IF EXISTS validate_activity_owner;

-- Drop the new activity_owners table
DROP TABLE IF EXISTS activity_owners;

-- Recreate the original activity_owners table
CREATE TABLE activity_owners (
    activity_id UUID NOT NULL REFERENCES activities(id),
    user_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (activity_id, user_id)
); 