-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    firebase_uid VARCHAR(255) UNIQUE NOT NULL,
    provider VARCHAR(50) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    avatar_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create tribes table
CREATE TABLE tribes (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('couple', 'polycule', 'friends', 'family', 'roommates', 'coworkers', 'custom')),
    description TEXT,
    visibility VARCHAR(20) NOT NULL CHECK (visibility IN ('private', 'shared', 'public')),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create tribe_members table
CREATE TABLE tribe_members (
    id UUID PRIMARY KEY,
    tribe_id UUID NOT NULL REFERENCES tribes(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    membership_type VARCHAR(20) NOT NULL CHECK (membership_type IN ('full', 'limited', 'guest')),
    display_name VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create lists table
CREATE TABLE lists (
    id UUID PRIMARY KEY,
    type TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    visibility TEXT NOT NULL,
    owner_id UUID NOT NULL REFERENCES users(id),
    owner_type TEXT NOT NULL CHECK (owner_type IN ('user', 'tribe')),
    
    -- Sync information
    sync_status TEXT NOT NULL DEFAULT 'none',
    sync_source TEXT,
    sync_id TEXT,
    last_sync_at TIMESTAMP WITH TIME ZONE,
    
    -- Menu generation settings
    default_weight DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    max_items INTEGER,
    cooldown_days INTEGER,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create list items table
CREATE TABLE list_items (
    id UUID PRIMARY KEY,
    list_id UUID NOT NULL REFERENCES lists(id),
    name TEXT NOT NULL,
    description TEXT,
    
    -- Item metadata
    metadata JSONB,
    external_id TEXT,
    
    -- Menu generation parameters
    weight DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    last_chosen TIMESTAMP WITH TIME ZONE,
    chosen_count INTEGER NOT NULL DEFAULT 0,
    
    -- Location-specific fields
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    address TEXT,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create list conflicts table
CREATE TABLE list_conflicts (
    id UUID PRIMARY KEY,
    list_id UUID NOT NULL REFERENCES lists(id),
    item_id UUID REFERENCES list_items(id),
    
    conflict_type TEXT NOT NULL,
    local_data JSONB,
    external_data JSONB,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE
);

-- Create list sharing table
CREATE TABLE list_sharing (
    id UUID PRIMARY KEY,
    list_id UUID NOT NULL REFERENCES lists(id),
    shared_with_id UUID NOT NULL,
    shared_with_type TEXT NOT NULL CHECK (shared_with_type IN ('user', 'tribe')),
    sharing_type TEXT NOT NULL CHECK (sharing_type IN ('view', 'edit', 'admin')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_firebase_uid ON users(firebase_uid) WHERE deleted_at IS NULL;

CREATE INDEX idx_tribes_type ON tribes(type) WHERE deleted_at IS NULL;
CREATE INDEX idx_tribes_visibility ON tribes(visibility) WHERE deleted_at IS NULL;

CREATE INDEX idx_tribe_members_tribe ON tribe_members(tribe_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tribe_members_user ON tribe_members(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tribe_members_expires ON tribe_members(expires_at) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_tribe_members_unique ON tribe_members (tribe_id, user_id) WHERE deleted_at IS NULL;

CREATE INDEX idx_lists_type ON lists(type) WHERE deleted_at IS NULL;
CREATE INDEX idx_lists_sync_source ON lists(sync_source) WHERE deleted_at IS NULL;
CREATE INDEX idx_lists_owner ON lists(owner_id, owner_type) WHERE deleted_at IS NULL;

CREATE INDEX idx_list_items_list_id ON list_items(list_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_list_items_external_id ON list_items(external_id) WHERE deleted_at IS NULL;

CREATE INDEX idx_list_conflicts_list_id ON list_conflicts(list_id) WHERE resolved_at IS NULL;

CREATE INDEX idx_list_sharing_list ON list_sharing(list_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_list_sharing_shared_with ON list_sharing(shared_with_id, shared_with_type) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_list_sharing_unique ON list_sharing (list_id, shared_with_id, shared_with_type) WHERE deleted_at IS NULL;

-- Create triggers for updated_at
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tribes_updated_at
    BEFORE UPDATE ON tribes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tribe_members_updated_at
    BEFORE UPDATE ON tribe_members
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_lists_updated_at
    BEFORE UPDATE ON lists
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_list_items_updated_at
    BEFORE UPDATE ON list_items
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_list_sharing_updated_at
    BEFORE UPDATE ON list_sharing
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column(); 