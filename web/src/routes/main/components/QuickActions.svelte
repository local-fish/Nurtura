<script lang="ts">
	import Icon from '$src/components/Icon.svelte';

	let { name = '', onselect = () => {} }: { name?: string; onselect?: (id: string) => void } =
		$props();

	const tiles = [
		{ id: 'input', label: 'Input data baru', icon: 'plus' },
		{ id: 'paspor', label: 'Paspor imunisasi', icon: 'syringe' },
		{ id: 'grafik', label: 'Grafik pertumbuhan', icon: 'chart' },
		// No report endpoint exists yet, so this stays inert rather than pretending.
		{ id: 'laporan', label: 'Unduh laporan', icon: 'download', disabled: true }
	];
</script>

<section class="quick">
	<header>
		<h3>Akses cepat</h3>
		<p>Kelola data {name}</p>
	</header>

	<div class="tiles">
		{#each tiles as t (t.id)}
			<button
				class="tile"
				disabled={t.disabled}
				title={t.disabled ? 'Belum tersedia' : undefined}
				onclick={() => onselect(t.id)}
			>
				<Icon name={t.icon} size={20} />
				<span>{t.label}</span>
			</button>
		{/each}
	</div>
</section>

<style>
	h3 {
		margin: 0;
		font-size: 17px;
		font-weight: 800;
		letter-spacing: -0.01em;
	}
	header p {
		margin: 3px 0 0;
		color: var(--ink-soft);
		font-size: 12.5px;
	}

	.tiles {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 14px;
		margin-top: 18px;
	}

	.tile {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		justify-content: space-between;
		gap: 20px;
		min-height: 92px;
		padding: 16px;
		border: 1px solid var(--teal-100);
		border-radius: 14px;
		background: var(--mint-50);
		color: var(--ink);
		font: inherit;
		text-align: left;
		cursor: pointer;
		transition:
			border-color 0.15s,
			background 0.15s,
			transform 0.08s;
	}
	.tile :global(svg) {
		color: var(--teal-600);
	}
	.tile span {
		font-size: 13px;
		font-weight: 700;
		line-height: 1.3;
	}
	.tile:hover:not(:disabled) {
		border-color: var(--teal-600);
		background: var(--mint-100);
	}
	.tile:active:not(:disabled) {
		transform: scale(0.99);
	}
	.tile:disabled {
		opacity: 0.55;
		cursor: not-allowed;
	}

	@media (max-width: 520px) {
		.tiles {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
