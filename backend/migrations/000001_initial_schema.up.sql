-- Create utility functions
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

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

-- Create enum types
DO $$ BEGIN
    CREATE TYPE owner_type AS ENUM ('user', 'tribe');
    CREATE TYPE visibility_type AS ENUM ('private', 'shared', 'public');
    CREATE TYPE tribe_type AS ENUM ('couple', 'polycule', 'friends', 'family', 'roommates', 'coworkers', 'custom');
    CREATE TYPE membership_type AS ENUM ('full', 'limited', 'guest');
    CREATE TYPE sharing_type AS ENUM ('view', 'edit', 'admin');
    CREATE TYPE sync_status_type AS ENUM ('none', 'pending', 'synced', 'error');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    firebase_uid VARCHAR(255) UNIQUE NOT NULL,
    provider VARCHAR(50) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    avatar_url TEXT,
    last_login TIMESTAMP WITH TIME ZONE,
    version INTEGER NOT NULL DEFAULT 1,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create tribes table
CREATE TABLE tribes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type tribe_type NOT NULL,
    description TEXT,
    visibility visibility_type NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create tribe_members table
CREATE TABLE tribe_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tribe_id UUID NOT NULL REFERENCES tribes(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    membership_type membership_type NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    version INTEGER NOT NULL DEFAULT 1,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create lists table
CREATE TABLE lists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    visibility visibility_type NOT NULL,
    owner_id UUID NOT NULL REFERENCES users(id),
    owner_type owner_type NOT NULL,
    
    -- Sync information
    sync_status sync_status_type NOT NULL DEFAULT 'none',
    sync_source TEXT,
    sync_id TEXT,
    last_sync_at TIMESTAMP WITH TIME ZONE,
    
    -- Menu generation settings
    default_weight DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    max_items INTEGER,
    cooldown_days INTEGER,
    
    version INTEGER NOT NULL DEFAULT 1,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create list items table
CREATE TABLE list_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id UUID NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    
    -- Item metadata
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    external_id TEXT,
    
    -- Menu generation parameters
    weight DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    last_chosen TIMESTAMP WITH TIME ZONE,
    chosen_count INTEGER NOT NULL DEFAULT 0,
    
    -- Location-specific fields
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    address TEXT,
    
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create list conflicts table
CREATE TABLE list_conflicts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id UUID NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    item_id UUID REFERENCES list_items(id) ON DELETE CASCADE,
    
    conflict_type TEXT NOT NULL,
    local_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    external_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE
);

-- Create list sharing table
CREATE TABLE list_sharing (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id UUID NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    shared_with_id UUID NOT NULL,
    shared_with_type owner_type NOT NULL,
    sharing_type sharing_type NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create activities table
CREATE TABLE activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    visibility visibility_type NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create activity_sharing table
CREATE TABLE activity_sharing (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    tribe_id UUID NOT NULL REFERENCES tribes(id) ON DELETE CASCADE,
    expires_at TIMESTAMP WITH TIME ZONE,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create activity_photos table
CREATE TABLE activity_photos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    caption TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
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

CREATE INDEX idx_activities_user ON activities(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_activities_type ON activities(type) WHERE deleted_at IS NULL;
CREATE INDEX idx_activity_sharing_activity ON activity_sharing(activity_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_activity_sharing_tribe ON activity_sharing(tribe_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_activity_photos_activity ON activity_photos(activity_id) WHERE deleted_at IS NULL;

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

CREATE TRIGGER update_activities_updated_at
    BEFORE UPDATE ON activities
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_activity_sharing_updated_at
    BEFORE UPDATE ON activity_sharing
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_activity_photos_updated_at
    BEFORE UPDATE ON activity_photos
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create version increment triggers
CREATE TRIGGER increment_users_version
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION increment_version();

CREATE TRIGGER increment_tribes_version
    BEFORE UPDATE ON tribes
    FOR EACH ROW
    EXECUTE FUNCTION increment_version();

CREATE TRIGGER increment_tribe_members_version
    BEFORE UPDATE ON tribe_members
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