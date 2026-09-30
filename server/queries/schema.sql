BEGIN;

CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email text NOT NULL UNIQUE,
  password_hash text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE patients (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name text NOT NULL,
  date_of_birth date NOT NULL,
  sex char(1) NOT NULL CHECK (sex IN ('M','F')),
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX patients_user_id_idx ON patients(user_id);

CREATE TABLE metrics(
  code text PRIMARY KEY,
  unit text NOT NULL,
  min_value numeric NOT NULL,
  max_value numeric NOT NULL,
  CHECK (min_value < max_value)
);

CREATE TABLE records (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  patient_id uuid NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
  metric text NOT NULL REFERENCES metrics(code),
  value numeric NOT NULL CHECK (value > 0),
  measured_at timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX records_patient_measured_idx ON records(patient_id, measured_at DESC);

CREATE TABLE logs (
  id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  patient_id uuid NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
  note       text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX logs_patient_created_idx ON logs(patient_id, created_at DESC);

CREATE TABLE vaccines (
  id serial PRIMARY KEY,
  name text NOT NULL,
  dose_number int NOT NULL DEFAULT 1,
  recommended_age_months int NOT NULL,
  UNIQUE (name, dose_number)
);

CREATE TABLE patient_vaccines (
  patient_id uuid NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
  vaccine_id int  NOT NULL REFERENCES vaccines(id),
  completed_at timestamptz,
  PRIMARY KEY (patient_id, vaccine_id)
);

CREATE TABLE growth_reference (
  sex char(1) NOT NULL CHECK (sex IN ('M','F')),
  age_months int NOT NULL,
  metric text NOT NULL REFERENCES metrics(code),
  p3 numeric NOT NULL,
  p50 numeric NOT NULL,
  p97 numeric NOT NULL,
  PRIMARY KEY (sex, age_months, metric)
);

CREATE TABLE patient_activity (
  patient_id uuid PRIMARY KEY REFERENCES patients(id) ON DELETE CASCADE,
  last_activity_at timestamptz NOT NULL DEFAULT now()
);
ALTER TABLE patient_activity ENABLE ROW LEVEL SECURITY;

CREATE FUNCTION touch_patient_child() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE pid uuid;
BEGIN
  pid := CASE WHEN TG_OP = 'DELETE' THEN OLD.patient_id ELSE NEW.patient_id END;
  INSERT INTO patient_activity (patient_id, last_activity_at)
  SELECT pid, now() WHERE EXISTS (SELECT 1 FROM patients WHERE id = pid)
  ON CONFLICT (patient_id) DO UPDATE SET last_activity_at = EXCLUDED.last_activity_at;
  RETURN NULL;
END $$;

CREATE FUNCTION touch_patient_self() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  INSERT INTO patient_activity (patient_id, last_activity_at) VALUES (NEW.id, now())
  ON CONFLICT (patient_id) DO UPDATE SET last_activity_at = EXCLUDED.last_activity_at;
  RETURN NULL;
END $$;

CREATE TRIGGER records_activity  AFTER INSERT OR UPDATE OR DELETE ON records FOR EACH ROW EXECUTE FUNCTION touch_patient_child();
CREATE TRIGGER logs_activity     AFTER INSERT OR UPDATE OR DELETE ON logs FOR EACH ROW EXECUTE FUNCTION touch_patient_child();
CREATE TRIGGER vaccines_activity AFTER INSERT OR UPDATE OR DELETE ON patient_vaccines FOR EACH ROW EXECUTE FUNCTION touch_patient_child();
CREATE TRIGGER patients_activity AFTER INSERT OR UPDATE ON patients FOR EACH ROW EXECUTE FUNCTION touch_patient_self();

ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE patients ENABLE ROW LEVEL SECURITY;
ALTER TABLE metrics ENABLE ROW LEVEL SECURITY;
ALTER TABLE records ENABLE ROW LEVEL SECURITY;
ALTER TABLE logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE vaccines ENABLE ROW LEVEL SECURITY;
ALTER TABLE patient_vaccines ENABLE ROW LEVEL SECURITY;
ALTER TABLE growth_reference ENABLE ROW LEVEL SECURITY;
COMMIT;
