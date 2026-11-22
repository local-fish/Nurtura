<script lang="ts">
	import GrowthChart from './components/GrowthChart.svelte';
	import Icon from '$src/components/Icon.svelte';
	import Button from '$src/components/Button.svelte';
	import { getGrowth, type Patient, type Normals, type Growth, type Band } from '$src/lib/api';

	let {
		patient,
		normals = null,
		nextCheckup = null,
		onremind = () => {}
	}: {
		patient: Patient;
		normals?: Normals | null;
		nextCheckup?: { label: string; date: string } | null;
		onremind?: () => void;
	} = $props();

	const METRICS = [
		{ key: 'weight', tab: 'Berat Badan', title: 'Berat badan menurut umur', axis: 'Berat (kg)' },
		{ key: 'height', tab: 'Tinggi Badan', title: 'Tinggi badan menurut umur', axis: 'Tinggi (cm)' },
		{
			key: 'head_circumference',
			tab: 'Lingkar Kepala',
			title: 'Lingkar kepala menurut umur',
			axis: 'Lingkar kepala (cm)'
		}
	];

	const RANGES = [
		{ months: 6, label: '6 bulan' },
		{ months: 12, label: '12 bulan' },
		{ months: 24, label: '2 tahun' },
		{ months: 60, label: '5 tahun' }
	];

	const BAND_LABEL: Record<Band, string> = {
		below_p3: 'Di bawah rentang normal',
		p3_to_p97: 'Normal · Standar WHO',
		above_p97: 'Di atas rentang normal',
		'': 'Belum ada pembanding'
	};

	const dateFmt = new Intl.DateTimeFormat('id-ID', {
		day: 'numeric',
		month: 'long',
		year: 'numeric'
	});

	let metric = $state('weight');
	let maxAge = $state(24);
	let data: Record<string, Growth> = $state({});
	let loading = $state(true);
	let error = $state('');
	let chart: GrowthChart | null = $state(null);

	const meta = $derived(METRICS.find((m) => m.key === metric) ?? METRICS[0]);
	const current = $derived(data[metric] ?? null);

	/** Loads all three metrics up front so switching tabs is instant and the
	 *  summary panel can show a delta for every metric, not just the open one. */
	async function load() {
		loading = true;
		error = '';
		try {
			const results = await Promise.all(METRICS.map((m) => getGrowth(patient.id, m.key)));
			data = Object.fromEntries(results.map((g) => [g.metric, g]));
			// Open on the narrowest range that still covers the child's age, so an
			// older child's latest measurements aren't cut off the right edge.
			const age = results[0]?.age_months ?? 0;
			maxAge = RANGES.find((r) => r.months >= age)?.months ?? RANGES[RANGES.length - 1].months;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Gagal memuat grafik pertumbuhan';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		// Re-fetch when the selected child changes.
		void patient.id;
		load();
	});

	/** Latest value and its change from the previous measurement. */
	function summary(key: string) {
		const g = data[key];
		const recs = g?.records ?? [];
		const last = recs.at(-1);
		const prev = recs.at(-2);
		return {
			value: last ? `${last.value}` : '-',
			unit: g?.unit ?? '',
			delta: last && prev ? last.value - prev.value : null
		};
	}

	const latest = $derived(current?.records.at(-1) ?? null);
	const sexLabel = $derived(patient.sex === 'F' ? 'Perempuan' : 'Laki-laki');
	const band = $derived(normals?.latest[metric]?.band ?? '');
	const measuredAt = $derived(latest ? dateFmt.format(new Date(latest.measured_at)) : null);

	const headline = $derived.by(() => {
		if (!latest) return 'Belum ada data pengukuran';
		if (band === 'p3_to_p97') return 'Pertumbuhan berada di rentang normal WHO';
		if (band === 'below_p3') return 'Pertumbuhan di bawah rentang normal WHO';
		if (band === 'above_p97') return 'Pertumbuhan di atas rentang normal WHO';
		return 'Belum ada kurva pembanding';
	});

	const blurb = $derived.by(() => {
		if (!latest) return `Tambahkan pengukuran ${meta.tab.toLowerCase()} untuk melihat grafiknya.`;
		const noun = meta.tab.toLowerCase();
		if (band === 'p3_to_p97') {
			return `${noun.charAt(0).toUpperCase() + noun.slice(1)} ${patient.name} berada pada rentang normal untuk anak ${sexLabel.toLowerCase()} usia ${current?.age_months ?? 0} bulan.`;
		}
		if (band === '') {
			return `Belum ada kurva standar WHO untuk ${noun} pada data ini.`;
		}
		return `${noun.charAt(0).toUpperCase() + noun.slice(1)} ${patient.name} berada di luar rentang normal. Konsultasikan dengan tenaga kesehatan.`;
	});

	function download() {
		const url = chart?.toDataURL();
		if (!url) return;
		const a = document.createElement('a');
		a.href = url;
		a.download = `grafik-${metric}-${patient.name.toLowerCase().replace(/\s+/g, '-')}.png`;
		a.click();
	}
</script>

<section class="grafik">
	<div class="chart-col">
		<div class="panel">
			<div class="toolbar">
				<div class="tabs" role="tablist" aria-label="Pilih metrik">
					{#each METRICS as m (m.key)}
						<button
							role="tab"
							aria-selected={metric === m.key}
							class:active={metric === m.key}
							onclick={() => (metric = m.key)}
						>
							{m.tab}
						</button>
					{/each}
				</div>
				<div class="tools">
					<select bind:value={maxAge} aria-label="Rentang usia">
						{#each RANGES as r (r.months)}
							<option value={r.months}>{r.label}</option>
						{/each}
					</select>
					<button class="download" onclick={download} disabled={!current}>
						<Icon name="download" size={16} />
						Unduh grafik
					</button>
				</div>
			</div>

			<header class="chart-head">
				<div>
					<h2>{meta.title}</h2>
					<p class="sub">WHO Baby Growth Standard · {sexLabel} (0–{maxAge} bulan)</p>
				</div>
				{#if latest}
					<div class="latest">
						<span>Data terbaru</span>
						<strong>{latest.value}{current?.unit}</strong>
					</div>
				{/if}
			</header>

			{#if loading}
				<div class="state">Memuat grafik...</div>
			{:else if error}
				<div class="state error">{error}</div>
			{:else if current && current.reference.length === 0 && current.records.length === 0}
				<div class="state">
					Belum ada data maupun kurva standar untuk {meta.tab.toLowerCase()}.
				</div>
			{:else if current}
				<GrowthChart
					bind:this={chart}
					reference={current.reference}
					records={current.records}
					unit={current.unit}
					label={`Data ${patient.name}`}
					caption={meta.axis}
					{maxAge}
				/>
				<ul class="legend">
					<li><span class="swatch outside"></span>Di luar rentang</li>
					<li><span class="swatch band"></span>Rentang normal (P3–P97)</li>
					<li><span class="swatch median"></span>Median (P50)</li>
					<li><span class="swatch child"></span>Data {patient.name}</li>
				</ul>
			{/if}
		</div>
	</div>

	<aside class="rail">
		<div class="status">
			<span class="status-icon"><Icon name="chart" size={20} /></span>
			<h3>{headline}</h3>
			<p>{blurb}</p>
			<span class="pill">
				<Icon name="check" size={14} />
				{BAND_LABEL[band]}
			</span>
		</div>

		<div class="panel small">
			<h4>Ringkasan pengukuran</h4>
			{#if measuredAt}
				<p class="sub">{measuredAt}</p>
			{/if}
			<ul class="summary">
				{#each METRICS as m (m.key)}
					{@const s = summary(m.key)}
					<li>
						<span class="s-label">{m.tab}</span>
						<div class="s-row">
							<strong>{s.value}{s.value === '-' ? '' : ` ${s.unit}`}</strong>
							{#if s.delta !== null}
								<span class="delta" class:down={s.delta < 0}>
									{s.delta > 0 ? '+' : ''}{s.delta.toFixed(1)}
								</span>
							{/if}
						</div>
					</li>
				{/each}
			</ul>
		</div>

		<div class="panel small">
			<div class="checkup">
				<span class="cal"><Icon name="calendar" size={18} /></span>
				<div>
					<h4>Pemeriksaan berikutnya</h4>
					{#if nextCheckup}
						<p class="sub">{nextCheckup.label} · {nextCheckup.date}</p>
					{:else}
						<p class="sub">Belum ada jadwal.</p>
					{/if}
				</div>
			</div>
			<Button fill={false} icon="info" onclick={onremind}>Atur pengingat</Button>
		</div>
	</aside>
</section>

<style>
	.grafik {
		display: grid;
		grid-template-columns: minmax(0, 2.4fr) minmax(280px, 1fr);
		gap: 24px;
		padding: 1rem;
	}

	.chart-col {
		min-width: 0;
	}

	.panel {
		padding: 20px 24px;
		border-radius: 1.5rem;
		border: 1px solid var(--teal-100);
		background: var(--surface);
	}
	.panel.small {
		padding: 18px 20px;
	}

	.toolbar {
		display: flex;
		flex-wrap: wrap;
		justify-content: space-between;
		align-items: center;
		gap: 12px;
	}

	.tabs {
		display: flex;
		gap: 8px;
	}
	.tabs button {
		padding: 9px 16px;
		border: 1px solid var(--line);
		border-radius: 999px;
		background: var(--surface);
		color: var(--ink-soft);
		font-size: 0.85rem;
		font-weight: 600;
		cursor: pointer;
	}
	.tabs button.active {
		background: var(--teal-700);
		border-color: var(--teal-700);
		color: #fff;
	}

	.tools {
		display: flex;
		gap: 8px;
		align-items: center;
	}
	.tools select,
	.download {
		height: 38px;
		padding: 0 14px;
		border: 1px solid var(--line);
		border-radius: 12px;
		background: var(--surface);
		color: var(--ink);
		font-size: 0.85rem;
		font-weight: 600;
		cursor: pointer;
	}
	.download {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		color: var(--teal-700);
	}
	.download:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.chart-head {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: 16px;
		margin: 20px 0 12px;
	}
	h2 {
		margin: 0;
		font-size: 1.1rem;
		font-weight: 800;
	}
	.sub {
		margin: 4px 0 0;
		font-size: 0.8rem;
		color: var(--ink-soft);
	}
	.latest {
		text-align: right;
		flex-shrink: 0;
	}
	.latest span {
		display: block;
		font-size: 0.75rem;
		color: var(--ink-soft);
	}
	.latest strong {
		font-size: 1.4rem;
		font-weight: 800;
		color: var(--teal-700);
	}

	.state {
		display: grid;
		place-content: center;
		height: 300px;
		color: var(--ink-soft);
		font-size: 0.9rem;
		text-align: center;
	}
	.state.error {
		color: var(--err);
	}

	.legend {
		display: flex;
		flex-wrap: wrap;
		gap: 18px;
		margin: 14px 0 0;
		padding: 0;
		list-style: none;
		font-size: 0.78rem;
		color: var(--ink-soft);
	}
	.legend li {
		display: flex;
		align-items: center;
		gap: 7px;
	}
	.swatch {
		width: 14px;
		height: 4px;
		border-radius: 2px;
	}
	.swatch.outside {
		background: #f6d98a;
	}
	.swatch.band {
		background: rgba(26, 138, 143, 0.5);
	}
	.swatch.median {
		background: repeating-linear-gradient(90deg, #0f7378 0 4px, transparent 4px 7px);
	}
	.swatch.child {
		background: #143d4a;
	}

	.rail {
		display: flex;
		flex-direction: column;
		gap: 20px;
		min-width: 0;
	}

	.status {
		padding: 22px;
		border-radius: 1.5rem;
		background: #5f8d79;
		color: #fff;
	}
	.status-icon {
		display: grid;
		place-items: center;
		width: 40px;
		height: 40px;
		margin-bottom: 14px;
		border-radius: 50%;
		background: rgb(255 255 255 / 0.18);
	}
	.status h3 {
		margin: 0;
		font-size: 1.05rem;
		font-weight: 800;
		line-height: 1.3;
	}
	.status p {
		margin: 10px 0 14px;
		font-size: 0.82rem;
		line-height: 1.5;
		color: rgb(255 255 255 / 0.85);
	}
	.pill {
		display: inline-flex;
		align-items: center;
		gap: 7px;
		padding: 7px 12px;
		border-radius: 999px;
		background: rgb(255 255 255 / 0.16);
		font-size: 0.78rem;
		font-weight: 700;
	}

	h4 {
		font-size: 0.95rem;
		font-weight: 700;
	}

	.summary {
		list-style: none;
		margin: 14px 0 0;
		padding: 0;
		display: grid;
		gap: 14px;
	}
	.s-label {
		font-size: 0.78rem;
		color: var(--ink-soft);
	}
	.s-row {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		gap: 10px;
	}
	.s-row strong {
		font-size: 1.15rem;
		font-weight: 800;
	}
	.delta {
		font-size: 0.78rem;
		font-weight: 700;
		color: var(--teal-700);
	}
	.delta.down {
		color: var(--err);
	}

	.checkup {
		display: flex;
		gap: 12px;
		align-items: flex-start;
		margin-bottom: 14px;
	}
	.cal {
		display: grid;
		place-items: center;
		flex-shrink: 0;
		width: 36px;
		height: 36px;
		border-radius: 10px;
		background: var(--mint-100);
		color: var(--teal-700);
	}

	@media (max-width: 1100px) {
		.grafik {
			grid-template-columns: 1fr;
		}
	}
	@media (max-width: 640px) {
		.toolbar {
			align-items: stretch;
		}
		.tabs {
			overflow-x: auto;
		}
	}
</style>
