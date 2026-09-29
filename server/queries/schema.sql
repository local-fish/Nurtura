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

ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE patients ENABLE ROW LEVEL SECURITY;
ALTER TABLE metrics ENABLE ROW LEVEL SECURITY;
ALTER TABLE records ENABLE ROW LEVEL SECURITY;
ALTER TABLE logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE vaccines ENABLE ROW LEVEL SECURITY;
ALTER TABLE patient_vaccines ENABLE ROW LEVEL SECURITY;
ALTER TABLE growth_reference ENABLE ROW LEVEL SECURITY;
COMMIT;
