-- name: ListMaintenanceParts :many
SELECT id, name, slug, category, description, created_at, updated_at
FROM maintenance_parts
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: CountMaintenanceParts :one
SELECT count(*) FROM maintenance_parts;

-- name: GetMaintenancePart :one
SELECT id, name, slug, category, description, created_at, updated_at
FROM maintenance_parts
WHERE id = $1;

-- name: CreateMaintenancePart :one
INSERT INTO maintenance_parts (name, slug, category, description)
VALUES ($1, $2, $3, $4)
RETURNING id, name, slug, category, description, created_at, updated_at;

-- name: UpdateMaintenancePart :one
UPDATE maintenance_parts
SET name = $2, slug = $3, category = $4, description = $5
WHERE id = $1
RETURNING id, name, slug, category, description, created_at, updated_at;

-- name: DeleteMaintenancePart :exec
DELETE FROM maintenance_parts WHERE id = $1;
