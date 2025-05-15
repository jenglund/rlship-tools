-- Add Firebase authentication fields to users table
ALTER TABLE users
    ADD COLUMN firebase_uid VARCHAR(128) UNIQUE,
    ADD COLUMN provider VARCHAR(50),
    ADD COLUMN last_login TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN email DROP NOT NULL; -- Email might not be available immediately with some OAuth providers

-- Create index for Firebase UID lookups
CREATE INDEX idx_users_firebase_uid ON users(firebase_uid);

-- Add constraints
ALTER TABLE users
    ADD CONSTRAINT users_provider_check 
    CHECK (provider IN ('google', 'email', 'phone')); 