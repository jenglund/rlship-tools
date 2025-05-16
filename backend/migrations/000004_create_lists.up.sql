-- Create lists table
CREATE TABLE lists (
    id UUID PRIMARY KEY,
    type TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    visibility TEXT NOT NULL,
    
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

-- Create indexes
CREATE INDEX idx_lists_type ON lists(type) WHERE deleted_at IS NULL;
CREATE INDEX idx_lists_sync_source ON lists(sync_source) WHERE deleted_at IS NULL;
CREATE INDEX idx_list_items_list_id ON list_items(list_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_list_items_external_id ON list_items(external_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_list_conflicts_list_id ON list_conflicts(list_id) WHERE resolved_at IS NULL;

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_lists_updated_at
    BEFORE UPDATE ON lists
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_list_items_updated_at
    BEFORE UPDATE ON list_items
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column(); 