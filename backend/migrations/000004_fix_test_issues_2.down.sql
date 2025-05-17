-- Revert foreign key constraints
ALTER TABLE activities
    DROP CONSTRAINT IF EXISTS activities_user_id_fkey,
    ADD CONSTRAINT activities_user_id_fkey
        FOREIGN KEY (user_id)
        REFERENCES users(id);

-- Drop test database
DROP DATABASE IF EXISTS testdb; 