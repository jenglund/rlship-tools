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
CREATE INDEX idx_interest_buttons_relationship_id ON interest_buttons(relationship_id);
CREATE INDEX idx_interest_button_activations_button_id ON interest_button_activations(button_id);
CREATE INDEX idx_interest_button_activations_user_id ON interest_button_activations(user_id); 