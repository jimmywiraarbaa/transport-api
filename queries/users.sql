-- name: CreateUser :one
INSERT INTO users (email, username, password_hash, name)
VALUES ($1, $2, $3, $4)
RETURNING id, email, username, password_hash, name, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, email, username, password_hash, name, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT id, email, username, password_hash, name, created_at, updated_at
FROM users
WHERE username = $1;
