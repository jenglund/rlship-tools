-- Create list owners table
CREATE TABLE list_owners (
    list_id UUID NOT NULL REFERENCES lists(id),
    owner_id UUID NOT NULL,
    owner_type VARCHAR(10) NOT NULL CHECK (owner_type IN ('user', 'tribe')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (list_id, owner_id)
);

-- Create list shares table
CREATE TABLE list_shares (
    list_id UUID NOT NULL REFERENCES lists(id),
    tribe_id UUID NOT NULL REFERENCES tribes(id),
    user_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (list_id, tribe_id, user_id)
);

-- Add indexes
CREATE INDEX idx_list_owners_owner ON list_owners(owner_id, owner_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_list_shares_tribe ON list_shares(tribe_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_list_shares_user ON list_shares(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_list_shares_expires ON list_shares(expires_at) WHERE deleted_at IS NULL; 