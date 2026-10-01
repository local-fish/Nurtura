<script lang="ts">
import type { NavItem } from '../../layout/components/interfaces';
import { Temporal } from '@js-temporal/polyfill';
import GrowthCard from './components/GrowthCard.svelte';
import Card from '../../components/Card.svelte';
import Icon from '../../components/Icon.svelte';
import Button from '../../components/Button.svelte';
import Main from '../../layout/main.svelte';

let active = $state('beranda');
const user = { name: 'placholder', age: 'placeholder', avatar: '' };
</script>

<Main
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
        <div class="imun-heading">
          <Icon name="calendar" size={24}/>
          <strong style="font-size: 0.9rem, text-align: center">Imunisasi Berikutnya</strong>
        </div>
          <h1> Polio 3 </h1>
          <div style="margin-bottom: 1rem">Lakukan sebelum [placeholder]. [placeholder]</div>
          <Button icon="check">Tandai Sudah Vaksin</Button>
      </Card>
    </div>

    <div class="row bottom">
      <Card bg="var(--surface)" border="1px solid var(--teal-100)">
        <h4> Catatan Terakhir </h4>
      </Card>
      <Card bg="var(--surface)" border="1px solid var(--teal-100)">
        <h4> Akses Cepat </h4>
      </Card>
    </div>
  </section>
</Main>

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
  .imun-heading {
    display: flex;
    flex-direction: row;
    gap: 8px;
    align-content: center;
  }

</style>
