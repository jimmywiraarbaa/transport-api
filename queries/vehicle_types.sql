-- name: ListVehicleTypes :many
SELECT id, name, slug, created_at, updated_at
FROM vehicle_types
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: CountVehicleTypes :one
SELECT count(*) FROM vehicle_types;

-- name: GetVehicleType :one
SELECT id, name, slug, created_at, updated_at
FROM vehicle_types
WHERE id = $1;

-- name: CreateVehicleType :one
INSERT INTO vehicle_types (name, slug)
VALUES ($1, $2)
RETURNING id, name, slug, created_at, updated_at;

-- name: UpdateVehicleType :one
UPDATE vehicle_types
SET name = $2, slug = $3
WHERE id = $1
RETURNING id, name, slug, created_at, updated_at;

-- name: DeleteVehicleType :exec
DELETE FROM vehicle_types WHERE id = $1;
