-- 000004_add_initial_odometer.down.sql

ALTER TABLE vehicles DROP COLUMN IF EXISTS initial_odometer_km;
