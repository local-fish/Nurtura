<script>
	import Icon from './Icon.svelte';

	let {
		title = '',
		subtitle = '',
		user = { name: 'placeholder', age: 'placeholder', avatar: '' },
		onmenu = () => {}
	} = $props();
</script>

<header class="topbar">
	<button class="menu" onclick={() => onmenu} aria-label="Buka menu">
		<Icon name="menu" size={20} />
	</button>

	<div class="heading">
		<h1>{title}</h1>
		{#if subtitle}<p>{subtitle}</p>{/if}
	</div>

	<button class="icon-btn" aria-label="Notifikasi">
		<Icon name="bell" size={18} />
	</button>

	<button class="profile" aria-label="Menu profil">
		{#if user.avatar}
			<img src={user.avatar} alt="" />
		{:else}
			<span class="avatar">{user.name.charAt(0)}</span>
		{/if}
		<span class="who">
			<strong>{user.name}</strong>
			<small>{user.age}</small>
		</span>
		<Icon name="chevron" size={14} />
	</button>
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

	.profile {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 0;
		border: 0;
		background: transparent;
		color: var(--ink-soft);
		cursor: pointer;
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
</style>
