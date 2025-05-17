-- Drop the old activity_owners table
DROP TABLE IF EXISTS activity_owners;

-- Create the new activity_owners table with correct schema
CREATE TABLE activity_owners (
    activity_id UUID NOT NULL REFERENCES activities(id),
    owner_id UUID NOT NULL,
    owner_type owner_type NOT NULL DEFAULT 'user',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (activity_id, owner_id, owner_type),
    UNIQUE (activity_id, owner_id)
);

-- Create owner validation function and trigger
CREATE OR REPLACE FUNCTION validate_activity_owner()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.owner_type = 'user' AND NOT EXISTS (SELECT 1 FROM users WHERE id = NEW.owner_id) THEN
        RAISE EXCEPTION 'User with ID % does not exist', NEW.owner_id;
    ELSIF NEW.owner_type = 'tribe' AND NOT EXISTS (SELECT 1 FROM tribes WHERE id = NEW.owner_id) THEN
        RAISE EXCEPTION 'Tribe with ID % does not exist', NEW.owner_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER validate_activity_owner_trigger
    BEFORE INSERT OR UPDATE ON activity_owners
    FOR EACH ROW
    EXECUTE FUNCTION validate_activity_owner();

-- Create trigger for updated_at
CREATE TRIGGER update_activity_owners_updated_at
    BEFORE UPDATE ON activity_owners
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column(); 