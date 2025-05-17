-- Create list_conflicts table if it doesn't exist
CREATE TABLE IF NOT EXISTS list_conflicts (
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

-- Create index
CREATE INDEX IF NOT EXISTS idx_list_conflicts_list_id ON list_conflicts(list_id);

-- Create triggers
CREATE TRIGGER update_list_conflicts_updated_at
    BEFORE UPDATE ON list_conflicts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column(); 