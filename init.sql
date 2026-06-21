-- init.sql
-- Transport Management — consolidated schema + seed data.
-- This file merges migrations 000001–000004 into a single script
-- suitable for Docker's /docker-entrypoint-initdb.d/ init.

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
    initial_odometer_km INT NOT NULL DEFAULT 0,
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

-- ══════════════════════════════════════════════════════
-- SEED DATA
-- ══════════════════════════════════════════════════════

-- ── vehicle types ─────────────────────────────────────
INSERT INTO vehicle_types (name, slug) VALUES
    ('Motor', 'motor'),
    ('Mobil', 'mobil'),
    ('Truk', 'truk')
ON CONFLICT (slug) DO NOTHING;

-- ── maintenance parts ─────────────────────────────────
INSERT INTO maintenance_parts (name, slug, category, description) VALUES
    ('Oli Mesin',          'oli-mesin',          'Pelumas',    'Penggantian oli mesin'),
    ('Filter Oli',         'filter-oli',         'Filter',     'Filter oli mesin'),
    ('Filter Udara',       'filter-udara',       'Filter',     'Filter udara intake'),
    ('Oli Transmisi',      'oli-transmisi',      'Pelumas',    'Penggantian oli gardan/transmisi'),
    ('Kampas Rem',         'kampas-rem',         'Rem',        'Kampas rem depan/belakang'),
    ('Busi',               'busi',               'Pengapian',  'Busi mesin'),
    ('Ban',                'ban',                'Ban',        'Penggantian ban'),
    ('Aki',                'aki',                'Kelistrikan','Aki kendaraan'),
    ('Rantai Keteng',      'rantai-keteng',      'Transmisi',  'Rantai dan keteng (motor)'),
    ('Coolant',            'coolant',            'Pendingin',  'Cairan pendingin radiator')
ON CONFLICT (slug) DO NOTHING;

-- ── schedule rules ────────────────────────────────────
INSERT INTO schedule_rules (part_id, vehicle_type_id, interval_km, interval_days, trigger_mode, notes)
SELECT p.id, vt.id, v.km, v.days, v.mode::trigger_mode, v.notes
FROM (VALUES
    -- slug,            type_slug, km,    days, mode,        notes
    ('oli-mesin',       'motor',   4000,  120,  'or',        'Oli mesin motor: 4000 km / 4 bulan'),
    ('oli-mesin',       'mobil',   10000, 180,  'or',        'Oli mesin mobil: 10.000 km / 6 bulan'),
    ('filter-oli',      'motor',   8000,  240,  'or',        'Filter oli motor'),
    ('filter-oli',      'mobil',   20000, 365,  'or',        'Filter oli mobil'),
    ('filter-udara',    'motor',   12000, 365,  'or',        'Filter udara motor'),
    ('filter-udara',    'mobil',   20000, 365,  'or',        'Filter udara mobil'),
    ('oli-transmisi',   'motor',   20000, 730,  'or',        'Oli transmisi/gardan motor'),
    ('oli-transmisi',   'mobil',   40000, 730,  'or',        'Oli transmisi mobil'),
    ('kampas-rem',      'motor',   15000, NULL, 'km_only',   'Periksa/kampas rem motor'),
    ('kampas-rem',      'mobil',   25000, NULL, 'km_only',   'Kampas rem mobil'),
    ('busi',            'motor',   12000, 365,  'or',        'Busi motor'),
    ('busi',            'mobil',   30000, 730,  'or',        'Busi mobil'),
    ('ban',             'motor',   20000, NULL, 'km_only',   'Ban motor'),
    ('ban',             'mobil',   40000, 1825, 'or',        'Ban mobil (max 5 thn)'),
    ('aki',             'motor',   NULL,  730,  'date_only', 'Aki motor tiap 2 tahun'),
    ('aki',             'mobil',   NULL,  730,  'date_only', 'Aki mobil tiap 2 tahun'),
    ('rantai-keteng',   'motor',   15000, 365,  'or',        'Rantai-keteng motor'),
    ('coolant',         'mobil',   40000, 730,  'or',        'Coolant mobil')
) AS v(slug, type_slug, km, days, mode, notes)
JOIN maintenance_parts p ON p.slug = v.slug
JOIN vehicle_types vt    ON vt.slug = v.type_slug
ON CONFLICT (part_id, vehicle_type_id) DO NOTHING;

-- ── default admin user ────────────────────────────────
-- username: admin, password: admin123
INSERT INTO users (email, username, password_hash, name)
VALUES ('admin@transport.local', 'admin', crypt('admin123', gen_salt('bf')), 'Admin')
ON CONFLICT (email) DO NOTHING;
