-- Schedule rules enriched with their maintenance part, scoped to a vehicle type.
-- Used by the alert computation to render part names alongside each rule.
-- name: ListRulesForVehicleType :many
SELECT
    sr.id              AS id,
    sr.part_id         AS part_id,
    sr.vehicle_type_id AS vehicle_type_id,
    sr.interval_km     AS interval_km,
    sr.interval_days   AS interval_days,
    sr.trigger_mode    AS trigger_mode,
    sr.notes           AS notes,
    sr.created_at      AS created_at,
    sr.updated_at      AS updated_at,
    p.name             AS part_name,
    p.slug             AS part_slug,
    p.category         AS part_category
FROM schedule_rules sr
JOIN maintenance_parts p ON p.id = sr.part_id
WHERE sr.vehicle_type_id = $1
ORDER BY p.name;
