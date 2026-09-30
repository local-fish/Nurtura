<script>
  import Sidebar from "./components/Sidebar.svelte";
  import Topbar from "./components/Topbar.svelte";

  let {
    active = 'beranda',
    title = '',
    subtitle = '',
    user,
    onnavigate,
    children
  } = $props();

  let menuOpen = $state(false);
</script>

<div class="shell">
  <Sidebar {active} open={menuOpen} {onnavigate} onclose={() => (menuOpen = false)} />

  <div class="main">
    <Topbar {title} {subtitle} {user} onmenu={() => (menuOpen = true)} />
    <main class="content">
      {@render children?.()}
    </main>
  </div>
</div>


<style>
  .shell { display: flex; height: 100%; }
  .main { flex: 1; min-width: 0; display: flex; flex-direction: column; }
  .content {
    flex: 1;
    overflow-y: auto;
    padding: 28px 32px;
    background: var(--mint-50);
  }
  @media (max-width: 900px) {
    .content { padding: 20px 16px; }
  }
</style>
