-- Drop indexes
DROP INDEX IF EXISTS idx_activities_type;
DROP INDEX IF EXISTS idx_activities_visibility;
DROP INDEX IF EXISTS idx_activity_owners_owner;
DROP INDEX IF EXISTS idx_activity_shares_tribe;
DROP INDEX IF EXISTS idx_activity_shares_user;
DROP INDEX IF EXISTS idx_activity_shares_expires;

-- Drop tables
DROP TABLE IF EXISTS activity_shares;
DROP TABLE IF EXISTS activity_owners;
DROP TABLE IF EXISTS activities;

-- Drop enums
DROP TYPE IF EXISTS visibility_type;
DROP TYPE IF EXISTS activity_type; 