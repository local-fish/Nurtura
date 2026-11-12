<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		bg = '#123456',
		border,
		children
	}: { bg?: string; border?: string; children?: Snippet } = $props();
</script>

<div class="card" style:--bg={bg} style:border>
	{@render children?.()}
</div>

<style>
	.card {
		border-radius: 2rem;
		background-color: var(--bg);
		height: 100%;
		padding: 1rem 2rem;
		flex: 0.5;
	}
</style>
