-- Demo data for local development: one account with two children, their growth
-- records, notes, and immunisation history. Everything here is made up.
--
--   login: demo@nurtura.app
--   password: demo12345678
--
-- !! LOCAL / DEV ONLY. This creates an account whose password is published in
-- !! this file. Never load it into a deployment that is reachable by anyone else.
--
-- Requires metrics.sql, growth_reference.sql and reference.sql to be loaded first:
-- the growth records are derived from the WHO P50 curves (see below) and the
-- immunisation history joins against the seeded vaccine schedule.
--
-- Re-runnable: it deletes the demo account first, which cascades to that
-- account's patients, records, logs and immunisations. Nothing else is touched.

BEGIN;

DELETE FROM users WHERE email = 'demo@nurtura.app';

INSERT INTO users (id, email, password_hash) VALUES
  ('11111111-1111-4111-8111-111111111111', 'demo@nurtura.app',
   '$2a$08$z3.J/2wpVoK4rBspfAeCzuEk.2KBiaQoUl9LK9E9NStytOml8s.tW');

-- Ages are relative to CURRENT_DATE so the demo keeps showing an 8-month-old and
-- a 2.5-year-old however long from now it is loaded.
INSERT INTO patients (id, user_id, name, date_of_birth, sex) VALUES
  ('22222222-2222-4222-8222-222222222222', '11111111-1111-4111-8111-111111111111',
   'Aura', (CURRENT_DATE - INTERVAL '8 months')::date, 'F'),
  ('33333333-3333-4333-8333-333333333333', '11111111-1111-4111-8111-111111111111',
   'Bima', (CURRENT_DATE - INTERVAL '30 months')::date, 'M');

-- Growth records, expressed as a multiple of the WHO median for that sex, metric
-- and age rather than as raw numbers. That keeps every point inside the normal
-- band no matter what the reference curves say, so the chart always demos well.
-- Aura tracks a little above the median, Bima a little below.
WITH plan(patient_id, metric, age_months, factor) AS (
  VALUES
    -- Aura: weight
    ('22222222-2222-4222-8222-222222222222'::uuid, 'weight',              0, 1.00),
    ('22222222-2222-4222-8222-222222222222'::uuid, 'weight',              1, 1.02),
    ('22222222-2222-4222-8222-222222222222'::uuid, 'weight',              2, 1.03),
    ('22222222-2222-4222-8222-222222222222'::uuid, 'weight',              4, 1.04),
    ('22222222-2222-4222-8222-222222222222'::uuid, 'weight',              6, 1.05),
    ('22222222-2222-4222-8222-222222222222'::uuid, 'weight',              8, 1.06),
    -- Aura: height
    ('22222222-2222-4222-8222-222222222222'::uuid, 'height',              0, 1.00),
    ('22222222-2222-4222-8222-222222222222'::uuid, 'height',              2, 1.01),
    ('22222222-2222-4222-8222-222222222222'::uuid, 'height',              4, 1.01),
    ('22222222-2222-4222-8222-222222222222'::uuid, 'height',              6, 1.02),
    ('22222222-2222-4222-8222-222222222222'::uuid, 'height',              8, 1.02),
    -- Aura: head circumference
    ('22222222-2222-4222-8222-222222222222'::uuid, 'head_circumference',  0, 1.00),
    ('22222222-2222-4222-8222-222222222222'::uuid, 'head_circumference',  4, 1.01),
    ('22222222-2222-4222-8222-222222222222'::uuid, 'head_circumference',  8, 1.01),
    -- Bima: weight
    ('33333333-3333-4333-8333-333333333333'::uuid, 'weight',              0, 1.00),
    ('33333333-3333-4333-8333-333333333333'::uuid, 'weight',              2, 0.99),
    ('33333333-3333-4333-8333-333333333333'::uuid, 'weight',              4, 0.98),
    ('33333333-3333-4333-8333-333333333333'::uuid, 'weight',              6, 0.98),
    ('33333333-3333-4333-8333-333333333333'::uuid, 'weight',              9, 0.97),
    ('33333333-3333-4333-8333-333333333333'::uuid, 'weight',             12, 0.97),
    ('33333333-3333-4333-8333-333333333333'::uuid, 'weight',             18, 0.96),
    ('33333333-3333-4333-8333-333333333333'::uuid, 'weight',             24, 0.96),
    ('33333333-3333-4333-8333-333333333333'::uuid, 'weight',             30, 0.95),
    -- Bima: height
    ('33333333-3333-4333-8333-333333333333'::uuid, 'height',              0, 1.00),
    ('33333333-3333-4333-8333-333333333333'::uuid, 'height',              6, 0.99),
    ('33333333-3333-4333-8333-333333333333'::uuid, 'height',             12, 0.99),
    ('33333333-3333-4333-8333-333333333333'::uuid, 'height',             18, 0.98),
    ('33333333-3333-4333-8333-333333333333'::uuid, 'height',             24, 0.98),
    ('33333333-3333-4333-8333-333333333333'::uuid, 'height',             30, 0.98),
    -- Bima: head circumference
    ('33333333-3333-4333-8333-333333333333'::uuid, 'head_circumference',  0, 1.00),
    ('33333333-3333-4333-8333-333333333333'::uuid, 'head_circumference',  6, 0.99),
    ('33333333-3333-4333-8333-333333333333'::uuid, 'head_circumference', 12, 0.99),
    ('33333333-3333-4333-8333-333333333333'::uuid, 'head_circumference', 24, 0.99)
)
INSERT INTO records (patient_id, metric, value, measured_at)
SELECT pl.patient_id,
       pl.metric,
       round(gr.p50 * pl.factor, 1),
       ((p.date_of_birth + (pl.age_months || ' months')::interval)::date + TIME '09:00')::timestamptz
FROM plan pl
JOIN patients p ON p.id = pl.patient_id
JOIN growth_reference gr
  ON gr.sex = p.sex AND gr.metric = pl.metric AND gr.age_months = pl.age_months;

INSERT INTO logs (patient_id, note, created_at) VALUES
  ('22222222-2222-4222-8222-222222222222', 'Mulai MPASI bubur wortel, habis setengah porsi.', now() - INTERVAL '2 days'),
  ('22222222-2222-4222-8222-222222222222', 'Tidur malam lebih nyenyak, bangun sekali saja.', now() - INTERVAL '6 days'),
  ('22222222-2222-4222-8222-222222222222', 'Sudah bisa duduk sendiri tanpa disangga.', now() - INTERVAL '13 days'),
  ('22222222-2222-4222-8222-222222222222', 'Demam ringan setelah imunisasi, turun keesokan harinya.', now() - INTERVAL '25 days'),
  ('33333333-3333-4333-8333-333333333333', 'Sudah hafal nama warna merah dan biru.', now() - INTERVAL '4 days'),
  ('33333333-3333-4333-8333-333333333333', 'Nafsu makan menurun, dicoba porsi kecil tapi sering.', now() - INTERVAL '11 days'),
  ('33333333-3333-4333-8333-333333333333', 'Kontrol gigi pertama, tidak ada keluhan.', now() - INTERVAL '21 days');

-- Immunisations already given. The cutoff is set below each child's current age
-- so there is always something still due: Aura has her 6-month doses outstanding
-- (overdue at 8 months), Bima has the 48-month doses coming up.
WITH done(patient_id, up_to_age_months) AS (
  VALUES
    ('22222222-2222-4222-8222-222222222222'::uuid, 4),
    ('33333333-3333-4333-8333-333333333333'::uuid, 15)
)
INSERT INTO patient_vaccines (patient_id, vaccine_id, completed_at)
SELECT d.patient_id,
       v.id,
       ((p.date_of_birth + (v.recommended_age_months || ' months')::interval)::date + TIME '10:00')::timestamptz
FROM done d
JOIN patients p ON p.id = d.patient_id
JOIN vaccines v ON v.recommended_age_months <= d.up_to_age_months;

COMMIT;
