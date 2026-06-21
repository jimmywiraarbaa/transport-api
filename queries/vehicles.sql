-- name: ListVehicles :many
SELECT id, user_id, vehicle_type_id, plate_number, brand, model, year, current_odometer_km, notes, created_at, updated_at, initial_odometer_km
FROM vehicles
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountVehicles :one
SELECT count(*) FROM vehicles WHERE user_id = $1;

-- name: GetVehicle :one
SELECT id, user_id, vehicle_type_id, plate_number, brand, model, year, current_odometer_km, notes, created_at, updated_at, initial_odometer_km
FROM vehicles
WHERE id = $1 AND user_id = $2;

-- name: CreateVehicle :one
INSERT INTO vehicles (user_id, vehicle_type_id, plate_number, brand, model, year, current_odometer_km, initial_odometer_km, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, user_id, vehicle_type_id, plate_number, brand, model, year, current_odometer_km, notes, created_at, updated_at, initial_odometer_km;

-- name: UpdateVehicle :one
UPDATE vehicles
SET vehicle_type_id = $3,
    plate_number    = $4,
    brand           = $5,
    model           = $6,
    year            = $7,
    notes           = $8
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, vehicle_type_id, plate_number, brand, model, year, current_odometer_km, notes, created_at, updated_at, initial_odometer_km;

-- name: UpdateVehicleOdometerIfHigher :one
UPDATE vehicles
SET current_odometer_km = $3
WHERE id = $1 AND user_id = $2 AND $3 > current_odometer_km
RETURNING id, user_id, vehicle_type_id, plate_number, brand, model, year, current_odometer_km, notes, created_at, updated_at, initial_odometer_km;

-- name: DeleteVehicle :exec
DELETE FROM vehicles WHERE id = $1 AND user_id = $2;
