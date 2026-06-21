-- 000004_add_initial_odometer.up.sql
-- Add initial_odometer_km to capture the odometer at vehicle creation.
-- This serves as the baseline for maintenance alert calculations when
-- no maintenance record exists yet.

ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS initial_odometer_km INT NOT NULL DEFAULT 0;

-- Backfill existing rows: use current_odometer_km as the initial value
-- since we don't have the original creation odometer for old data.
UPDATE vehicles SET initial_odometer_km = current_odometer_km WHERE initial_odometer_km = 0;
