<script>
	let { name, size = 18, stroke = 1.6 } = $props();

	const files = import.meta.glob('$src/assets/icons/*.svg', {
		query: '?raw',
		import: 'default',
		eager: true
	});

	// '/src/assets/icons/plus.svg' -> 'plus', and strip the outer <svg> wrapper
	const icons = Object.fromEntries(
		Object.entries(files).map(([path, raw]) => {
			const key = path.split('/').pop()?.replace('.svg', '');
			const inner = raw.replace(/^[\s\S]*?<svg[^>]*>/i, '').replace(/<\/svg>\s*$/i, '');
			return [key, inner];
		})
	);

	const markup = $derived(icons[name]);
</script>

{#if markup}
	<svg
		width={size}
		height={size}
		fill="none"
		stroke="currentColor"
		stroke-width={stroke}
		stroke-linecap="round"
		stroke-linejoin="round"
		aria-hidden="true"
		style="display: block"
	>
		{@html markup}
	</svg>
{:else if import.meta.env.DEV}
	{console.warn(`Icon "${name}" not found in assets/icons`)}
{/if}
