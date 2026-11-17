INSERT INTO metrics (code, unit, min_value, max_value) VALUES
  ('height', 'cm', 20, 250),
  ('weight', 'kg', 0.3, 300),
  ('head_circumference', 'cm', 10, 70)
ON CONFLICT (code) DO NOTHING;
