-- Check if invited_by column already exists, add it if not
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'tribe_members' AND column_name = 'invited_by'
    ) THEN
        ALTER TABLE tribe_members ADD COLUMN invited_by UUID REFERENCES users(id);
    END IF;
END $$;

-- Check if invited_at column already exists, add it if not
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'tribe_members' AND column_name = 'invited_at'
    ) THEN
        ALTER TABLE tribe_members ADD COLUMN invited_at TIMESTAMP WITH TIME ZONE;
    END IF;
END $$; 