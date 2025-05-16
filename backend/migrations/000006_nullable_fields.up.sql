-- Make required fields NOT NULL
ALTER TABLE activities ALTER COLUMN type SET NOT NULL;
ALTER TABLE activities ALTER COLUMN name SET NOT NULL;
ALTER TABLE activities ALTER COLUMN description SET NOT NULL;
ALTER TABLE activities ALTER COLUMN visibility SET NOT NULL;
ALTER TABLE activities ALTER COLUMN created_at SET NOT NULL;
ALTER TABLE activities ALTER COLUMN updated_at SET NOT NULL;

ALTER TABLE lists ALTER COLUMN type SET NOT NULL;
ALTER TABLE lists ALTER COLUMN name SET NOT NULL;
ALTER TABLE lists ALTER COLUMN description SET NOT NULL;
ALTER TABLE lists ALTER COLUMN visibility SET NOT NULL;
ALTER TABLE lists ALTER COLUMN sync_status SET NOT NULL;
ALTER TABLE lists ALTER COLUMN default_weight SET NOT NULL;
ALTER TABLE lists ALTER COLUMN created_at SET NOT NULL;
ALTER TABLE lists ALTER COLUMN updated_at SET NOT NULL;

ALTER TABLE list_items ALTER COLUMN name SET NOT NULL;
ALTER TABLE list_items ALTER COLUMN description SET NOT NULL;
ALTER TABLE list_items ALTER COLUMN weight SET NOT NULL;
ALTER TABLE list_items ALTER COLUMN available SET NOT NULL;
ALTER TABLE list_items ALTER COLUMN seasonal SET NOT NULL;
ALTER TABLE list_items ALTER COLUMN chosen_count SET NOT NULL;
ALTER TABLE list_items ALTER COLUMN use_count SET NOT NULL;
ALTER TABLE list_items ALTER COLUMN created_at SET NOT NULL;
ALTER TABLE list_items ALTER COLUMN updated_at SET NOT NULL;

ALTER TABLE tribes ALTER COLUMN name SET NOT NULL;
ALTER TABLE tribes ALTER COLUMN type SET NOT NULL;
ALTER TABLE tribes ALTER COLUMN description SET NOT NULL;
ALTER TABLE tribes ALTER COLUMN visibility SET NOT NULL;
ALTER TABLE tribes ALTER COLUMN created_at SET NOT NULL;
ALTER TABLE tribes ALTER COLUMN updated_at SET NOT NULL;

ALTER TABLE tribe_members ALTER COLUMN membership_type SET NOT NULL;
ALTER TABLE tribe_members ALTER COLUMN display_name SET NOT NULL;
ALTER TABLE tribe_members ALTER COLUMN created_at SET NOT NULL;
ALTER TABLE tribe_members ALTER COLUMN updated_at SET NOT NULL;

ALTER TABLE users ALTER COLUMN firebase_uid SET NOT NULL;
ALTER TABLE users ALTER COLUMN provider SET NOT NULL;
ALTER TABLE users ALTER COLUMN email SET NOT NULL;
ALTER TABLE users ALTER COLUMN name SET NOT NULL;
ALTER TABLE users ALTER COLUMN created_at SET NOT NULL;
ALTER TABLE users ALTER COLUMN updated_at SET NOT NULL;

-- Add constraints for location data completeness
ALTER TABLE list_items ADD CONSTRAINT location_data_complete 
    CHECK (
        (latitude IS NULL AND longitude IS NULL AND address IS NULL) OR
        (latitude IS NOT NULL AND longitude IS NOT NULL AND address IS NOT NULL)
    );

-- Add constraints for seasonal data completeness
ALTER TABLE list_items ADD CONSTRAINT seasonal_dates_complete
    CHECK (
        (NOT seasonal) OR
        (seasonal AND start_date IS NOT NULL AND end_date IS NOT NULL)
    );

-- Add constraints for guest membership expiration
ALTER TABLE tribe_members ADD CONSTRAINT guest_membership_expires
    CHECK (
        membership_type != 'guest' OR
        (membership_type = 'guest' AND expires_at IS NOT NULL)
    );

-- Add default values for boolean fields
ALTER TABLE list_items ALTER COLUMN available SET DEFAULT true;
ALTER TABLE list_items ALTER COLUMN seasonal SET DEFAULT false;
ALTER TABLE list_items ALTER COLUMN chosen_count SET DEFAULT 0;
ALTER TABLE list_items ALTER COLUMN use_count SET DEFAULT 0;

-- Add default values for numeric fields
ALTER TABLE lists ALTER COLUMN default_weight SET DEFAULT 1.0; 