<script lang="ts">
	import Icon from '$src/components/Icon.svelte';
	import Button from '$src/components/Button.svelte';
	import { createPatient, createRecord } from '$src/lib/api';

	let {
		mode = 'welcome',
		onCreated = () => {},
		onCancel
	}: {
		mode?: 'welcome' | 'add';
		onCreated?: (patientId: string) => void;
		onCancel?: () => void;
	} = $props();

	const copy = $derived(
		mode === 'add'
			? {
					eyebrow: 'Tambah anak',
					heading: 'Tambahkan profil anak berikutnya.',
					lead: 'Setiap anak memiliki catatan pertumbuhan dan jadwal imunisasinya sendiri.'
				}
			: {
					eyebrow: 'Selamat datang di Nurtura',
					heading: 'Kenali si kecil, mulai perjalanannya.',
					lead: 'Lengkapi profil anak agar catatan pertumbuhan dan imunisasi tersimpan pada profil yang tepat.'
				}
	);

	let name = $state('');
	let dob = $state('');
	let sex: 'M' | 'F' | '' = $state('');

	let weight = $state('');
	let length = $state('');
	let headCirc = $state('');

	let saving = $state(false);
	let error = $state('');
	let skipped: string[] = $state([]);
	let createdId = $state('');

	const today = new Date().toISOString().slice(0, 10);

	// Accepts "3,2" as well as "3.2" — Indonesian keyboards produce the comma.
	function parseMeasure(raw: string) {
		const n = Number(raw.trim().replace(',', '.'));
		return Number.isFinite(n) && n > 0 ? n : null;
	}

	const measures = $derived([
		{ metric: 'weight', label: 'Berat saat lahir', raw: weight },
		{ metric: 'height', label: 'Panjang saat lahir', raw: length },
		{ metric: 'head_circumference', label: 'Lingkar kepala saat lahir', raw: headCirc }
	]);

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		skipped = [];

		if (!name.trim()) return (error = 'Nama anak wajib diisi.');
		if (!dob) return (error = 'Tanggal lahir wajib diisi.');
		if (dob > today) return (error = 'Tanggal lahir tidak boleh di masa depan.');
		if (!sex) return (error = 'Pilih jenis kelamin terlebih dahulu.');

		const filled = measures.filter((m) => m.raw.trim() !== '');
		for (const m of filled) {
			if (parseMeasure(m.raw) === null) {
				return (error = `${m.label} harus berupa angka lebih dari 0.`);
			}
		}

		saving = true;
		try {
			const patient = await createPatient(name.trim(), dob, sex);

			// Birth measurements are recorded as of the date of birth, not today.
			const failed: string[] = [];
			for (const m of filled) {
				try {
					await createRecord(patient.id, m.metric, parseMeasure(m.raw)!, `${dob}T00:00:00Z`);
				} catch {
					failed.push(m.label);
				}
			}

			createdId = patient.id;
			if (failed.length) {
				skipped = failed;
				return;
			}
			onCreated(patient.id);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Gagal menyimpan profil anak.';
		} finally {
			saving = false;
		}
	}
</script>

<div class="page">
	<header class="topbar">
		<div class="brand">
			<span class="logo">N</span>
			<div>
				<div class="brand-name">Nurtura</div>
				<div class="brand-tag">Tumbuh bersama</div>
			</div>
		</div>
		<a class="help" href="#/bantuan">Butuh bantuan?</a>
	</header>

	<div class="body">
		<aside class="intro">
			<p class="eyebrow">{copy.eyebrow}</p>
			<h1>{copy.heading}</h1>
			<p class="lead">{copy.lead}</p>

			<div class="note-card">
				<span class="note-icon"><Icon name="book" size={26} /></span>
				<h2>Siapkan buku KIA atau catatan kelahiran</h2>
				<p>
					Masukkan hasil pengukuran saat lahir sesuai catatan. Jika belum tersedia, Anda dapat
					melengkapinya nanti.
				</p>
			</div>

			<p class="privacy">
				<Icon name="shield-check" size={16} />
				<span>
					Data ini hanya digunakan untuk membantu pengelolaan catatan kesehatan anak Anda.
				</span>
			</p>
		</aside>

		<form class="panel" onsubmit={submit} novalidate>
			<div class="section-head">
				<h2>Profil anak</h2>
				<p>Informasi dasar dan pengukuran saat lahir</p>
			</div>

			{#if error}
				<p class="alert" role="alert">{error}</p>
			{/if}

			<div class="grid two">
				<div class="field">
					<label for="child-name">Nama anak <span class="req">*</span></label>
					<input
						id="child-name"
						bind:value={name}
						placeholder="Contoh: Aura"
						autocomplete="off"
						required
					/>
				</div>
				<div class="field">
					<label for="child-dob">Tanggal lahir <span class="req">*</span></label>
					<input id="child-dob" type="date" bind:value={dob} max={today} required />
				</div>
			</div>

			<fieldset class="field sex">
				<legend>Jenis kelamin <span class="req">*</span></legend>
				<div class="grid two">
					{#each [{ value: 'F', label: 'Perempuan' }, { value: 'M', label: 'Laki-laki' }] as opt (opt.value)}
						<label class="choice" class:selected={sex === opt.value}>
							<input type="radio" name="sex" value={opt.value} bind:group={sex} />
							<span class="dot" aria-hidden="true"></span>
							<span>{opt.label}</span>
						</label>
					{/each}
				</div>
				<p class="hint">
					Kurva WHO berbeda untuk anak perempuan dan laki-laki. Pilihan ini digunakan untuk
					menampilkan kurva pertumbuhan yang sesuai.
				</p>
			</fieldset>

			<hr />

			<div class="section-head">
				<h2>Pengukuran saat lahir</h2>
				<p>Opsional · Isi sesuai catatan kelahiran, bukan hasil pengukuran terbaru.</p>
			</div>

			<div class="grid three">
				<div class="field">
					<label for="birth-weight">Berat saat lahir</label>
					<div class="with-unit">
						<input
							id="birth-weight"
							bind:value={weight}
							inputmode="decimal"
							placeholder="Contoh: 3.2"
						/>
						<span class="unit">kg</span>
					</div>
				</div>
				<div class="field">
					<label for="birth-length">Panjang saat lahir</label>
					<div class="with-unit">
						<input
							id="birth-length"
							bind:value={length}
							inputmode="decimal"
							placeholder="Contoh: 49.0"
						/>
						<span class="unit">cm</span>
					</div>
				</div>
				<div class="field">
					<label for="birth-head">Lingkar kepala saat lahir</label>
					<div class="with-unit">
						<input
							id="birth-head"
							bind:value={headCirc}
							inputmode="decimal"
							placeholder="Contoh: 34.0"
						/>
						<span class="unit">cm</span>
					</div>
				</div>
			</div>

			{#if skipped.length}
				<div class="banner warn" role="alert">
					<Icon name="info" size={16} />
					<span>
						Profil anak berhasil dibuat, tetapi pengukuran berikut gagal disimpan: {skipped.join(
							', '
						)}. Anda dapat menambahkannya lagi dari halaman input data.
					</span>
				</div>
			{:else}
				<div class="banner">
					<Icon name="info" size={16} />
					<span>
						Belum memiliki data kelahiran? Kosongkan isian pengukuran. Profil tetap dapat dibuat
						tanpa menambahkan angka perkiraan.
					</span>
				</div>
			{/if}

			<div class="actions">
				<span class="required-note">* Wajib diisi</span>
				<div class="action-buttons">
					{#if skipped.length}
						<Button
							fill={false}
							style="--btn-bg: var(--teal-700); --btn-color: #fff; font-size: 0.9rem"
							onclick={() => onCreated(createdId)}
						>
							Lanjut ke beranda
						</Button>
					{:else}
						{#if onCancel}
							<Button fill={false} style="font-size: 0.9rem" onclick={onCancel}>Batal</Button>
						{/if}
						<Button
							type="submit"
							icon="check"
							loading={saving}
							fill={false}
							style="--btn-bg: var(--teal-700); --btn-color: #fff; font-size: 0.9rem"
						>
							{mode === 'add' ? 'Simpan profil anak' : 'Simpan profil & mulai'}
						</Button>
					{/if}
				</div>
			</div>
		</form>
	</div>
</div>

<style>
	.page {
		min-height: 100%;
		display: flex;
		flex-direction: column;
		background: var(--mint-50);
	}

	.topbar {
		display: flex;
		align-items: center;
		gap: 16px;
		padding: 14px 40px;
		background: var(--surface);
		border-bottom: 1px solid var(--line);
	}
	.brand {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-right: auto;
	}
	.logo {
		display: grid;
		place-items: center;
		width: 38px;
		height: 38px;
		border-radius: 11px;
		background: var(--teal-700);
		color: #fff;
		font-weight: 800;
		font-size: 19px;
	}
	.brand-name {
		font-size: 19px;
		font-weight: 700;
		line-height: 1.15;
	}
	.brand-tag {
		font-size: 11px;
		color: var(--ink-soft);
	}
	.help {
		color: var(--teal-700);
		font-size: 13px;
		font-weight: 600;
		text-decoration: none;
	}
	.help:hover {
		text-decoration: underline;
	}

	.body {
		flex: 1;
		display: grid;
		grid-template-columns: minmax(240px, 300px) minmax(0, 1fr);
		gap: 40px;
		width: 100%;
		max-width: 1100px;
		margin: 0 auto;
		padding: 40px 40px 56px;
	}

	.eyebrow {
		margin: 0 0 14px;
		color: var(--teal-700);
		font-size: 11px;
		font-weight: 700;
		letter-spacing: 0.09em;
		text-transform: uppercase;
	}
	.intro h1 {
		margin: 0 0 16px;
		font-size: 31px;
		font-weight: 800;
		line-height: 1.28;
		letter-spacing: -0.015em;
	}
	.lead {
		margin: 0;
		color: var(--ink-soft);
		font-size: 14px;
		line-height: 1.65;
	}

	.note-card {
		margin: 28px 0 20px;
		padding: 22px;
		border: 1px solid #d5e7dd;
		border-radius: 16px;
		background: #eaf4ee;
	}
	.note-icon {
		display: block;
		margin-bottom: 16px;
		color: var(--teal-700);
	}
	.note-card h2 {
		margin: 0 0 10px;
		font-size: 15px;
		font-weight: 700;
		line-height: 1.4;
	}
	.note-card p {
		margin: 0;
		color: var(--ink-soft);
		font-size: 12.5px;
		line-height: 1.65;
	}

	.privacy {
		display: flex;
		gap: 9px;
		margin: 0;
		color: var(--ink-soft);
		font-size: 12px;
		line-height: 1.55;
	}
	.privacy :global(svg) {
		flex-shrink: 0;
		margin-top: 1px;
		color: var(--teal-600);
	}

	.panel {
		padding: 28px 32px 26px;
		border: 1px solid var(--line);
		border-radius: 18px;
		background: var(--surface);
		box-shadow: 0 1px 3px rgb(22 33 31 / 0.04);
	}

	.section-head h2 {
		margin: 0 0 5px;
		font-size: 17px;
		font-weight: 700;
	}
	.section-head p {
		margin: 0;
		color: var(--ink-soft);
		font-size: 12.5px;
	}

	.grid {
		display: grid;
		gap: 18px;
	}
	.two {
		grid-template-columns: repeat(2, minmax(0, 1fr));
	}
	.three {
		grid-template-columns: repeat(3, minmax(0, 1fr));
	}
	.section-head + .grid,
	.section-head + .alert {
		margin-top: 20px;
	}

	.field {
		min-width: 0;
	}
	/* Direct child only — the gender options are also <label>s inside .field.sex
	   and must not inherit the field-label styling. */
	.field > label,
	.sex legend {
		display: block;
		margin-bottom: 7px;
		padding: 0;
		font-size: 12.5px;
		font-weight: 600;
	}
	.req {
		color: var(--ink-soft);
		font-weight: 500;
	}

	input[type='date'],
	input:not([type]) {
		width: 100%;
		height: 46px;
		padding: 0 14px;
		border: 1px solid var(--line);
		border-radius: 11px;
		background: var(--surface);
		font: inherit;
		font-size: 14px;
		color: var(--ink);
	}
	input::placeholder {
		color: #9bada9;
	}
	input:focus-visible {
		outline: none;
		border-color: var(--teal-600);
		box-shadow: 0 0 0 3px var(--teal-100);
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

	.sex {
		margin: 22px 0 0;
		padding: 0;
		border: 0;
	}

	.choice {
		position: relative;
		display: flex;
		align-items: center;
		gap: 11px;
		height: 50px;
		padding: 0 16px;
		border: 1px solid var(--line);
		border-radius: 11px;
		background: var(--surface);
		font-size: 14px;
		cursor: pointer;
		transition:
			border-color 0.15s,
			background 0.15s;
	}
	.choice:hover {
		border-color: var(--teal-100);
		background: var(--mint-50);
	}
	.choice.selected {
		border-color: var(--teal-600);
		background: var(--mint-50);
		color: var(--teal-700);
		font-weight: 600;
	}
	.choice input {
		position: absolute;
		opacity: 0;
		pointer-events: none;
	}
	.choice:focus-within {
		box-shadow: 0 0 0 3px var(--teal-100);
	}
	.dot {
		display: grid;
		place-items: center;
		width: 17px;
		height: 17px;
		flex-shrink: 0;
		border: 1.5px solid #c3d3cf;
		border-radius: 50%;
	}
	.choice.selected .dot {
		border-color: var(--teal-600);
	}
	.choice.selected .dot::after {
		content: '';
		width: 7px;
		height: 7px;
		border-radius: 50%;
		background: var(--teal-600);
	}

	.hint {
		margin: 11px 0 0;
		color: var(--ink-soft);
		font-size: 11.5px;
		line-height: 1.55;
	}

	hr {
		margin: 24px 0;
		border: 0;
		border-top: 1px solid var(--line);
	}

	.banner {
		display: flex;
		gap: 10px;
		margin-top: 20px;
		padding: 13px 15px;
		border-radius: 11px;
		background: var(--mint-100);
		color: var(--ink-soft);
		font-size: 12px;
		line-height: 1.6;
	}
	.banner :global(svg) {
		flex-shrink: 0;
		margin-top: 2px;
		color: var(--teal-600);
	}
	.banner.warn {
		background: #fdf3e6;
		color: #8a5a1b;
	}
	.banner.warn :global(svg) {
		color: #c2820f;
	}

	.alert {
		margin: 0 0 4px;
		padding: 11px 14px;
		border-radius: 11px;
		background: #fbeaea;
		color: var(--err);
		font-size: 12.5px;
	}

	.actions {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		margin-top: 24px;
	}
	.required-note {
		color: var(--ink-soft);
		font-size: 12px;
	}
	.action-buttons {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	@media (max-width: 900px) {
		.topbar {
			padding: 14px 20px;
		}
		.body {
			grid-template-columns: minmax(0, 1fr);
			gap: 28px;
			padding: 28px 20px 40px;
		}
		.intro h1 {
			font-size: 26px;
		}
		.panel {
			padding: 22px 20px;
		}
	}

	@media (max-width: 640px) {
		.two,
		.three {
			grid-template-columns: minmax(0, 1fr);
		}
		.actions {
			flex-direction: column-reverse;
			align-items: stretch;
			text-align: center;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.choice {
			transition: none;
		}
	}
</style>
