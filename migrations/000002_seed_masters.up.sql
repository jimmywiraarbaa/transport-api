-- 000002_seed_masters.up.sql
-- Seed master data: vehicle types, maintenance parts, and example schedule rules.

-- ── vehicle types ─────────────────────────────────────
INSERT INTO vehicle_types (name, slug) VALUES
    ('Motor', 'motor'),
    ('Mobil', 'mobil'),
    ('Truk', 'truk')
ON CONFLICT (slug) DO NOTHING;

-- ── maintenance parts ─────────────────────────────────
-- slugs are referenced by schedule_rules below; keep stable.
INSERT INTO maintenance_parts (name, slug, category, description) VALUES
    ('Oli Mesin',          'oli-mesin',          'Pelumas',  'Penggantian oli mesin'),
    ('Filter Oli',         'filter-oli',         'Filter',   'Filter oli mesin'),
    ('Filter Udara',       'filter-udara',       'Filter',   'Filter udara intake'),
    ('Oli Transmisi',      'oli-transmisi',      'Pelumas',  'Penggantian oli gardan/transmisi'),
    ('Kampas Rem',         'kampas-rem',         'Rem',      'Kampas rem depan/belakang'),
    ('Busi',               'busi',               'Pengapian','Busi mesin'),
    ('Ban',                'ban',                'Ban',      'Penggantian ban'),
    ('Aki',                'aki',                'Kelistrikan','Aki kendaraan'),
    ('Rantai Keteng',      'rantai-keteng',      'Transmisi','Rantai dan keteng (motor)'),
    ('Coolant',            'coolant',            'Pendingin','Cairan pendingin radiator')
ON CONFLICT (slug) DO NOTHING;

-- ── schedule rules (example intervals) ────────────────
-- Using subqueries on slug so the migration is id-stable.
-- interval_km, interval_days, trigger_mode per (part, vehicle_type).
INSERT INTO schedule_rules (part_id, vehicle_type_id, interval_km, interval_days, trigger_mode, notes)
SELECT p.id, vt.id, v.km, v.days, v.mode::trigger_mode, v.notes
FROM (VALUES
    -- slug,              type_slug, km,    days, mode,        notes
    ('oli-mesin',         'motor',   4000,  120,  'or',        'Oli mesin motor: 4000 km / 4 bulan'),
    ('oli-mesin',         'mobil',   10000, 180,  'or',        'Oli mesin mobil: 10.000 km / 6 bulan'),
    ('filter-oli',        'motor',   8000,  240,  'or',        'Filter oli motor'),
    ('filter-oli',        'mobil',   20000, 365,  'or',        'Filter oli mobil'),
    ('filter-udara',      'motor',   12000, 365,  'or',        'Filter udara motor'),
    ('filter-udara',      'mobil',   20000, 365,  'or',        'Filter udara mobil'),
    ('oli-transmisi',     'motor',   20000, 730,  'or',        'Oli transmisi/gardan motor'),
    ('oli-transmisi',     'mobil',   40000, 730,  'or',        'Oli transmisi mobil'),
    ('kampas-rem',        'motor',   15000, NULL, 'km_only',   'Periksa/kampas rem motor'),
    ('kampas-rem',        'mobil',   25000, NULL, 'km_only',   'Kampas rem mobil'),
    ('busi',              'motor',   12000, 365,  'or',        'Busi motor'),
    ('busi',              'mobil',   30000, 730,  'or',        'Busi mobil'),
    ('ban',               'motor',   20000, NULL, 'km_only',   'Ban motor'),
    ('ban',               'mobil',   40000, 1825, 'or',        'Ban mobil (max 5 thn)'),
    ('aki',               'motor',   NULL,  730,  'date_only', 'Aki motor tiap 2 tahun'),
    ('aki',               'mobil',   NULL,  730,  'date_only', 'Aki mobil tiap 2 tahun'),
    ('rantai-keteng',     'motor',   15000, 365,  'or',        'Rantai-keteng motor'),
    ('coolant',           'mobil',   40000, 730,  'or',        'Coolant mobil')
) AS v(slug, type_slug, km, days, mode, notes)
JOIN maintenance_parts p   ON p.slug = v.slug
JOIN vehicle_types vt      ON vt.slug = v.type_slug
ON CONFLICT (part_id, vehicle_type_id) DO NOTHING;
