<script lang="ts">
	import Auth from './routes/login/Page.svelte';
	import Main from './routes/main/Page.svelte';
	import { getToken, logout } from './routes/login/auth';
	let token = $state(getToken());
</script>

{#if token}
	<Main
		onLogout={() => {
			logout();
			token = null;
		}}
	/>
{:else}
	<Auth onLogin={(t: string) => (token = t)} />
{/if}
