<script lang="ts">
import Icon from "./Icon.svelte";

let {
  children,
  icon = null,
  type = 'button',
  loading = false,
  disabled = false,
  fill = true,
  ...rest
} = $props();
</script>

<button
  {type}
  class="btn"
  class:fill
  disabled={disabled || loading}
  aria-busy={loading}
  {...rest}
>
  {#if loading}
    <span class="spinner" aria-hidden="true"></span>
  {:else if icon}
    <Icon name={icon}/>
  {/if}
  {@render children?.()}
</button>

<style>
.btn {
  --btn-color: var(--teal-700);
  --btn-bg: var(--surface);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.6rem;
  height: 44px;
  padding: 0 1.25rem;
  border: none;
  border-radius: 16px;
  background: var(--btn-bg);
  color: var(--btn-color);
  font: inherit;
  font-size: 1.125rem;
  font-weight: 700;
  cursor: pointer;
  transition: transform 0.08s ease, filter 0.15s ease;
}
.fill { width: 100%; }
.btn:hover:not(:disabled) { filter: brightness(0.97); }
.btn:active:not(:disabled) { transform: scale(0.98); }
.btn:focus-visible { outline: 2px solid var(--btn-color); outline-offset: 2px; }
.btn:disabled { opacity: 0.55; cursor: not-allowed; }

.spinner {
  width: 18px;
  height: 18px;
  border: 2px solid currentColor;
  border-right-color: transparent;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
@media (prefers-reduced-motion: reduce) {
  .btn { transition: none; }
  .spinner { animation-duration: 2s; }
}
</style>
