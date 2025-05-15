-- Add updated_at column to tribes table
ALTER TABLE tribes
    ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

-- Add deleted_at column to tribes table
ALTER TABLE tribes
    ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE; 