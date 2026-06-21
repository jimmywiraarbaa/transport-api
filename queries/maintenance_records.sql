-- name: ListMaintenanceRecords :many
SELECT mr.id, mr.vehicle_id, mr.part_id, mr.user_id, mr.performed_at, mr.odometer_km, mr.cost, mr.technician, mr.notes, mr.created_at, mr.updated_at
FROM maintenance_records mr
WHERE mr.vehicle_id = $1 AND mr.user_id = $2
ORDER BY mr.performed_at DESC
LIMIT $3 OFFSET $4;

-- name: CountMaintenanceRecords :one
SELECT count(*) FROM maintenance_records WHERE vehicle_id = $1 AND user_id = $2;

-- name: ListAllMaintenanceRecords :many
SELECT id, vehicle_id, part_id, user_id, performed_at, odometer_km, cost, technician, notes, created_at, updated_at
FROM maintenance_records
WHERE user_id = $1
  AND (sqlc.narg('vehicle_id')::uuid IS NULL OR vehicle_id = sqlc.narg('vehicle_id'))
ORDER BY performed_at DESC
LIMIT $2 OFFSET $3;

-- name: CountAllMaintenanceRecords :one
SELECT count(*) FROM maintenance_records
WHERE user_id = $1
  AND (sqlc.narg('vehicle_id')::uuid IS NULL OR vehicle_id = sqlc.narg('vehicle_id'));

-- name: GetMaintenanceRecord :one
SELECT id, vehicle_id, part_id, user_id, performed_at, odometer_km, cost, technician, notes, created_at, updated_at
FROM maintenance_records
WHERE id = $1 AND user_id = $2;

-- name: CreateMaintenanceRecord :one
INSERT INTO maintenance_records (vehicle_id, part_id, user_id, performed_at, odometer_km, cost, technician, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, vehicle_id, part_id, user_id, performed_at, odometer_km, cost, technician, notes, created_at, updated_at;

-- name: UpdateMaintenanceRecord :one
UPDATE maintenance_records
SET part_id      = $3,
    performed_at = $4,
    odometer_km  = $5,
    cost         = $6,
    technician   = $7,
    notes        = $8
WHERE id = $1 AND user_id = $2
RETURNING id, vehicle_id, part_id, user_id, performed_at, odometer_km, cost, technician, notes, created_at, updated_at;

-- name: DeleteMaintenanceRecord :exec
DELETE FROM maintenance_records WHERE id = $1 AND user_id = $2;

-- Latest record per part for a vehicle — used by the alert computation.
-- name: LatestRecordPerPart :many
SELECT DISTINCT ON (part_id)
       part_id,
       id,
       vehicle_id,
       user_id,
       performed_at,
       odometer_km,
       cost,
       technician,
       notes,
       created_at,
       updated_at
FROM maintenance_records
WHERE vehicle_id = $1 AND user_id = $2
ORDER BY part_id, performed_at DESC, odometer_km DESC;
