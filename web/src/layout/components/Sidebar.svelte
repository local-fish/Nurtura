<script lang="ts">
	import type { Section, NavItem } from './interfaces.ts';
	import Icon from './Icon.svelte';

	let {
		active = 'beranda',
		open = false,
		onnavigate = () => {},
		onclose = () => {},
		lastUpdated = 'belum ada aktivitas'
	}: {
		active?: string;
		open?: boolean;
		onnavigate?: (item: NavItem, e: MouseEvent) => void;
		onclose?: () => void;
		lastUpdated?: string;
	} = $props();

	const sections: Section[] = [
		{
			label: 'Menu utama',
			items: [
				{ id: 'beranda', label: 'Beranda', icon: 'home', href: '#/' },
				{ id: 'paspor', label: 'Paspor imunisasi', icon: 'syringe', href: '#/paspor' },
				{ id: 'input', label: 'Input data', icon: 'clipboard', href: '#/input' },
				{ id: 'grafik', label: 'Grafik pertumbuhan', icon: 'chart', href: '#/grafik' }
			]
		}
		/*
    {
      label: 'Bantuan',
      items: [
        { id: 'panduan', label: 'Panduan kesehatan', icon: 'book', href: '#/panduan' },
        { id: 'bantuan', label: 'Pusat bantuan', icon: 'help', href: '#/bantuan' }
      ]
    }
    */
	];

	function handleClick(e: MouseEvent, item: NavItem) {
		onnavigate(item, e);
		onclose();
	}
</script>

<aside class="sidebar" class:open aria-label="Navigasi utama">
	<div class="brand">
		<span class="logo">N</span>
		<div>
			<div class="brand-name">Nurtura</div>
			<div class="brand-tag">Tumbuh bersama</div>
		</div>
		<button class="close" onclick={onclose} aria-label="Tutup menu">
			<Icon name="close" />
		</button>
	</div>

	<nav>
		{#each sections as section}
			<p class="section-label">{section.label}</p>
			<ul>
				{#each section.items as item}
					<li>
						<a
							href={item.href}
							class:active={active === item.id}
							aria-current={active === item.id ? 'page' : undefined}
							onclick={(e) => handleClick(e, item)}
						>
							<Icon name={item.icon} />
							<span>{item.label}</span>
						</a>
					</li>
				{/each}
			</ul>
		{/each}
	</nav>

	<div class="status">
		<span class="status-icon"><Icon name="heart" size={16} /></span>
		<strong>Update Terakhir</strong>
		<span>Data terakhir diperbarui {lastUpdated}.</span>
	</div>
</aside>

{#if open}
	<button class="scrim" onclick={onclose} aria-label="Tutup menu"></button>
{/if}

<style>
	.sidebar {
		width: var(--sidebar-w);
		flex-shrink: 0;
		display: flex;
		flex-direction: column;
		gap: 24px;
		padding: 24px 16px;
		background: var(--surface);
		border-right: 1px solid var(--line);
		overflow-y: auto;
	}

	.brand {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 0 0 4px;
	}
	.logo {
		display: grid;
		place-items: center;
		width: 36px;
		height: 36px;
		border-radius: 10px;
		background: var(--teal-600);
		color: #fff;
		font-weight: 800;
		font-size: 18px;
	}
	.brand-name {
		font-weight: 700;
		font-size: 18px;
		line-height: 1.1;
	}
	.brand-tag {
		font-size: 11px;
		color: var(--ink-soft);
	}

	.close {
		display: none;
	}

	.section-label {
		margin: 0 0 8px;
		padding: 0 14px;
		font-size: 11px;
		color: var(--ink-soft);
	}
	/* Without this the flex column squeezes nav below its content height and the
	   last link overlaps the status card instead of letting the sidebar scroll. */
	nav {
		flex-shrink: 0;
	}
	nav ul {
		list-style: none;
		margin: 0 0 18px;
		padding: 0;
		display: grid;
		gap: 6px;
	}

	a {
		position: relative;
		display: flex;
		align-items: center;
		gap: 12px;
		/* Extra right padding clears the .active::after indicator (3px bar at
		   right: 10px) so long labels don't butt up against it. */
		padding: 13px 24px 13px 14px;
		border-radius: 10px;
		color: var(--ink-soft);
		font-size: 14px;
		font-weight: 500;
		line-height: 1.25;
		text-decoration: none;
		transition:
			background 0.15s,
			color 0.15s;
	}
	a:hover {
		background: var(--mint-50);
		color: var(--ink);
	}
	a.active {
		background: var(--teal-100);
		color: var(--teal-700);
		font-weight: 700;
	}
	a.active::after {
		content: '';
		position: absolute;
		right: 10px;
		top: 50%;
		width: 3px;
		height: 18px;
		transform: translateY(-50%);
		border-radius: 2px;
		background: var(--teal-600);
	}

	.status {
		margin-top: auto;
		flex-shrink: 0;
		display: grid;
		gap: 4px;
		padding: 14px;
		border-radius: 14px;
		background: var(--mint-100);
		font-size: 11px;
		color: var(--ink-soft);
	}
	.status strong {
		font-size: 13px;
		color: var(--ink);
	}
	.status-icon {
		display: grid;
		place-items: center;
		width: 32px;
		height: 32px;
		margin-bottom: 6px;
		border-radius: 50%;
		background: #fff;
		color: var(--teal-600);
	}

	.scrim {
		display: none;
	}

	@media (max-width: 900px) {
		.sidebar {
			position: fixed;
			inset: 0 auto 0 0;
			z-index: 30;
			transform: translateX(-100%);
			transition: transform 0.2s ease;
		}
		.sidebar.open {
			transform: none;
			box-shadow: 0 0 40px rgb(0 0 0 / 0.15);
		}
		.close {
			display: grid;
			place-items: center;
			margin-left: auto;
			width: 32px;
			height: 32px;
			border: 0;
			border-radius: 8px;
			background: transparent;
			cursor: pointer;
		}
		.scrim {
			display: block;
			position: fixed;
			inset: 0;
			z-index: 20;
			border: 0;
			background: rgb(22 33 31 / 0.35);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.sidebar,
		a {
			transition: none;
		}
	}
</style>
