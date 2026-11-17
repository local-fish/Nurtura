<script lang="ts">
	import Icon from './Icon.svelte';
	import { ageMonthsFromDob, formatAge } from '$src/lib/age';
	import type { Patient } from '$src/lib/api';

	let {
		title = '',
		subtitle = '',
		user = { name: 'placeholder', age: 'placeholder', avatar: '' },
		patients = [],
		activeId = '',
		onmenu = () => {},
		onselectchild = () => {},
		onaddchild = () => {},
		onlogout = () => {}
	}: {
		title?: string;
		subtitle?: string;
		user?: { name: string; age: string; avatar: string };
		patients?: Patient[];
		activeId?: string;
		onmenu?: () => void;
		onselectchild?: (id: string) => void;
		onaddchild?: () => void;
		onlogout?: () => void;
	} = $props();

	let open = $state(false);
	let root: HTMLElement | undefined = $state();

	// Close on outside click or Escape while the menu is open.
	$effect(() => {
		if (!open) return;
		const onClick = (e: MouseEvent) => {
			if (root && !root.contains(e.target as Node)) open = false;
		};
		const onKey = (e: KeyboardEvent) => {
			if (e.key === 'Escape') open = false;
		};
		document.addEventListener('click', onClick);
		document.addEventListener('keydown', onKey);
		return () => {
			document.removeEventListener('click', onClick);
			document.removeEventListener('keydown', onKey);
		};
	});

	function pick(id: string) {
		open = false;
		if (id !== activeId) onselectchild(id);
	}
</script>

<header class="topbar">
	<button class="menu" onclick={onmenu} aria-label="Buka menu">
		<Icon name="menu" size={20} />
	</button>

	<div class="heading">
		<h1>{title}</h1>
		{#if subtitle}<p>{subtitle}</p>{/if}
	</div>

	<button class="icon-btn" aria-label="Notifikasi">
		<Icon name="bell" size={18} />
	</button>

	<div class="profile-wrap" bind:this={root}>
		<button
			class="profile"
			onclick={() => (open = !open)}
			aria-haspopup="menu"
			aria-expanded={open}
			aria-label="Menu profil"
		>
			{#if user.avatar}
				<img src={user.avatar} alt="" />
			{:else}
				<span class="avatar">{user.name.charAt(0)}</span>
			{/if}
			<span class="who">
				<strong>{user.name}</strong>
				<small>{user.age}</small>
			</span>
			<span class="chev" class:open><Icon name="chevron" size={14} /></span>
		</button>

		{#if open}
			<div class="panel" role="menu">
				{#if patients.length}
					<p class="panel-label">Anak</p>
					{#each patients as p (p.id)}
						<button
							class="item"
							role="menuitemradio"
							aria-checked={p.id === activeId}
							onclick={() => pick(p.id)}
						>
							<span class="mini">{p.name.charAt(0)}</span>
							<span class="item-text">
								<strong>{p.name}</strong>
								<small>{formatAge(ageMonthsFromDob(p.date_of_birth))}</small>
							</span>
							{#if p.id === activeId}
								<span class="tick"><Icon name="check" size={15} /></span>
							{/if}
						</button>
					{/each}
					<hr />
				{/if}

				<button
					class="item"
					role="menuitem"
					onclick={() => {
						open = false;
						onaddchild();
					}}
				>
					<span class="slot"><Icon name="plus" size={16} /></span>
					<span>Tambah anak</span>
				</button>

				<hr />

				<button
					class="item danger"
					role="menuitem"
					onclick={() => {
						open = false;
						onlogout();
					}}
				>
					<span class="slot"><Icon name="log-out" size={16} /></span>
					<span>Keluar</span>
				</button>
			</div>
		{/if}
	</div>
</header>

<style>
	.topbar {
		height: var(--topbar-h);
		flex-shrink: 0;
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 0 32px;
		background: var(--surface);
		border-bottom: 1px solid var(--line);
	}

	.heading {
		margin-right: auto;
		min-width: 0;
	}
	h1 {
		margin: 0;
		font-size: 20px;
		font-weight: 700;
		line-height: 1.2;
	}
	.heading p {
		margin: 2px 0 0;
		font-size: 12px;
		color: var(--ink-soft);
	}

	.icon-btn,
	.menu {
		display: grid;
		place-items: center;
		width: 38px;
		height: 38px;
		border: 1px solid var(--line);
		border-radius: 50%;
		background: var(--mint-50);
		color: var(--ink);
		cursor: pointer;
	}
	.menu {
		display: none;
		border-radius: 10px;
	}

	.profile-wrap {
		position: relative;
	}
	.profile {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 4px;
		border: 0;
		border-radius: 10px;
		background: transparent;
		color: var(--ink-soft);
		cursor: pointer;
	}
	.profile:hover {
		background: var(--mint-50);
	}
	.avatar,
	img {
		width: 34px;
		height: 34px;
		border-radius: 50%;
		object-fit: cover;
	}
	.avatar {
		display: grid;
		place-items: center;
		background: var(--teal-100);
		color: var(--teal-700);
		font-weight: 700;
	}
	.who {
		display: grid;
		text-align: left;
		line-height: 1.2;
	}
	.who strong {
		font-size: 12px;
		color: var(--ink);
	}
	.who small {
		font-size: 11px;
	}
	.chev {
		display: grid;
		place-items: center;
		transition: transform 0.15s ease;
	}
	.chev.open {
		transform: rotate(180deg);
	}

	.panel {
		position: absolute;
		top: calc(100% + 8px);
		right: 0;
		z-index: 40;
		min-width: 248px;
		padding: 6px;
		border: 1px solid var(--line);
		border-radius: 14px;
		background: var(--surface);
		box-shadow: 0 12px 32px rgb(22 33 31 / 0.12);
	}
	.panel-label {
		margin: 6px 10px 6px;
		font-size: 11px;
		color: var(--ink-soft);
	}

	.item {
		display: flex;
		align-items: center;
		gap: 10px;
		width: 100%;
		padding: 9px 10px;
		border: 0;
		border-radius: 10px;
		background: transparent;
		color: var(--ink);
		font: inherit;
		font-size: 13px;
		text-align: left;
		cursor: pointer;
	}
	.item:hover {
		background: var(--mint-50);
	}
	.item.danger {
		color: var(--err);
	}
	.item.danger:hover {
		background: #fbeaea;
	}

	.mini {
		display: grid;
		place-items: center;
		width: 28px;
		height: 28px;
		flex-shrink: 0;
		border-radius: 50%;
		background: var(--teal-100);
		color: var(--teal-700);
		font-size: 12px;
		font-weight: 700;
	}
	.item-text {
		display: grid;
		margin-right: auto;
		line-height: 1.25;
		min-width: 0;
	}
	.item-text strong {
		font-size: 13px;
		font-weight: 600;
	}
	.item-text small {
		font-size: 11px;
		color: var(--ink-soft);
	}
	.tick {
		color: var(--teal-600);
	}
	.slot {
		display: grid;
		place-items: center;
		width: 28px;
		flex-shrink: 0;
	}

	hr {
		margin: 6px 8px;
		border: 0;
		border-top: 1px solid var(--line);
	}

	@media (max-width: 900px) {
		.topbar {
			padding: 0 16px;
		}
		.menu {
			display: grid;
		}
		.who {
			display: none;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.chev {
			transition: none;
		}
	}
</style>
