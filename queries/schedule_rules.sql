-- name: ListScheduleRules :many
SELECT id, part_id, vehicle_type_id, interval_km, interval_days, trigger_mode, notes, created_at, updated_at
FROM schedule_rules
ORDER BY created_at
LIMIT $1 OFFSET $2;

-- name: CountScheduleRules :one
SELECT count(*) FROM schedule_rules;

-- name: GetScheduleRule :one
SELECT id, part_id, vehicle_type_id, interval_km, interval_days, trigger_mode, notes, created_at, updated_at
FROM schedule_rules
WHERE id = $1;

-- name: ListScheduleRulesByVehicleType :many
SELECT id, part_id, vehicle_type_id, interval_km, interval_days, trigger_mode, notes, created_at, updated_at
FROM schedule_rules
WHERE vehicle_type_id = $1
ORDER BY created_at;

-- name: CreateScheduleRule :one
INSERT INTO schedule_rules (part_id, vehicle_type_id, interval_km, interval_days, trigger_mode, notes)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, part_id, vehicle_type_id, interval_km, interval_days, trigger_mode, notes, created_at, updated_at;

-- name: UpdateScheduleRule :one
UPDATE schedule_rules
SET part_id = $2,
    vehicle_type_id = $3,
    interval_km = $4,
    interval_days = $5,
    trigger_mode = $6,
    notes = $7
WHERE id = $1
RETURNING id, part_id, vehicle_type_id, interval_km, interval_days, trigger_mode, notes, created_at, updated_at;

-- name: DeleteScheduleRule :exec
DELETE FROM schedule_rules WHERE id = $1;
