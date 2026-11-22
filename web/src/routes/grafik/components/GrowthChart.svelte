<script lang="ts">
	import Chart from 'chart.js/auto';
	import type { ChartConfiguration, Plugin } from 'chart.js';
	import type { RefPoint, GrowthPoint } from '$src/lib/api';

	let {
		reference = [],
		records = [],
		unit = '',
		label = '',
		caption = '',
		maxAge = 24
	}: {
		reference?: RefPoint[];
		records?: GrowthPoint[];
		unit?: string;
		label?: string;
		caption?: string;
		maxAge?: number;
	} = $props();

	const OUTSIDE = '#F6D98A';
	const BAND = 'rgba(26, 138, 143, 0.28)';
	const MEDIAN = '#0f7378';
	const CHILD = '#143d4a';

	let canvas: HTMLCanvasElement;
	let chart: Chart | null = null;

	/** Tints the whole plot area as "out of range", so the p3-p97 band drawn on
	 *  top of it reads as the normal corridor without needing extra datasets. */
	const outsideBg: Plugin = {
		id: 'outsideBg',
		beforeDatasetsDraw(c) {
			const { ctx, chartArea } = c;
			if (!chartArea) return;
			ctx.save();
			ctx.fillStyle = OUTSIDE;
			ctx.fillRect(
				chartArea.left,
				chartArea.top,
				chartArea.right - chartArea.left,
				chartArea.bottom - chartArea.top
			);
			ctx.restore();
		}
	};

	/** Value labels above the child's points, as in the design. Skipped when the
	 *  series is dense enough that labels would collide. */
	const pointLabels: Plugin = {
		id: 'pointLabels',
		afterDatasetsDraw(c) {
			const meta = c.getDatasetMeta(3);
			if (!meta || meta.hidden) return;
			const pts = meta.data ?? [];
			if (pts.length === 0 || pts.length > 14) return;
			const { ctx } = c;
			ctx.save();
			ctx.font = '600 11px "Plus Jakarta Sans", system-ui, sans-serif';
			ctx.fillStyle = CHILD;
			ctx.textAlign = 'center';
			ctx.textBaseline = 'bottom';
			pts.forEach((p, i) => {
				// Read the value off the dataset rather than the source array: the
				// dataset holds the range-filtered points, so indexes differ.
				const v = (c.data.datasets[3].data[i] as { y: number } | undefined)?.y;
				if (v === undefined) return;
				ctx.fillText(`${v}${unit}`, p.x, p.y - 8);
			});
			ctx.restore();
		}
	};

	const visibleRef = $derived(reference.filter((r) => r.age_months <= maxAge));
	const visibleRecs = $derived(records.filter((r) => r.age_months <= maxAge));

	function config(): ChartConfiguration {
		return {
			type: 'line',
			data: {
				datasets: [
					{
						label: 'Persentil 97',
						data: visibleRef.map((r) => ({ x: r.age_months, y: r.p97 })),
						borderColor: 'rgba(26, 138, 143, 0.45)',
						borderWidth: 1,
						pointRadius: 0,
						// Fills down to the p3 dataset, painting the normal corridor.
						fill: '+1',
						backgroundColor: BAND,
						tension: 0.35
					},
					{
						label: 'Persentil 3',
						data: visibleRef.map((r) => ({ x: r.age_months, y: r.p3 })),
						borderColor: 'rgba(26, 138, 143, 0.45)',
						borderWidth: 1,
						pointRadius: 0,
						fill: false,
						tension: 0.35
					},
					{
						label: 'Median (P50)',
						data: visibleRef.map((r) => ({ x: r.age_months, y: r.p50 })),
						borderColor: MEDIAN,
						borderWidth: 1.5,
						borderDash: [5, 4],
						pointRadius: 0,
						fill: false,
						tension: 0.35
					},
					{
						label,
						data: visibleRecs.map((r) => ({ x: r.age_months, y: r.value })),
						borderColor: CHILD,
						borderWidth: 2.5,
						pointRadius: 4,
						pointBackgroundColor: '#fff',
						pointBorderColor: CHILD,
						pointBorderWidth: 2,
						fill: false,
						tension: 0.3
					}
				]
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				interaction: { mode: 'nearest', intersect: false },
				layout: { padding: { top: 18, right: 12 } },
				scales: {
					x: {
						type: 'linear',
						min: 0,
						max: maxAge,
						title: { display: true, text: 'Usia (bulan)', color: '#5d6b69' },
						ticks: {
							stepSize: maxAge <= 6 ? 1 : maxAge <= 12 ? 2 : 3,
							color: '#5d6b69',
							callback: (v) => (Number(v) === 0 ? 'Lahir' : `${v}m`)
						},
						grid: { color: 'rgba(22, 33, 31, 0.08)' }
					},
					y: {
						title: { display: true, text: caption, color: '#5d6b69' },
						ticks: { color: '#5d6b69' },
						grid: { color: 'rgba(22, 33, 31, 0.08)' }
					}
				},
				plugins: {
					legend: { display: false },
					tooltip: {
						callbacks: {
							title: (items) => {
								const m = Number(items[0]?.parsed.x);
								return m === 0 ? 'Lahir' : `Usia ${m} bulan`;
							},
							label: (i) => `${i.dataset.label}: ${i.parsed.y}${unit}`
						}
					}
				}
			},
			plugins: [outsideBg, pointLabels]
		};
	}

	$effect(() => {
		// Touch the reactive inputs so the chart rebuilds when the metric,
		// range, or freshly loaded data changes.
		void visibleRef;
		void visibleRecs;
		void label;
		void unit;
		void maxAge;

		chart?.destroy();
		chart = new Chart(canvas, config());
		return () => {
			chart?.destroy();
			chart = null;
		};
	});

	/** PNG of the current chart, for the "Unduh grafik" action. */
	export function toDataURL(): string | null {
		if (!chart) return null;
		// Composite onto white: the canvas itself is transparent outside the plot.
		const src = chart.canvas;
		const out = document.createElement('canvas');
		out.width = src.width;
		out.height = src.height;
		const ctx = out.getContext('2d');
		if (!ctx) return null;
		ctx.fillStyle = '#ffffff';
		ctx.fillRect(0, 0, out.width, out.height);
		ctx.drawImage(src, 0, 0);
		return out.toDataURL('image/png');
	}
</script>

<div class="wrap">
	<canvas bind:this={canvas} aria-label={`Grafik ${label}`}></canvas>
</div>

<style>
	.wrap {
		position: relative;
		width: 100%;
		height: 380px;
	}
	@media (max-width: 900px) {
		.wrap {
			height: 300px;
		}
	}
</style>
