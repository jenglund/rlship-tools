-- Add membership type enum
CREATE TYPE membership_type AS ENUM ('full', 'guest');

-- Add new columns to tribe_members table
ALTER TABLE tribe_members
    ADD COLUMN type membership_type NOT NULL DEFAULT 'full',
    ADD COLUMN expires_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN invited_by UUID REFERENCES users(id);

-- Add indexes
CREATE INDEX idx_tribe_members_type ON tribe_members(type);
CREATE INDEX idx_tribe_members_expires ON tribe_members(expires_at);
CREATE INDEX idx_tribe_members_invited_by ON tribe_members(invited_by); 