-- 000003_add_username.up.sql
-- Add username column for username-based authentication.

ALTER TABLE users ADD COLUMN IF NOT EXISTS username TEXT NOT NULL DEFAULT '';

-- Ensure existing rows have unique usernames (derive from email prefix)
UPDATE users
SET username = split_part(email, '@', 1)
WHERE username = '' OR username IS NULL;

-- Fix remaining duplicates by appending a suffix
UPDATE users u
SET username = u.username || '_' || substr(md5(u.id::text), 1, 6)
WHERE u.id IN (
    SELECT id FROM (
        SELECT id, row_number() OVER (PARTITION BY username ORDER BY created_at) AS rn
        FROM users
    ) dup WHERE rn > 1
);

ALTER TABLE users DROP CONSTRAINT IF EXISTS uq_users_username;
ALTER TABLE users ADD CONSTRAINT uq_users_username UNIQUE (username);

-- Seed a default admin user (username: admin, password: admin123)
INSERT INTO users (email, username, password_hash, name)
VALUES ('admin@transport.local', 'admin', crypt('admin123', gen_salt('bf')), 'Admin')
ON CONFLICT (email) DO NOTHING;
