-- Note: these aren't final, just random numbers i got an LLM to generate. For the future, data still has to be verified
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

INSERT INTO growth_reference (sex, age_months, metric, p3, p50, p97) VALUES
  ('M',  0,'height', 46.3, 49.9, 53.4), ('M',  0,'weight',  2.5,  3.3,  4.3),
  ('M',  6,'height', 63.6, 67.6, 71.6), ('M',  6,'weight',  6.4,  7.9,  9.7),
  ('M', 12,'height', 71.3, 75.7, 80.2), ('M', 12,'weight',  7.8,  9.6, 11.8),
  ('M', 24,'height', 82.1, 87.1, 92.0), ('M', 24,'weight',  9.8, 12.2, 15.1),
  ('M', 36,'height', 89.1, 96.1,103.1), ('M', 36,'weight', 11.4, 14.3, 17.8),
  ('M', 48,'height', 95.4,103.3,111.2), ('M', 48,'weight', 12.9, 16.3, 20.5),
  ('M', 60,'height',101.2,110.0,118.7), ('M', 60,'weight', 14.3, 18.3, 23.2),
  ('F',  0,'height', 45.6, 49.1, 52.7), ('F',  0,'weight',  2.4,  3.2,  4.2),
  ('F',  6,'height', 61.5, 65.7, 70.0), ('F',  6,'weight',  5.8,  7.3,  9.2),
  ('F', 12,'height', 69.2, 74.0, 78.9), ('F', 12,'weight',  7.1,  8.9, 11.3),
  ('F', 24,'height', 80.3, 85.7, 91.0), ('F', 24,'weight',  9.2, 11.5, 14.6),
  ('F', 36,'height', 88.0, 95.1,102.2), ('F', 36,'weight', 10.8, 13.9, 17.8),
  ('F', 48,'height', 94.6,102.7,110.8), ('F', 48,'weight', 12.3, 16.1, 20.9),
  ('F', 60,'height',100.5,109.4,118.4), ('F', 60,'weight', 13.7, 18.2, 23.5)
ON CONFLICT DO NOTHING;

COMMIT;
