-- Drop triggers
DROP TRIGGER IF EXISTS update_activity_photos_updated_at ON activity_photos;
DROP TRIGGER IF EXISTS update_activity_sharing_updated_at ON activity_sharing;
DROP TRIGGER IF EXISTS update_activities_updated_at ON activities;

-- Drop tables
DROP TABLE IF EXISTS activity_photos;
DROP TABLE IF EXISTS activity_sharing;
DROP TABLE IF EXISTS activities;

-- Remove last_login column from users
ALTER TABLE users DROP COLUMN IF EXISTS last_login; 