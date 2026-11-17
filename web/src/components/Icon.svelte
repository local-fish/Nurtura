<script>
	let { name, size = 18, stroke = 1.6 } = $props();

	const files = import.meta.glob('$src/assets/icons/*.svg', {
		query: '?raw',
		import: 'default',
		eager: true
	});

	// '/src/assets/icons/plus.svg' -> 'plus', and strip the outer <svg> wrapper.
	// The source viewBox is kept, otherwise the paths render at raw user units
	// inside a {size}px box and get clipped.
	const icons = Object.fromEntries(
		Object.entries(files).map(([path, raw]) => {
			const key = path.split('/').pop()?.replace('.svg', '');
			const viewBox = raw.match(/viewBox="([^"]+)"/i)?.[1] ?? '0 0 24 24';
			const inner = raw.replace(/^[\s\S]*?<svg[^>]*>/i, '').replace(/<\/svg>\s*$/i, '');
			return [key, { inner, viewBox }];
		})
	);

	const icon = $derived(icons[name]);
</script>

{#if icon}
	<svg
		width={size}
		height={size}
		viewBox={icon.viewBox}
		fill="none"
		stroke="currentColor"
		stroke-width={stroke}
		stroke-linecap="round"
		stroke-linejoin="round"
		aria-hidden="true"
		style="display: block"
	>
		{@html icon.inner}
	</svg>
{:else if import.meta.env.DEV}
	{console.warn(`Icon "${name}" not found in assets/icons`)}
{/if}
