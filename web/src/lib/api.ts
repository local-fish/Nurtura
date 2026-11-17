import { authFetch } from '$src/routes/login/auth';

async function json<T>(res: Response): Promise<T> {
	const data = await res.json().catch(() => ({}));
	if (!res.ok) {
		throw new Error(data?.error ?? data?.message ?? `Request failed (${res.status})`);
	}
	return data as T;
}

export interface Patient {
	id: string;
	name: string;
	date_of_birth: string;
	sex: 'M' | 'F';
	created_at: string;
}

export const listPatients = () => authFetch('/patients').then((r) => json<Patient[]>(r));

export const createPatient = (name: string, dateOfBirth: string, sex: 'M' | 'F') =>
	authFetch('/patients', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ name, date_of_birth: dateOfBirth, sex })
	}).then((r) => json<Patient>(r));

export interface MetricRecord {
	id: string;
	patient_id: string;
	metric: string;
	value: number;
	measured_at: string;
	created_at: string;
	updated_at: string;
}

export const createRecord = (
	patientId: string,
	metric: string,
	value: number,
	measuredAt?: string
) =>
	authFetch(`/patients/${patientId}/records`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ metric, value, measured_at: measuredAt ?? null })
	}).then((r) => json<MetricRecord>(r));

export interface Reference {
	reference_age_months: number;
	p3: number;
	p50: number;
	p97: number;
}

export type Band = 'below_p3' | 'p3_to_p97' | 'above_p97' | '';

export interface Comparison {
	value: number;
	measured_at: string;
	age_months_at_measurement: number;
	reference: Reference | null;
	band: Band;
}

export interface Normals {
	patient_id: string;
	sex: 'M' | 'F';
	age_months: number;
	reference: Record<string, Reference | null>;
	latest: Record<string, Comparison | null>;
}

export const getNormals = (patientId: string) =>
	authFetch(`/patients/${patientId}/normals`).then((r) => json<Normals>(r));

export interface VaccineStatus {
	vaccine_id: number;
	name: string;
	dose_number: number;
	recommended_age_months: number;
	completed: boolean;
	completed_at: string | null;
	status: 'completed' | 'overdue' | 'upcoming';
}

export interface VaccinesResponse {
	patient_id: string;
	age_months: number;
	vaccines: VaccineStatus[];
}

export const listVaccines = (patientId: string) =>
	authFetch(`/patients/${patientId}/vaccines`).then((r) => json<VaccinesResponse>(r));

export const completeVaccine = (patientId: string, vaccineId: number) =>
	authFetch(`/patients/${patientId}/vaccines/${vaccineId}/complete`, { method: 'POST' }).then((r) =>
		json<{ patient_id: string; vaccine_id: number; completed: boolean; completed_at: string }>(r)
	);

export interface LogEntry {
	id: string;
	patient_id: string;
	note: string;
	created_at: string;
}

export const listLogs = (patientId: string, limit = 5) =>
	authFetch(`/patients/${patientId}/logs?limit=${limit}`).then((r) => json<LogEntry[]>(r));

export const createLog = (patientId: string, note: string) =>
	authFetch(`/patients/${patientId}/logs`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ note })
	}).then((r) => json<LogEntry>(r));

/** Records for a patient, newest first. Optionally filtered to one metric. */
export const listRecords = (patientId: string, metric?: string) =>
	authFetch(
		`/patients/${patientId}/records${metric ? `?metric=${encodeURIComponent(metric)}` : ''}`
	).then((r) => json<MetricRecord[]>(r));

export const getActivity = (patientId: string) =>
	authFetch(`/patients/${patientId}/activity`).then((r) =>
		json<{ patient_id: string; last_activity_at: string | null }>(r)
	);
