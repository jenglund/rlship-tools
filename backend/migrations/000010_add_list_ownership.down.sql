-- Drop indexes
DROP INDEX IF EXISTS idx_list_owners_owner;
DROP INDEX IF EXISTS idx_list_shares_tribe;
DROP INDEX IF EXISTS idx_list_shares_user;
DROP INDEX IF EXISTS idx_list_shares_expires;

-- Drop tables
DROP TABLE IF EXISTS list_shares;
DROP TABLE IF EXISTS list_owners; 