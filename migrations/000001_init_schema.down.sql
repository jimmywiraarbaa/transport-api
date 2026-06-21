-- 000001_init_schema.down.sql
DROP TABLE IF EXISTS maintenance_records;
DROP TABLE IF EXISTS schedule_rules;
DROP TABLE IF EXISTS maintenance_parts;
DROP TABLE IF EXISTS vehicles;
DROP TABLE IF EXISTS vehicle_types;
DROP TABLE IF EXISTS users;
DROP FUNCTION IF EXISTS set_updated_at();
DROP TYPE IF EXISTS trigger_mode;
