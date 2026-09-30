<script lang="ts">
import type { NavItem } from '$src/layout/components/interfaces';
import Layout from '$src/layout/main.svelte';
import { Temporal } from '@js-temporal/polyfill';
import GrowthCard from './components/GrowthCard.svelte';
import Card from '../../components/Card.svelte';

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
  <section class="dashboard">
  <div class="row top">
    <GrowthCard />

    <Card bg="#E3AD53">
      <h1> Polio 3 </h1>
    </Card>
  </div>

  <div class="row bottom">
  </div>
  </section>
</Layout>

<style>
  .dashboard {
    display: flex;
    flex-direction: column;
    gap: 24px;
    margin: 0 auto;
    padding: 1rem;
  }
  .row {
    display: grid;
    gap: 24px;
    align-items: stretch;
  }
  .top    { grid-template-columns: 2.2fr 1fr; }
  .bottom { grid-template-columns: 1.8fr 1fr; }
</style>
