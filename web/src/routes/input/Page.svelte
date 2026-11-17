<script lang="ts">
	import Icon from '$src/components/Icon.svelte';
	import Button from '$src/components/Button.svelte';
	import { ageMonthsFromDob, formatAge } from '$src/lib/age';
	import {
		createRecord,
		createLog,
		listRecords,
		listVaccines,
		completeVaccine,
		type Patient,
		type MetricRecord,
		type VaccineStatus
	} from '$src/lib/api';

	let {
		patient,
		patients = [],
		onsaved = () => {},
		onselectchild = () => {}
	}: {
		patient: Patient;
		patients?: Patient[];
		onsaved?: () => void;
		onselectchild?: (id: string) => void;
	} = $props();

	type Tab = 'pertumbuhan' | 'imunisasi';
	let tab: Tab = $state('pertumbuhan');

	const today = new Date().toISOString().slice(0, 10);
	const dateFmt = new Intl.DateTimeFormat('id-ID', {
		day: 'numeric',
		month: 'long',
		year: 'numeric'
	});

	const FIELDS = [
		{ metric: 'weight', label: 'Berat badan (BB)', short: 'Berat badan', unit: 'kg' },
		{ metric: 'height', label: 'Tinggi badan (TB)', short: 'Tinggi badan', unit: 'cm' },
		{
			metric: 'head_circumference',
			label: 'Lingkar kepala (LK)',
			short: 'Lingkar kepala',
			unit: 'cm'
		}
	];

	let date = $state(today);
	let values: Record<string, string> = $state({ weight: '', height: '', head_circumference: '' });
	let note = $state('');

	let saving = $state(false);
	let error = $state('');
	let notice = $state('');
	let picking = $state(false);

	let previous: Record<string, MetricRecord> = $state({});
	let vaccines: VaccineStatus[] = $state([]);
	let markingId = $state(0);

	const draftKey = $derived(`nurtura:draft:growth:${patient.id}`);

	function parseNum(raw: string) {
		const n = Number(raw.trim().replace(',', '.'));
		return Number.isFinite(n) && n > 0 ? n : null;
	}

	// Latest record per metric — the API returns newest first, so the first wins.
	async function loadPrevious() {
		try {
			const recs = await listRecords(patient.id);
			const latest: Record<string, MetricRecord> = {};
			for (const r of recs) if (!latest[r.metric]) latest[r.metric] = r;
			previous = latest;
		} catch {
			previous = {};
		}
	}

	async function loadVaccines() {
		try {
			vaccines = (await listVaccines(patient.id)).vaccines;
		} catch {
			vaccines = [];
		}
	}

	// Reload whenever the active child changes, and restore any saved draft.
	$effect(() => {
		const id = patient.id;
		loadPrevious();
		loadVaccines();
		const raw = localStorage.getItem(`nurtura:draft:growth:${id}`);
		if (raw) {
			try {
				const d = JSON.parse(raw);
				date = d.date ?? today;
				values = { weight: '', height: '', head_circumference: '', ...(d.values ?? {}) };
				note = d.note ?? '';
				notice = 'Draf tersimpan dimuat.';
			} catch {
				/* ignore malformed draft */
			}
		} else {
			date = today;
			values = { weight: '', height: '', head_circumference: '' };
			note = '';
		}
	});

	const previousDate = $derived.by(() => {
		const dates = FIELDS.map((f) => previous[f.metric]?.measured_at).filter(Boolean) as string[];
		if (!dates.length) return '';
		return dateFmt.format(new Date(dates.reduce((a, b) => (a > b ? a : b))));
	});

	function delta(metric: string) {
		const prev = previous[metric];
		const now = parseNum(values[metric] ?? '');
		if (!prev || now === null) return null;
		const d = now - prev.value;
		return Math.abs(d) < 0.001 ? null : d;
	}

	function saveDraft() {
		localStorage.setItem(draftKey, JSON.stringify({ date, values, note }));
		error = '';
		notice = 'Draf disimpan di perangkat ini.';
	}

	async function save(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		notice = '';

		if (!date) return (error = 'Tanggal pemeriksaan wajib diisi.');
		if (date > today) return (error = 'Tanggal pemeriksaan tidak boleh di masa depan.');

		const filled = FIELDS.filter((f) => (values[f.metric] ?? '').trim() !== '');
		if (!filled.length && !note.trim()) {
			return (error = 'Isi minimal satu hasil pengukuran atau catatan.');
		}
		for (const f of filled) {
			if (parseNum(values[f.metric]) === null) {
				return (error = `${f.label} harus berupa angka lebih dari 0.`);
			}
		}

		saving = true;
		try {
			for (const f of filled) {
				await createRecord(patient.id, f.metric, parseNum(values[f.metric])!, `${date}T00:00:00Z`);
			}
			if (note.trim()) await createLog(patient.id, note.trim());

			localStorage.removeItem(draftKey);
			values = { weight: '', height: '', head_circumference: '' };
			note = '';
			notice = 'Data pemeriksaan tersimpan.';
			await loadPrevious();
			onsaved();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Gagal menyimpan data pemeriksaan.';
		} finally {
			saving = false;
		}
	}

	async function markVaccine(v: VaccineStatus) {
		markingId = v.vaccine_id;
		error = '';
		try {
			await completeVaccine(patient.id, v.vaccine_id);
			await loadVaccines();
			notice = `${v.name} ditandai selesai.`;
			onsaved();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Gagal menandai imunisasi.';
		} finally {
			markingId = 0;
		}
	}

	const sexLabel = $derived(patient.sex === 'F' ? 'Anak perempuan' : 'Anak laki-laki');
</script>

<div class="page">
	<div class="tabs" role="tablist">
		<button
			role="tab"
			aria-selected={tab === 'pertumbuhan'}
			class:active={tab === 'pertumbuhan'}
			onclick={() => (tab = 'pertumbuhan')}
		>
			Pertumbuhan
		</button>
		<button
			role="tab"
			aria-selected={tab === 'imunisasi'}
			class:active={tab === 'imunisasi'}
			onclick={() => (tab = 'imunisasi')}
		>
			Imunisasi
		</button>
	</div>

	<div class="cols">
		<div class="main-col">
			{#if tab === 'pertumbuhan'}
				<form class="card" onsubmit={save} novalidate>
					<h2>Data pemeriksaan</h2>
					<p class="sub">Isi berdasarkan hasil pemeriksaan terbaru</p>

					{#if error}<p class="alert" role="alert">{error}</p>{/if}
					{#if notice}<p class="ok" role="status">{notice}</p>{/if}

					<label class="lbl" for="exam-date">Tanggal pemeriksaan</label>
					<input id="exam-date" type="date" bind:value={date} max={today} />

					<h3 class="group">Hasil pengukuran fisik</h3>

					<div class="grid-2">
						{#each FIELDS.slice(0, 2) as f (f.metric)}
							<div>
								<label class="lbl" for="m-{f.metric}">{f.label}</label>
								<div class="with-unit">
									<input
										id="m-{f.metric}"
										inputmode="decimal"
										bind:value={values[f.metric]}
										placeholder="0.0"
									/>
									<span class="unit">{f.unit}</span>
								</div>
							</div>
						{/each}
					</div>

					{#each FIELDS.slice(2) as f (f.metric)}
						<label class="lbl" for="m-{f.metric}">{f.label}</label>
						<div class="with-unit">
							<input
								id="m-{f.metric}"
								inputmode="decimal"
								bind:value={values[f.metric]}
								placeholder="0.0"
							/>
							<span class="unit">{f.unit}</span>
						</div>
					{/each}

					<label class="lbl" for="exam-note">Catatan pemeriksaan (Opsional)</label>
					<textarea
						id="exam-note"
						rows="4"
						bind:value={note}
						placeholder="Contoh: Nafsu makan baik, anak aktif dan ceria."></textarea>

					<div class="actions">
						<Button fill={false} style="font-size: 0.85rem" onclick={saveDraft}>
							Simpan sebagai draf
						</Button>
						<Button
							type="submit"
							icon="check"
							loading={saving}
							fill={false}
							style="--btn-bg: var(--teal-700); --btn-color: #fff; font-size: 0.85rem; flex: 1"
						>
							Simpan data pertumbuhan
						</Button>
					</div>
				</form>
			{:else}
				<div class="card">
					<h2>Imunisasi</h2>
					<p class="sub">Tandai imunisasi yang sudah diberikan</p>

					{#if error}<p class="alert" role="alert">{error}</p>{/if}
					{#if notice}<p class="ok" role="status">{notice}</p>{/if}

					{#if !vaccines.length}
						<p class="muted">Belum ada jadwal imunisasi.</p>
					{:else}
						<ul class="vax">
							{#each vaccines as v (v.vaccine_id)}
								<li>
									<span class="vax-text">
										<strong>{v.name} · dosis {v.dose_number}</strong>
										<small>Dianjurkan pada usia {v.recommended_age_months} bulan</small>
									</span>
									{#if v.completed}
										<span class="pill done">Selesai</span>
									{:else}
										<Button
											fill={false}
											loading={markingId === v.vaccine_id}
											style="font-size: 0.8rem; height: 36px"
											onclick={() => markVaccine(v)}
										>
											Tandai selesai
										</Button>
									{/if}
								</li>
							{/each}
						</ul>
					{/if}
				</div>
			{/if}
		</div>

		<aside class="side">
			<div class="card tight">
				<header class="side-head">
					<h3>Profil terpilih</h3>
					{#if patients.length > 1}
						<button class="link" onclick={() => (picking = !picking)}>Ganti</button>
					{/if}
				</header>
				<div class="profile">
					<span class="avatar">{patient.name.charAt(0)}</span>
					<span class="profile-text">
						<strong>{patient.name}</strong>
						<small>{formatAge(ageMonthsFromDob(patient.date_of_birth))} · {sexLabel}</small>
					</span>
					<span class="pill">Profil aktif</span>
				</div>
				{#if picking}
					<ul class="switch">
						{#each patients.filter((p) => p.id !== patient.id) as p (p.id)}
							<li>
								<button
									onclick={() => {
										picking = false;
										onselectchild(p.id);
									}}
								>
									<span class="avatar sm">{p.name.charAt(0)}</span>
									<span>{p.name}</span>
								</button>
							</li>
						{/each}
					</ul>
				{/if}
			</div>

			<div class="card who">
				<span class="who-icon"><Icon name="shield-check" size={20} /></span>
				<h3>Data terintegrasi WHO</h3>
				<p>Setelah disimpan, hasil akan langsung dipetakan ke kurva standar tumbuh kembang WHO.</p>
				<a
					class="who-link"
					href="https://www.who.int/tools/child-growth-standards"
					target="_blank"
					rel="noopener noreferrer"
				>
					Pelajari standar WHO <Icon name="arrow-right" size={14} />
				</a>
			</div>

			<div class="card tight">
				<h3>Pengukuran sebelumnya</h3>
				<p class="sub">{previousDate || 'Belum ada pengukuran'}</p>
				<ul class="prev">
					{#each FIELDS as f (f.metric)}
						{@const prev = previous[f.metric]}
						{@const d = delta(f.metric)}
						<li>
							<span class="prev-label">{f.short}</span>
							<span class="prev-value">{prev ? `${prev.value} ${f.unit}` : '—'}</span>
							{#if d !== null}
								<span class="pill delta" class:down={d < 0}>
									{d > 0 ? '+' : ''}{d.toFixed(1)}
									{f.unit}
								</span>
							{/if}
						</li>
					{/each}
				</ul>
			</div>
		</aside>
	</div>
</div>

<style>
	.page {
		display: flex;
		flex-direction: column;
		gap: 20px;
		padding: 1rem;
	}

	.tabs {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 12px;
		max-width: 600px;
	}
	.tabs button {
		height: 42px;
		border: 1px solid var(--line);
		border-radius: 12px;
		background: var(--surface);
		color: var(--ink-soft);
		font: inherit;
		font-size: 13px;
		font-weight: 700;
		cursor: pointer;
	}
	.tabs button.active {
		border-color: var(--teal-700);
		background: var(--teal-700);
		color: #fff;
	}

	.cols {
		display: grid;
		grid-template-columns: minmax(0, 1.9fr) minmax(0, 1fr);
		gap: 20px;
		align-items: start;
	}

	.card {
		padding: 22px 24px;
		border: 1px solid var(--line);
		border-radius: 16px;
		background: var(--surface);
	}
	.card.tight {
		padding: 18px 20px;
	}

	h2 {
		margin: 0 0 4px;
		font-size: 17px;
		font-weight: 800;
	}
	h3 {
		margin: 0;
		font-size: 14px;
		font-weight: 700;
	}
	.sub {
		margin: 0 0 16px;
		color: var(--ink-soft);
		font-size: 12.5px;
	}
	.muted {
		margin: 0;
		color: var(--ink-soft);
		font-size: 13px;
	}

	.lbl {
		display: block;
		margin: 0 0 7px;
		font-size: 12.5px;
		font-weight: 600;
	}

	.group {
		margin: 22px 0 14px;
		padding-left: 10px;
		border-left: 3px solid var(--teal-600);
		font-size: 13px;
	}

	input[type='date'],
	input:not([type]),
	textarea {
		width: 100%;
		padding: 0 14px;
		border: 1px solid var(--line);
		border-radius: 11px;
		background: var(--surface);
		font: inherit;
		font-size: 14px;
		color: var(--ink);
	}
	input[type='date'],
	input:not([type]) {
		height: 46px;
	}
	textarea {
		padding: 12px 14px;
		resize: vertical;
		line-height: 1.6;
	}
	input::placeholder,
	textarea::placeholder {
		color: #9bada9;
	}
	input:focus-visible,
	textarea:focus-visible {
		outline: none;
		border-color: var(--teal-600);
		box-shadow: 0 0 0 3px var(--teal-100);
	}

	.grid-2 {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 16px;
	}
	.lbl:not(:first-of-type) {
		margin-top: 16px;
	}

	.with-unit {
		position: relative;
	}
	.with-unit input {
		padding-right: 52px;
	}
	.unit {
		position: absolute;
		top: 50%;
		right: 10px;
		transform: translateY(-50%);
		padding: 3px 9px;
		border-radius: 7px;
		background: var(--mint-100);
		color: var(--teal-700);
		font-size: 11px;
		font-weight: 700;
	}

	.actions {
		display: flex;
		gap: 12px;
		margin-top: 20px;
	}

	.alert,
	.ok {
		margin: 0 0 14px;
		padding: 10px 13px;
		border-radius: 10px;
		font-size: 12.5px;
	}
	.alert {
		background: #fbeaea;
		color: var(--err);
	}
	.ok {
		background: var(--mint-100);
		color: var(--teal-700);
	}

	.side {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.side-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 10px;
		margin-bottom: 14px;
	}
	.link {
		border: 0;
		background: none;
		color: var(--teal-700);
		font: inherit;
		font-size: 12px;
		font-weight: 700;
		cursor: pointer;
	}
	.link:hover {
		text-decoration: underline;
	}

	.profile {
		display: flex;
		align-items: center;
		gap: 11px;
	}
	.avatar {
		display: grid;
		place-items: center;
		width: 42px;
		height: 42px;
		flex-shrink: 0;
		border-radius: 50%;
		background: var(--teal-100);
		color: var(--teal-700);
		font-weight: 700;
	}
	.avatar.sm {
		width: 26px;
		height: 26px;
		font-size: 12px;
	}
	.profile-text {
		display: grid;
		margin-right: auto;
		min-width: 0;
		line-height: 1.3;
	}
	.profile-text strong {
		font-size: 14px;
	}
	.profile-text small {
		color: var(--ink-soft);
		font-size: 11.5px;
	}

	.pill {
		flex-shrink: 0;
		padding: 4px 10px;
		border-radius: 8px;
		background: var(--mint-100);
		color: var(--teal-700);
		font-size: 11px;
		font-weight: 700;
		white-space: nowrap;
	}
	.pill.done {
		background: #e4f3e8;
		color: #2c6b43;
	}
	.pill.delta.down {
		background: #fdf0e6;
		color: #9a5a1b;
	}

	.switch {
		list-style: none;
		margin: 14px 0 0;
		padding: 12px 0 0;
		border-top: 1px solid var(--line);
		display: grid;
		gap: 6px;
	}
	.switch button {
		display: flex;
		align-items: center;
		gap: 10px;
		width: 100%;
		padding: 7px 8px;
		border: 0;
		border-radius: 9px;
		background: none;
		color: var(--ink);
		font: inherit;
		font-size: 13px;
		text-align: left;
		cursor: pointer;
	}
	.switch button:hover {
		background: var(--mint-50);
	}

	.who {
		background: var(--mint-100);
		border-color: var(--teal-100);
	}
	.who-icon {
		display: grid;
		place-items: center;
		width: 40px;
		height: 40px;
		margin-bottom: 14px;
		border-radius: 12px;
		background: var(--surface);
		color: var(--teal-600);
	}
	.who p {
		margin: 8px 0 14px;
		color: var(--ink-soft);
		font-size: 12px;
		line-height: 1.6;
	}
	.who-link {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		color: var(--teal-700);
		font-size: 12px;
		font-weight: 700;
		text-decoration: none;
	}
	.who-link:hover {
		text-decoration: underline;
	}

	.prev {
		list-style: none;
		margin: 16px 0 0;
		padding: 0;
		display: grid;
		gap: 12px;
	}
	.prev li {
		display: flex;
		align-items: center;
		gap: 10px;
		font-size: 13px;
	}
	.prev-label {
		margin-right: auto;
		color: var(--ink-soft);
	}
	.prev-value {
		font-weight: 700;
	}

	.vax {
		list-style: none;
		margin: 0;
		padding: 0;
		display: grid;
		gap: 10px;
	}
	.vax li {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px 14px;
		border: 1px solid var(--line);
		border-radius: 12px;
	}
	.vax-text {
		display: grid;
		margin-right: auto;
		min-width: 0;
		line-height: 1.35;
	}
	.vax-text strong {
		font-size: 13px;
	}
	.vax-text small {
		color: var(--ink-soft);
		font-size: 11.5px;
	}

	@media (max-width: 1000px) {
		.cols {
			grid-template-columns: minmax(0, 1fr);
		}
	}
	@media (max-width: 560px) {
		.grid-2 {
			grid-template-columns: minmax(0, 1fr);
		}
		.actions {
			flex-direction: column-reverse;
		}
	}
</style>
