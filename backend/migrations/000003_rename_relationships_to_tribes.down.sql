-- Revert interest_buttons changes
ALTER TABLE interest_buttons
    RENAME COLUMN tribe_id TO relationship_id;

ALTER TABLE interest_buttons
    DROP CONSTRAINT interest_buttons_tribe_id_fkey,
    ADD CONSTRAINT interest_buttons_relationship_id_fkey
        FOREIGN KEY (relationship_id) REFERENCES relationships(id) ON DELETE CASCADE;

ALTER INDEX idx_interest_buttons_tribe_id RENAME TO idx_interest_buttons_relationship_id;

-- Revert tribe_members changes
ALTER TABLE tribe_members
    RENAME COLUMN tribe_id TO relationship_id;

ALTER TABLE tribe_members
    DROP CONSTRAINT tribe_members_tribe_id_fkey,
    ADD CONSTRAINT relationship_members_relationship_id_fkey
        FOREIGN KEY (relationship_id) REFERENCES relationships(id) ON DELETE CASCADE;

ALTER INDEX idx_tribe_members_user_id RENAME TO idx_relationship_members_user_id;

-- Rename tribe_members back to relationship_members
ALTER TABLE tribe_members RENAME TO relationship_members;

-- Remove name from tribes
ALTER TABLE tribes DROP COLUMN name;

-- Rename tribes back to relationships
ALTER TABLE tribes RENAME TO relationships; 