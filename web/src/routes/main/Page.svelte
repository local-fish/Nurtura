<script lang="ts">
	import type { NavItem } from '../../layout/components/interfaces';
	import { Temporal } from '@js-temporal/polyfill';
	import GrowthCard from './components/GrowthCard.svelte';
	import Card from '../../components/Card.svelte';
	import Icon from '../../components/Icon.svelte';
	import Button from '../../components/Button.svelte';
	import Main from '../../layout/main.svelte';
	import Onboarding from '../onboarding/Page.svelte';
	import InputData from '../input/Page.svelte';
	import Grafik from '../grafik/Page.svelte';
	import QuickActions from './components/QuickActions.svelte';
	import {
		listPatients,
		getNormals,
		listVaccines,
		listLogs,
		getActivity,
		completeVaccine,
		type Patient,
		type Normals,
		type VaccineStatus,
		type LogEntry,
		type Band
	} from '$src/lib/api';
	import { formatAge } from '$src/lib/age';

	let { onLogout = () => {} }: { onLogout?: () => void } = $props();

	let active = $state('beranda');

	let patients: Patient[] = $state([]);
	let activeId = $state('');
	let addingChild = $state(false);
	let booted = $state(false);

	let normals: Normals | null = $state(null);
	let vaccines: VaccineStatus[] = $state([]);
	let logs: LogEntry[] = $state([]);
	let lastActivityAt: string | null = $state(null);

	let loading = $state(true);
	let error = $state('');
	let markingVaccine = $state(false);

	const METRIC_META: Record<string, { label: string; unit: string }> = {
		height: { label: 'Tinggi', unit: 'cm' },
		weight: { label: 'Berat', unit: 'kg' },
		head_circumference: { label: 'Lingkar Kepala', unit: 'cm' }
	};
	const METRIC_ORDER = ['height', 'weight', 'head_circumference'];
	const BAND_LABEL: Record<Band, string> = {
		below_p3: 'Di bawah normal',
		p3_to_p97: 'Normal',
		above_p97: 'Di atas normal',
		'': 'Belum ada pembanding'
	};

	const dateFmt = new Intl.DateTimeFormat('id-ID', {
		day: 'numeric',
		month: 'long',
		year: 'numeric'
	});
	const shortDateFmt = new Intl.DateTimeFormat('id-ID', { day: 'numeric', month: 'short' });

	const patient = $derived(patients.find((p) => p.id === activeId) ?? null);

	function greeting() {
		const hour = Temporal.Now.zonedDateTimeISO(Temporal.Now.timeZoneId()).hour;
		if (hour < 11) return 'pagi';
		if (hour < 15) return 'siang';
		if (hour < 19) return 'sore';
		return 'malam';
	}

	/** Reloads the patient list and the active child's dashboard data.
	 *  `preferId` keeps a just-added or just-selected child active. */
	async function loadDashboard(preferId?: string) {
		loading = true;
		error = '';
		try {
			const list = await listPatients();
			patients = list;

			const wanted = preferId ?? activeId;
			activeId = list.some((p) => p.id === wanted) ? wanted : (list[0]?.id ?? '');
			if (!activeId) return;

			const [n, v, l, a] = await Promise.all([
				getNormals(activeId),
				listVaccines(activeId),
				listLogs(activeId, 5),
				getActivity(activeId)
			]);
			normals = n;
			vaccines = v.vaccines;
			logs = l;
			lastActivityAt = a.last_activity_at;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Gagal memuat data dasbor';
		} finally {
			loading = false;
			booted = true;
		}
	}

	loadDashboard();

	function childCreated(id: string) {
		addingChild = false;
		loadDashboard(id);
	}

	async function markVaccineDone() {
		if (!patient || !nextVaccine) return;
		markingVaccine = true;
		error = '';
		try {
			await completeVaccine(patient.id, nextVaccine.vaccine_id);
			const v = await listVaccines(patient.id);
			vaccines = v.vaccines;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Gagal menandai imunisasi';
		} finally {
			markingVaccine = false;
		}
	}

	function navigate(item: NavItem, e: MouseEvent) {
		e.preventDefault();
		active = item.id;
	}

	const growthMetrics = $derived.by(() => {
		if (!normals) return [];
		return METRIC_ORDER.map((key) => {
			const meta = METRIC_META[key];
			const c = normals!.latest[key];
			if (!c) return { label: meta.label, value: '-', unit: meta.unit, status: 'Belum ada data' };
			return {
				label: meta.label,
				value: c.value.toString(),
				unit: meta.unit,
				status: BAND_LABEL[c.band]
			};
		});
	});

	const growthUpdatedAt = $derived.by(() => {
		if (!normals) return '-';
		const dates = METRIC_ORDER.map((k) => normals!.latest[k]?.measured_at).filter(
			Boolean
		) as string[];
		if (!dates.length) return 'Belum ada data';
		const latest = dates.reduce((a, b) => (a > b ? a : b));
		return dateFmt.format(new Date(latest));
	});

	const nextVaccine = $derived(vaccines.find((v) => !v.completed) ?? null);

	const vaccineDueDate = $derived.by(() => {
		if (!nextVaccine || !patient) return null;
		const dob = Temporal.PlainDate.from(patient.date_of_birth);
		return dob.add({ months: nextVaccine.recommended_age_months });
	});

	const lastUpdatedLabel = $derived(
		lastActivityAt ? dateFmt.format(new Date(lastActivityAt)) : undefined
	);
</script>

{#if !booted && loading}
	<div class="boot">Memuat dasbor...</div>
{:else if error && !patients.length}
	<div class="boot">
		<p>{error}</p>
		<Button fill={false} onclick={() => loadDashboard()}>Coba lagi</Button>
	</div>
{:else if addingChild}
	<Onboarding mode="add" onCreated={childCreated} onCancel={() => (addingChild = false)} />
{:else if !patient}
	<Onboarding onCreated={childCreated} />
{:else}
	<Main
		{active}
		user={{
			name: patient.name,
			age: normals ? formatAge(normals.age_months) : '',
			avatar: ''
		}}
		{patients}
		{activeId}
		title={`Selamat ${greeting()}, ${patient.name}`}
		subtitle={Temporal.Now.zonedDateTimeISO(Temporal.Now.timeZoneId()).toLocaleString('id-ID', {
			day: 'numeric',
			month: 'long',
			year: 'numeric'
		})}
		lastUpdated={lastUpdatedLabel}
		onnavigate={navigate}
		onselectchild={(id: string) => loadDashboard(id)}
		onaddchild={() => (addingChild = true)}
		onlogout={onLogout}
	>
		{#if error}
			<div class="banner">{error}</div>
		{/if}

		{#if active === 'input'}
			<InputData
				{patient}
				{patients}
				onsaved={() => loadDashboard(activeId)}
				onselectchild={(id: string) => loadDashboard(id)}
			/>
		{:else if active === 'grafik'}
			<Grafik
				{patient}
				{normals}
				nextCheckup={nextVaccine
					? {
							label: `${nextVaccine.name} (dosis ${nextVaccine.dose_number})`,
							date: vaccineDueDate ? dateFmt.format(new Date(vaccineDueDate.toString())) : '-'
						}
					: null}
				onremind={() => (active = 'paspor')}
			/>
		{:else if active !== 'beranda'}
			<section class="dashboard">
				<Card bg="var(--surface)" border="1px solid var(--teal-100)">
					<h4>Belum tersedia</h4>
					<p class="muted">Halaman ini belum dibuat.</p>
				</Card>
			</section>
		{:else}
			<section class="dashboard">
				<div class="row top">
					<GrowthCard
						name={patient.name}
						updatedAt={growthUpdatedAt}
						metrics={growthMetrics}
						onopen={() => (active = 'grafik')}
					/>

					<Card bg="#E3AD53">
						<div class="imun-heading">
							<Icon name="calendar" size={24} />
							<strong style="font-size: 0.9rem, text-align: center">Imunisasi Berikutnya</strong>
						</div>
						{#if nextVaccine}
							<h1>{nextVaccine.name} (dosis {nextVaccine.dose_number})</h1>
							<div style="margin-bottom: 1rem">
								{#if nextVaccine.status === 'overdue'}
									Sudah jatuh tempo sejak usia {nextVaccine.recommended_age_months} bulan.
								{:else}
									Lakukan sebelum {vaccineDueDate
										? dateFmt.format(new Date(vaccineDueDate.toString()))
										: '-'}.
								{/if}
							</div>
							<Button icon="check" loading={markingVaccine} onclick={markVaccineDone}>
								Tandai Sudah Vaksin
							</Button>
						{:else}
							<h1>Semua selesai</h1>
							<div style="margin-bottom: 1rem">
								Semua imunisasi terjadwal sudah ditandai selesai.
							</div>
						{/if}
					</Card>
				</div>

				<div class="row bottom">
					<Card bg="var(--surface)" border="1px solid var(--teal-100)">
						<h4>Catatan Terakhir</h4>
						{#if logs.length === 0}
							<p class="muted">Belum ada catatan.</p>
						{:else}
							<ul class="logs">
								{#each logs as log (log.id)}
									<li>
										<span class="log-date">{shortDateFmt.format(new Date(log.created_at))}</span>
										<span class="log-note">{log.note}</span>
									</li>
								{/each}
							</ul>
						{/if}
					</Card>
					<Card bg="var(--surface)" border="1px solid var(--teal-100)">
						<QuickActions name={patient.name} onselect={(id: string) => (active = id)} />
					</Card>
				</div>
			</section>
		{/if}
	</Main>
{/if}

<style>
	.dashboard {
		display: flex;
		flex-direction: column;
		gap: 24px;
		margin: 0 auto;
		padding: 1rem;
	}
	.row {
		display: grid;
		gap: 24px;
		align-items: stretch;
	}
	.top {
		grid-template-columns: 2.2fr 1fr;
	}
	.bottom {
		grid-template-columns: 1.8fr 1fr;
	}
	.imun-heading {
		display: flex;
		flex-direction: row;
		gap: 8px;
		align-content: center;
	}
	.banner {
		padding: 0.75rem 1rem;
		margin: 0 1rem;
		border-radius: 0.75rem;
		background: #fbeaea;
		color: var(--err);
		font-size: 0.9rem;
	}
	.boot {
		display: grid;
		place-content: center;
		justify-items: center;
		gap: 1rem;
		height: 100%;
		padding: 2rem;
		background: var(--mint-50);
		color: var(--ink-soft);
		text-align: center;
	}
	.muted {
		color: var(--ink-soft);
		font-size: 0.9rem;
	}
	.logs {
		list-style: none;
		margin: 0.5rem 0 0;
		padding: 0;
		display: grid;
		gap: 10px;
	}
	.logs li {
		display: flex;
		gap: 10px;
		font-size: 0.9rem;
		align-items: baseline;
	}
	.log-date {
		flex-shrink: 0;
		color: var(--ink-soft);
		font-size: 0.8rem;
	}
	.log-note {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
</style>
