<script lang="ts">
let { bg = "#123456", children } = $props();
</script>

<div class="card" style:--bg={bg}>
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
