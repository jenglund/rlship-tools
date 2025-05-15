-- Add activity type enum
CREATE TYPE activity_type AS ENUM ('location', 'interest', 'list', 'activity');

-- Add visibility type enum
CREATE TYPE visibility_type AS ENUM ('private', 'public');

-- Create activities table
CREATE TABLE activities (
    id UUID PRIMARY KEY,
    type activity_type NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    visibility visibility_type NOT NULL DEFAULT 'private',
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create activity owners table
CREATE TABLE activity_owners (
    activity_id UUID NOT NULL REFERENCES activities(id),
    owner_id UUID NOT NULL,
    owner_type VARCHAR(10) NOT NULL CHECK (owner_type IN ('user', 'tribe')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (activity_id, owner_id)
);

-- Create activity shares table
CREATE TABLE activity_shares (
    activity_id UUID NOT NULL REFERENCES activities(id),
    tribe_id UUID NOT NULL REFERENCES tribes(id),
    user_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (activity_id, tribe_id, user_id)
);

-- Add indexes
CREATE INDEX idx_activities_type ON activities(type) WHERE deleted_at IS NULL;
CREATE INDEX idx_activities_visibility ON activities(visibility) WHERE deleted_at IS NULL;
CREATE INDEX idx_activity_owners_owner ON activity_owners(owner_id, owner_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_activity_shares_tribe ON activity_shares(tribe_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_activity_shares_user ON activity_shares(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_activity_shares_expires ON activity_shares(expires_at) WHERE deleted_at IS NULL; 