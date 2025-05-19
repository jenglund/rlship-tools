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
CREATE TYPE sync_status AS ENUM ('none', 'pending', 'synced', 'conflict');
CREATE TYPE sync_source AS ENUM ('none', 'google_maps', 'manual', 'imported', '', 'unknown_source');
CREATE TYPE activity_type AS ENUM ('location', 'interest', 'list', 'custom', 'activity', 'event');
CREATE TYPE owner_type AS ENUM ('user', 'tribe');
CREATE TYPE membership_type AS ENUM ('full', 'limited', 'guest', 'pending');

-- Create users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    firebase_uid TEXT NOT NULL UNIQUE,
    provider TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    avatar_url TEXT,
    last_login TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create tribes table
CREATE TABLE tribes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    type tribe_type NOT NULL DEFAULT 'custom',
    visibility visibility_type NOT NULL DEFAULT 'private',
    metadata JSONB NOT NULL DEFAULT '{}' CHECK (metadata IS NOT NULL AND metadata != 'null'::jsonb),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create partial unique index for tribe names (only where not deleted)
CREATE UNIQUE INDEX idx_unique_tribe_name ON tribes (name) WHERE deleted_at IS NULL;

-- Create tribe_members table
CREATE TABLE tribe_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tribe_id UUID NOT NULL REFERENCES tribes(id),
    user_id UUID NOT NULL REFERENCES users(id),
    membership_type membership_type NOT NULL DEFAULT 'full',
    display_name TEXT NOT NULL DEFAULT 'Member',
    expires_at TIMESTAMP WITH TIME ZONE,
    invited_by UUID REFERENCES users(id),
    invited_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB NOT NULL DEFAULT '{}' CHECK (metadata IS NOT NULL AND metadata != 'null'::jsonb),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1,
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE (tribe_id, user_id)
);

-- Create lists table
CREATE TABLE lists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL,
    owner_type owner_type NOT NULL DEFAULT 'user',
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    visibility visibility_type NOT NULL DEFAULT 'private',
    type TEXT NOT NULL,
    default_weight FLOAT NOT NULL DEFAULT 1.0,
    max_items INTEGER,
    cooldown_days INTEGER,
    sync_source sync_source NOT NULL DEFAULT 'none',
    sync_id TEXT,
    sync_status sync_status NOT NULL DEFAULT 'none',
    last_sync_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB NOT NULL DEFAULT '{}' CHECK (metadata IS NOT NULL AND metadata != 'null'::jsonb),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create list_owners table
CREATE TABLE list_owners (
    list_id UUID NOT NULL REFERENCES lists(id),
    owner_id UUID NOT NULL,
    owner_type owner_type NOT NULL DEFAULT 'user',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (list_id, owner_id, owner_type)
);

-- Create list_items table
CREATE TABLE list_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id UUID NOT NULL REFERENCES lists(id),
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    weight FLOAT NOT NULL DEFAULT 1.0,
    external_id TEXT,
    latitude FLOAT,
    longitude FLOAT,
    address TEXT,
    seasonal BOOLEAN NOT NULL DEFAULT false,
    season_begin TEXT,
    season_end TEXT,
    cooldown INTEGER,
    last_chosen TIMESTAMP WITH TIME ZONE,
    chosen_count INTEGER NOT NULL DEFAULT 0,
    metadata JSONB NOT NULL DEFAULT '{}' CHECK (metadata IS NOT NULL AND metadata != 'null'::jsonb),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create activities table
CREATE TABLE activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    type activity_type NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    visibility visibility_type NOT NULL DEFAULT 'private',
    metadata JSONB NOT NULL DEFAULT '{}' CHECK (metadata IS NOT NULL AND metadata != 'null'::jsonb),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create activity_owners table
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

-- Create activity_photos table
CREATE TABLE activity_photos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL REFERENCES activities(id),
    url TEXT NOT NULL,
    caption TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}' CHECK (metadata IS NOT NULL AND metadata != 'null'::jsonb),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    version INTEGER NOT NULL DEFAULT 1
);

-- Create activity_shares table
CREATE TABLE activity_shares (
    activity_id UUID NOT NULL REFERENCES activities(id),
    tribe_id UUID NOT NULL REFERENCES tribes(id),
    user_id UUID NOT NULL REFERENCES users(id),
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    version INTEGER NOT NULL DEFAULT 1,
    PRIMARY KEY (activity_id, tribe_id)
);

-- Create list_sharing table
CREATE TABLE list_sharing (
    list_id UUID NOT NULL REFERENCES lists(id),
    tribe_id UUID NOT NULL REFERENCES tribes(id),
    user_id UUID NOT NULL REFERENCES users(id),
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    version INTEGER NOT NULL DEFAULT 1,
    PRIMARY KEY (list_id, tribe_id)
);

-- Create sync_conflicts table
CREATE TABLE sync_conflicts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id UUID NOT NULL REFERENCES lists(id),
    type TEXT NOT NULL,
    local_data TEXT NOT NULL,
    remote_data TEXT NOT NULL,
    resolved BOOLEAN NOT NULL DEFAULT false,
    resolution TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create list_conflicts table
CREATE TABLE list_conflicts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id UUID NOT NULL REFERENCES lists(id),
    type TEXT NOT NULL,
    local_data JSONB NOT NULL,
    remote_data JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes
CREATE INDEX idx_users_firebase_uid ON users(firebase_uid);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_tribes_name ON tribes(name);
CREATE INDEX idx_tribe_members_tribe_id ON tribe_members(tribe_id);
CREATE INDEX idx_tribe_members_user_id ON tribe_members(user_id);
CREATE INDEX idx_lists_owner_user ON lists(owner_id) WHERE owner_type = 'user';
CREATE INDEX idx_lists_owner_tribe ON lists(owner_id) WHERE owner_type = 'tribe';
CREATE INDEX idx_lists_sync_id ON lists(sync_id);
CREATE INDEX idx_list_items_list_id ON list_items(list_id);
CREATE INDEX idx_activities_user_id ON activities(user_id);
CREATE INDEX idx_activity_photos_activity_id ON activity_photos(activity_id);
CREATE INDEX idx_activity_shares_activity_id ON activity_shares(activity_id);
CREATE INDEX idx_list_sharing_list_id ON list_sharing(list_id);
CREATE INDEX idx_sync_conflicts_list_id ON sync_conflicts(list_id);
CREATE INDEX idx_list_conflicts_list_id ON list_conflicts(list_id);

-- Create test database role if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'user') THEN
        CREATE ROLE "user" WITH LOGIN PASSWORD 'user';
    END IF;
END
$$;

-- Create triggers for updated_at and version
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_users_version
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION increment_version();

CREATE TRIGGER update_tribes_updated_at
    BEFORE UPDATE ON tribes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_tribes_version
    BEFORE UPDATE ON tribes
    FOR EACH ROW
    EXECUTE FUNCTION increment_version();

CREATE TRIGGER update_tribe_members_updated_at
    BEFORE UPDATE ON tribe_members
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_tribe_members_version
    BEFORE UPDATE ON tribe_members
    FOR EACH ROW
    EXECUTE FUNCTION increment_version();

CREATE TRIGGER update_activities_updated_at
    BEFORE UPDATE ON activities
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_activities_version
    BEFORE UPDATE ON activities
    FOR EACH ROW
    EXECUTE FUNCTION increment_version();

CREATE TRIGGER update_activity_photos_updated_at
    BEFORE UPDATE ON activity_photos
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_activity_photos_version
    BEFORE UPDATE ON activity_photos
    FOR EACH ROW
    EXECUTE FUNCTION increment_version();

CREATE TRIGGER update_activity_shares_updated_at
    BEFORE UPDATE ON activity_shares
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_lists_updated_at
    BEFORE UPDATE ON lists
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_lists_version
    BEFORE UPDATE ON lists
    FOR EACH ROW
    EXECUTE FUNCTION increment_version();

CREATE TRIGGER update_list_items_updated_at
    BEFORE UPDATE ON list_items
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_list_items_version
    BEFORE UPDATE ON list_items
    FOR EACH ROW
    EXECUTE FUNCTION increment_version();

CREATE TRIGGER update_list_sharing_updated_at
    BEFORE UPDATE ON list_sharing
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_list_sharing_version
    BEFORE UPDATE ON list_sharing
    FOR EACH ROW
    EXECUTE FUNCTION increment_version();

CREATE TRIGGER update_list_conflicts_updated_at
    BEFORE UPDATE ON list_conflicts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE OR REPLACE FUNCTION validate_list_owner()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.owner_type = 'user' THEN
        IF NOT EXISTS (SELECT 1 FROM users WHERE id = NEW.owner_id) THEN
            RAISE EXCEPTION 'User with ID % does not exist', NEW.owner_id;
        END IF;
    ELSIF NEW.owner_type = 'tribe' THEN
        IF NOT EXISTS (SELECT 1 FROM tribes WHERE id = NEW.owner_id) THEN
            RAISE EXCEPTION 'Tribe with ID % does not exist', NEW.owner_id;
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER validate_list_owner_trigger
    BEFORE INSERT OR UPDATE ON list_owners
    FOR EACH ROW
    EXECUTE FUNCTION validate_list_owner();

CREATE OR REPLACE FUNCTION validate_activity_owner()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.owner_type = 'user' THEN
        IF NOT EXISTS (SELECT 1 FROM users WHERE id = NEW.owner_id) THEN
            RAISE EXCEPTION 'User with ID % does not exist', NEW.owner_id;
        END IF;
    ELSIF NEW.owner_type = 'tribe' THEN
        IF NOT EXISTS (SELECT 1 FROM tribes WHERE id = NEW.owner_id) THEN
            RAISE EXCEPTION 'Tribe with ID % does not exist', NEW.owner_id;
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER validate_activity_owner_trigger
    BEFORE INSERT OR UPDATE ON activity_owners
    FOR EACH ROW
    EXECUTE FUNCTION validate_activity_owner();

CREATE TRIGGER update_activity_owners_updated_at
    BEFORE UPDATE ON activity_owners
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column(); 