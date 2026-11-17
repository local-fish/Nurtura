import { Temporal } from '@js-temporal/polyfill';

/** Completed months between a YYYY-MM-DD date of birth and today. */
export function ageMonthsFromDob(dob: string): number {
	const birth = Temporal.PlainDate.from(dob);
	const today = Temporal.Now.plainDateISO();
	if (Temporal.PlainDate.compare(birth, today) >= 0) return 0;
	const d = birth.until(today, { largestUnit: 'year' });
	return d.years * 12 + d.months;
}

export function formatAge(months: number): string {
	if (months < 1) return 'Baru lahir';
	if (months < 12) return `${months} bulan`;
	const years = Math.floor(months / 12);
	const rest = months % 12;
	return rest === 0 ? `${years} tahun` : `${years} tahun ${rest} bulan`;
}
