-- Add last_login column to users
ALTER TABLE users ADD COLUMN last_login TIMESTAMP WITH TIME ZONE;

-- Create activities table
CREATE TABLE activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    type VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    metadata JSONB,
    visibility VARCHAR(20) NOT NULL CHECK (visibility IN ('private', 'shared', 'public')),
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
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create activity_photos table
CREATE TABLE activity_photos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes
CREATE INDEX idx_activities_user ON activities(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_activities_type ON activities(type) WHERE deleted_at IS NULL;
CREATE INDEX idx_activity_sharing_activity ON activity_sharing(activity_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_activity_sharing_tribe ON activity_sharing(tribe_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_activity_photos_activity ON activity_photos(activity_id) WHERE deleted_at IS NULL;

-- Create triggers for updated_at
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