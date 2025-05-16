-- Convert metadata columns to JSONB type
ALTER TABLE activities ALTER COLUMN metadata TYPE JSONB USING metadata::jsonb;
ALTER TABLE lists ALTER COLUMN metadata TYPE JSONB USING metadata::jsonb;
ALTER TABLE list_items ALTER COLUMN metadata TYPE JSONB USING metadata::jsonb;
ALTER TABLE tribes ALTER COLUMN metadata TYPE JSONB USING metadata::jsonb;
ALTER TABLE tribe_members ALTER COLUMN metadata TYPE JSONB USING metadata::jsonb;

-- Add NOT NULL constraint with default empty object
ALTER TABLE activities ALTER COLUMN metadata SET DEFAULT '{}'::jsonb;
ALTER TABLE lists ALTER COLUMN metadata SET DEFAULT '{}'::jsonb;
ALTER TABLE list_items ALTER COLUMN metadata SET DEFAULT '{}'::jsonb;
ALTER TABLE tribes ALTER COLUMN metadata SET DEFAULT '{}'::jsonb;
ALTER TABLE tribe_members ALTER COLUMN metadata SET DEFAULT '{}'::jsonb;

ALTER TABLE activities ALTER COLUMN metadata SET NOT NULL;
ALTER TABLE lists ALTER COLUMN metadata SET NOT NULL;
ALTER TABLE list_items ALTER COLUMN metadata SET NOT NULL;
ALTER TABLE tribes ALTER COLUMN metadata SET NOT NULL;
ALTER TABLE tribe_members ALTER COLUMN metadata SET NOT NULL;

-- Add check constraints to ensure valid JSON
ALTER TABLE activities ADD CONSTRAINT activities_metadata_is_object CHECK (jsonb_typeof(metadata) = 'object');
ALTER TABLE lists ADD CONSTRAINT lists_metadata_is_object CHECK (jsonb_typeof(metadata) = 'object');
ALTER TABLE list_items ADD CONSTRAINT list_items_metadata_is_object CHECK (jsonb_typeof(metadata) = 'object');
ALTER TABLE tribes ADD CONSTRAINT tribes_metadata_is_object CHECK (jsonb_typeof(metadata) = 'object');
ALTER TABLE tribe_members ADD CONSTRAINT tribe_members_metadata_is_object CHECK (jsonb_typeof(metadata) = 'object'); 