-- Remove check constraints
ALTER TABLE activities DROP CONSTRAINT activities_metadata_is_object;
ALTER TABLE lists DROP CONSTRAINT lists_metadata_is_object;
ALTER TABLE list_items DROP CONSTRAINT list_items_metadata_is_object;
ALTER TABLE tribes DROP CONSTRAINT tribes_metadata_is_object;
ALTER TABLE tribe_members DROP CONSTRAINT tribe_members_metadata_is_object;

-- Remove NOT NULL constraints and defaults
ALTER TABLE activities ALTER COLUMN metadata DROP NOT NULL;
ALTER TABLE lists ALTER COLUMN metadata DROP NOT NULL;
ALTER TABLE list_items ALTER COLUMN metadata DROP NOT NULL;
ALTER TABLE tribes ALTER COLUMN metadata DROP NOT NULL;
ALTER TABLE tribe_members ALTER COLUMN metadata DROP NOT NULL;

ALTER TABLE activities ALTER COLUMN metadata DROP DEFAULT;
ALTER TABLE lists ALTER COLUMN metadata DROP DEFAULT;
ALTER TABLE list_items ALTER COLUMN metadata DROP DEFAULT;
ALTER TABLE tribes ALTER COLUMN metadata DROP DEFAULT;
ALTER TABLE tribe_members ALTER COLUMN metadata DROP DEFAULT;

-- Convert metadata columns back to JSON type
ALTER TABLE activities ALTER COLUMN metadata TYPE JSON USING metadata::json;
ALTER TABLE lists ALTER COLUMN metadata TYPE JSON USING metadata::json;
ALTER TABLE list_items ALTER COLUMN metadata TYPE JSON USING metadata::json;
ALTER TABLE tribes ALTER COLUMN metadata TYPE JSON USING metadata::json;
ALTER TABLE tribe_members ALTER COLUMN metadata TYPE JSON USING metadata::json; 