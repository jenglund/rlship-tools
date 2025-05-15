-- Remove Firebase authentication fields from users table
ALTER TABLE users
    DROP CONSTRAINT users_provider_check,
    DROP COLUMN firebase_uid,
    DROP COLUMN provider,
    DROP COLUMN last_login,
    ALTER COLUMN email SET NOT NULL;

-- Remove index
DROP INDEX IF EXISTS idx_users_firebase_uid; 