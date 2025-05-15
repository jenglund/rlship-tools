-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    avatar_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Relationships table
CREATE TABLE relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Relationship members table (for handling multiple people in a relationship)
CREATE TABLE relationship_members (
    relationship_id UUID REFERENCES relationships(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (relationship_id, user_id)
);

-- Activity lists table
CREATE TABLE activity_lists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    relationship_id UUID REFERENCES relationships(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Activities table
CREATE TABLE activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id UUID REFERENCES activity_lists(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    location TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Activity notes table
CREATE TABLE activity_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID REFERENCES activities(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    is_private BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Activity events table (for tracking when activities are done)
CREATE TABLE activity_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID REFERENCES activities(id) ON DELETE CASCADE,
    relationship_id UUID REFERENCES relationships(id) ON DELETE CASCADE,
    date TIMESTAMP WITH TIME ZONE NOT NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Activity event photos table
CREATE TABLE activity_event_photos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID REFERENCES activity_events(id) ON DELETE CASCADE,
    photo_url TEXT NOT NULL,
    uploaded_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Interest buttons table
CREATE TABLE interest_buttons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    relationship_id UUID REFERENCES relationships(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    default_duration INTEGER NOT NULL, -- in minutes
    allow_early_cancel BOOLEAN DEFAULT true,
    is_loud BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Interest button activations table
CREATE TABLE interest_button_activations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    button_id UUID REFERENCES interest_buttons(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    duration INTEGER NOT NULL, -- in minutes
    activated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    cancelled_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes
CREATE INDEX idx_relationship_members_user_id ON relationship_members(user_id);
CREATE INDEX idx_activity_lists_relationship_id ON activity_lists(relationship_id);
CREATE INDEX idx_activities_list_id ON activities(list_id);
CREATE INDEX idx_activity_notes_activity_id ON activity_notes(activity_id);
CREATE INDEX idx_activity_notes_user_id ON activity_notes(user_id);
CREATE INDEX idx_activity_events_activity_id ON activity_events(activity_id);
CREATE INDEX idx_activity_events_relationship_id ON activity_events(relationship_id);
CREATE INDEX idx_activity_event_photos_event_id ON activity_event_photos(event_id);
CREATE INDEX idx_interest_buttons_relationship_id ON interest_buttons(relationship_id);
CREATE INDEX idx_interest_button_activations_button_id ON interest_button_activations(button_id);
CREATE INDEX idx_interest_button_activations_user_id ON interest_button_activations(user_id); 