-- 000001_init_schema.up.sql
-- Transport Management — initial schema

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ── enums ─────────────────────────────────────────────
CREATE TYPE trigger_mode AS ENUM ('or', 'and', 'km_only', 'date_only');

-- ── users ─────────────────────────────────────────────
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL UNIQUE,
    username      TEXT NOT NULL UNIQUE DEFAULT '',
    password_hash TEXT NOT NULL,
    name          TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ── vehicle_types (master) ────────────────────────────
CREATE TABLE vehicle_types (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    slug       TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ── vehicles ──────────────────────────────────────────
CREATE TABLE vehicles (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vehicle_type_id     UUID NOT NULL REFERENCES vehicle_types(id) ON DELETE RESTRICT,
    plate_number        TEXT NOT NULL,
    brand               TEXT NOT NULL DEFAULT '',
    model               TEXT NOT NULL DEFAULT '',
    year                INT,
    current_odometer_km INT NOT NULL DEFAULT 0 CHECK (current_odometer_km >= 0),
    notes               TEXT NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_vehicles_user_id ON vehicles(user_id);
CREATE UNIQUE INDEX uq_vehicles_user_plate ON vehicles(user_id, plate_number);

-- ── maintenance_parts (master) ────────────────────────
CREATE TABLE maintenance_parts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    slug        TEXT NOT NULL UNIQUE,
    category    TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ── schedule_rules (master) ───────────────────────────
CREATE TABLE schedule_rules (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    part_id         UUID NOT NULL REFERENCES maintenance_parts(id) ON DELETE CASCADE,
    vehicle_type_id UUID NOT NULL REFERENCES vehicle_types(id) ON DELETE CASCADE,
    interval_km     INT CHECK (interval_km IS NULL OR interval_km > 0),
    interval_days   INT CHECK (interval_days IS NULL OR interval_days > 0),
    trigger_mode    trigger_mode NOT NULL DEFAULT 'or',
    notes           TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_schedule_interval CHECK (
        interval_km IS NOT NULL OR interval_days IS NOT NULL
    )
);

CREATE UNIQUE INDEX uq_schedule_rule_part_type ON schedule_rules(part_id, vehicle_type_id);

-- ── maintenance_records ───────────────────────────────
CREATE TABLE maintenance_records (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id   UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    part_id      UUID NOT NULL REFERENCES maintenance_parts(id) ON DELETE RESTRICT,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    performed_at DATE NOT NULL,
    odometer_km  INT NOT NULL CHECK (odometer_km >= 0),
    cost         NUMERIC(14, 2) NOT NULL DEFAULT 0,
    technician   TEXT NOT NULL DEFAULT '',
    notes        TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_records_vehicle_id ON maintenance_records(vehicle_id);
CREATE INDEX idx_records_vehicle_part ON maintenance_records(vehicle_id, part_id, performed_at DESC);

-- ── updated_at trigger ────────────────────────────────
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
DECLARE t TEXT;
BEGIN
    FOR t IN SELECT unnest(ARRAY[
        'users','vehicle_types','vehicles','maintenance_parts','schedule_rules','maintenance_records'
    ])
    LOOP
        EXECUTE format(
            'CREATE TRIGGER trg_%s_updated BEFORE UPDATE ON %I
             FOR EACH ROW EXECUTE FUNCTION set_updated_at();', t, t
        );
    END LOOP;
END $$;
