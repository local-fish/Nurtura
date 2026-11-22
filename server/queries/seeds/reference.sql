-- Immunisation schedule.
--
-- Note: these aren't final, just random numbers i got an LLM to generate. For the future, data still has to be verified
-- against a real national schedule (e.g. IDAI / Kemenkes) before this is shown to anyone as medical guidance.
--
-- The growth percentile curves that used to live in this file are now in
-- growth_reference.sql, generated from the published WHO Child Growth Standards.

BEGIN;

INSERT INTO vaccines (name, dose_number, recommended_age_months) VALUES
  ('Hepatitis B', 1, 0),
  ('Hepatitis B', 2, 1),
  ('Rotavirus',   1, 2),
  ('DTaP',        1, 2),
  ('Hib',         1, 2),
  ('PCV',         1, 2),
  ('IPV (Polio)', 1, 2),
  ('DTaP',        2, 4),
  ('DTaP',        3, 6),
  ('Hepatitis B', 3, 6),
  ('MMR',         1, 12),
  ('Varicella',   1, 12),
  ('Hepatitis A', 1, 12),
  ('DTaP',        4, 15),
  ('MMR',         2, 48),
  ('DTaP',        5, 48)
ON CONFLICT (name, dose_number) DO NOTHING;

COMMIT;
