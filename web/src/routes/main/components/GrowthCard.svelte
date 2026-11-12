<script lang="ts">
	import type { Component } from 'svelte';
	import Button from '../../../components/Button.svelte';

	let {
		name = 'placeholder',
		updatedAt = 'placeholder',
		metrics = [
			{ label: 'placeholder', value: '-', unit: '', status: null },
			{ label: 'placeholder', value: '-', unit: '', status: null },
			{ label: 'placeholder', value: '-', unit: '', status: null }
		]
	} = $props();
</script>

<section class="card">
	<span class="circle" aria-hidden="true"></span>

	<header class="head">
		<div>
			<h2>Pertumbuhan {name}</h2>
			<p class="updated">Pembaruan terakhir · {updatedAt}</p>
		</div>
		<Button style="width: 18rem" icon="arrow-up-right">
			<span style="font-size: 0.9rem">Lihat Grafik Lengkap</span>
		</Button>
	</header>
	<div class="metrics">
		{#each metrics as m}
			<div class="metric">
				<span class="label">{m.label}</span>
				<div class="value">
					{m.value}<span class="unit">{m.unit}</span>
				</div>
				{m.status}
			</div>
		{/each}
	</div>
</section>

<style>
.card {
  --bg: #7fa996;
  --text: #ffffff;
  --muted: rgba(255, 255, 255, 0.82);
  --badge-bg: #cdf0c8;
  --badge-text: #1f5a3a;
  --cta-text: #1b7f86;

  flex: 1;

  position: relative;
  overflow: hidden; 
  box-sizing: border-box;
  padding: 28px 30px 26px;
  border-radius: 2rem;
  background: var(--bg);
  color: var(--text);
  font-family: 'Plus Jakarta Sans', system-ui, sans-serif;
}

.circle {
  position: absolute;
  top: -40px;
  right: -60px;
  width: 300px;
  height: 300px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.12);
  pointer-events: none;
}

.head {
  position: relative;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
}

h2 {
  margin: 0;
  font-size: 26px;
  font-weight: 800;
  letter-spacing: -0.01em;
}

.updated {
  margin: 6px 0 0;
  font-size: 13px;
  color: var(--muted);
}

.cta {
  display: inline-flex;
  justify-content: center;
  align-items: center;
  gap: 8px;
  padding: 12px 22px;
  border-radius: 14px;
  background: #fff;
  color: var(--cta-text);
  font-size: 13px;
  font-weight: 700;
  text-decoration: none;,m,,lk
  white-space: nowrap;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.08);
  }
  .cta:focus-visible {
    outline: 3px solid rgba(255, 255, 255, 0.7);
    outline-offset: 2px;
  }

  .metrics {
    position: relative;
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    margin-top: 26px;
  }

  .metric {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 6px;
    padding: 0 28px;
  }
  .metric:first-child { padding-left: 0; }
  .metric + .metric { border-left: 1px solid rgba(255, 255, 255, 0.28); }

  .label {
    font-size: 13px;
    color: var(--muted);
  }

  .value {
    font-size: 34px;
    font-weight: 800;
    letter-spacing: 0.02em;
    line-height: 1.15;
  }
  .unit {
    margin-left: 6px;
    font-size: 15px;
    font-weight: 600;
    letter-spacing: 0;
  }

  .badge {
    margin-top: 4px;
    padding: 5px 12px;
    border-radius: 8px;
    background: var(--badge-bg);
    color: var(--badge-text);
    font-size: 12px;
    font-weight: 700;
  }

  @media (max-width: 640px) {
    .card { padding: 22px 20px; border-radius: 26px; }
    .head { flex-direction: column; }
    .metrics { grid-template-columns: 1fr; gap: 18px; }
    .metric { padding: 0; }
    .metric + .metric {
      border-left: 0;
      border-top: 1px solid rgba(255, 255, 255, 0.28);
      padding-top: 18px;
    }
  }
</style>
