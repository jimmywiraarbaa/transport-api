-- 000003_add_username.down.sql
-- Reverse: drop username column.

DELETE FROM users WHERE username = 'admin';

ALTER TABLE users DROP CONSTRAINT IF EXISTS uq_users_username;
ALTER TABLE users DROP COLUMN IF EXISTS username;
