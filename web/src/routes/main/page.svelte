<script lang="ts">
import type { NavItem } from '$src/layout/components/interfaces';
import Layout from '$src/layout/main.svelte';
import { Temporal } from '@js-temporal/polyfill';

let active = $state('beranda');
const user = { name: 'placholder', age: 'placeholder', avatar: '' };
</script>

<Layout
  {active}
  {user}
  title="Selamat pagi, [Placeholder]"
  subtitle={Temporal.Now.zonedDateTimeISO(Temporal.Now.timeZoneId()).toLocaleString("id-ID", { day: 'numeric', month: 'long', year: 'numeric' })}
  onnavigate={(item: NavItem, e: MouseEvent) => { e.preventDefault(); active = item.id; }}
>
  <div>text</div>
</Layout>

