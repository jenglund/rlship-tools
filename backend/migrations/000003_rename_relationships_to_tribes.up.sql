-- Rename relationships table to tribes
ALTER TABLE relationships RENAME TO tribes;

-- Add name field to tribes
ALTER TABLE tribes
    ADD COLUMN name VARCHAR(255) NOT NULL DEFAULT 'My Tribe';

-- Rename relationship_members to tribe_members
ALTER TABLE relationship_members RENAME TO tribe_members;

-- Update foreign key references
ALTER TABLE tribe_members
    RENAME COLUMN relationship_id TO tribe_id;

-- Update foreign key constraints
ALTER TABLE tribe_members
    DROP CONSTRAINT relationship_members_relationship_id_fkey,
    ADD CONSTRAINT tribe_members_tribe_id_fkey
        FOREIGN KEY (tribe_id) REFERENCES tribes(id) ON DELETE CASCADE;

-- Rename indexes
ALTER INDEX idx_relationship_members_user_id RENAME TO idx_tribe_members_user_id;

-- Update interest_buttons foreign key
ALTER TABLE interest_buttons
    RENAME COLUMN relationship_id TO tribe_id;

ALTER TABLE interest_buttons
    DROP CONSTRAINT interest_buttons_relationship_id_fkey,
    ADD CONSTRAINT interest_buttons_tribe_id_fkey
        FOREIGN KEY (tribe_id) REFERENCES tribes(id) ON DELETE CASCADE;

-- Rename index
ALTER INDEX idx_interest_buttons_relationship_id RENAME TO idx_interest_buttons_tribe_id; 