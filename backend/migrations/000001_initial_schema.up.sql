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
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create enum types
CREATE TYPE visibility_type AS ENUM ('private', 'public', 'shared');
CREATE TYPE tribe_type AS ENUM ('custom', 'couple', 'polycule', 'friends', 'family', 'roommates', 'coworkers');
CREATE TYPE sync_status AS ENUM ('pending', 'syncing', 'synced', 'error');
CREATE TYPE sync_source AS ENUM ('google_maps', 'manual', 'imported');

-- Create users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    firebase_uid TEXT NOT NULL UNIQUE,
    provider TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    avatar_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1
);

-- Create tribes table
CREATE TABLE tribes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    type tribe_type NOT NULL DEFAULT 'custom',
    visibility visibility_type NOT NULL DEFAULT 'private',
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1
);

-- Create tribe_members table
CREATE TABLE tribe_members (
    tribe_id UUID NOT NULL REFERENCES tribes(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1,
    PRIMARY KEY (tribe_id, user_id)
);

-- Create lists table
CREATE TABLE lists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL CHECK (type IN ('wishlist', 'places', 'activities', 'custom')),
    visibility visibility_type NOT NULL DEFAULT 'private',
    metadata JSONB NOT NULL DEFAULT '{}',
    sync_status sync_status NOT NULL DEFAULT 'pending',
    sync_source sync_source NOT NULL DEFAULT 'manual',
    sync_id TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1
);

-- Create list_owners table
CREATE TABLE list_owners (
    list_id UUID NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    owner_id UUID NOT NULL,
    owner_type TEXT NOT NULL CHECK (owner_type IN ('user', 'tribe')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (list_id, owner_id)
);

-- Create list_items table
CREATE TABLE list_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id UUID NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}',
    last_chosen_at TIMESTAMP WITH TIME ZONE,
    times_chosen INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1
);

-- Create list_shares table
CREATE TABLE list_shares (
    list_id UUID NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    tribe_id UUID NOT NULL REFERENCES tribes(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1,
    PRIMARY KEY (list_id, tribe_id)
);

-- Create activities table
CREATE TABLE activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tribe_id UUID REFERENCES tribes(id) ON DELETE CASCADE,
    list_id UUID REFERENCES lists(id) ON DELETE CASCADE,
    item_id UUID REFERENCES list_items(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    visibility visibility_type NOT NULL DEFAULT 'private',
    metadata JSONB NOT NULL DEFAULT '{}',
    scheduled_for TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1
);

-- Create sync_conflicts table
CREATE TABLE sync_conflicts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id UUID NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    conflict_data JSONB NOT NULL,
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1
);

-- Create indexes
CREATE INDEX idx_users_firebase_uid ON users(firebase_uid);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_tribes_type ON tribes(type);
CREATE INDEX idx_tribes_visibility ON tribes(visibility);
CREATE INDEX idx_lists_sync_status ON lists(sync_status);
CREATE INDEX idx_lists_sync_source ON lists(sync_source);
CREATE INDEX idx_list_items_list_id ON list_items(list_id);
CREATE INDEX idx_list_items_last_chosen_at ON list_items(last_chosen_at);
CREATE INDEX idx_activities_user_id ON activities(user_id);
CREATE INDEX idx_activities_tribe_id ON activities(tribe_id);
CREATE INDEX idx_activities_list_id ON activities(list_id);
CREATE INDEX idx_activities_item_id ON activities(item_id);
CREATE INDEX idx_activities_scheduled_for ON activities(scheduled_for);
CREATE INDEX idx_activities_completed_at ON activities(completed_at);
CREATE INDEX idx_sync_conflicts_list_id ON sync_conflicts(list_id);

-- Create test database role if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'user') THEN
        CREATE ROLE "user" WITH LOGIN PASSWORD 'user';
    END IF;
END
$$;

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

CREATE TRIGGER update_list_shares_updated_at
    BEFORE UPDATE ON list_shares
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_activities_updated_at
    BEFORE UPDATE ON activities
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sync_conflicts_updated_at
    BEFORE UPDATE ON sync_conflicts
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

CREATE TRIGGER increment_list_shares_version
    BEFORE UPDATE ON list_shares
    FOR EACH ROW
    EXECUTE FUNCTION increment_version();

CREATE TRIGGER increment_activities_version
    BEFORE UPDATE ON activities
    FOR EACH ROW
    EXECUTE FUNCTION increment_version();

CREATE TRIGGER increment_sync_conflicts_version
    BEFORE UPDATE ON sync_conflicts
    FOR EACH ROW
    EXECUTE FUNCTION increment_version();

-- Create trigger function to validate owner_id based on owner_type
CREATE OR REPLACE FUNCTION validate_list_owner()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.owner_type = 'user' THEN
        IF NOT EXISTS (SELECT 1 FROM users WHERE id = NEW.owner_id) THEN
            RAISE EXCEPTION 'Invalid user ID in list_owners';
        END IF;
    ELSIF NEW.owner_type = 'tribe' THEN
        IF NOT EXISTS (SELECT 1 FROM tribes WHERE id = NEW.owner_id) THEN
            RAISE EXCEPTION 'Invalid tribe ID in list_owners';
        END IF;
    ELSE
        RAISE EXCEPTION 'Invalid owner_type in list_owners';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for list_owners validation
CREATE TRIGGER validate_list_owner_insert
    BEFORE INSERT ON list_owners
    FOR EACH ROW
    EXECUTE FUNCTION validate_list_owner();

CREATE TRIGGER validate_list_owner_update
    BEFORE UPDATE ON list_owners
    FOR EACH ROW
    EXECUTE FUNCTION validate_list_owner(); 