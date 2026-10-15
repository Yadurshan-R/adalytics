<!-- The activity box on a token page, like Birdeye's: pick 5m, 1h, 4h or 24h
     and see the price change, volume, buys and sells, and whether buyers or
     sellers were stronger (net buy and the green/red bar). Refreshed every
     minute from the trades the backend already holds. Right after the backend
     starts, a period that is not fully loaded yet shows its numbers faded. -->
<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { formatCompact } from '../format.js'
import { useTokens } from '../useTokens.js'

const props = defineProps({ id: { type: String, required: true } }) // the token's id
const { day } = useTokens()
// HOURS is each period's length; ready is true once that much history is loaded.
const HOURS = { '5m': 1 / 12, '1h': 1, '4h': 4, '24h': 24 }
const ready = (key) => day.value.ready || day.value.hours >= HOURS[key]

const periods = ref([])
const picked = ref('24h')
let timer = null

// load asks the backend for this token's activity.
async function load() {
  try {
    const res = await fetch(`/api/tokens/${encodeURIComponent(props.id)}/activity`, { cache: 'no-store' })
    if (res.ok) periods.value = (await res.json()).periods ?? []
  } catch {
    // The page shows the last activity it had; the next minute tries again.
  }
}
onMounted(() => {
  load()
  timer = setInterval(load, 60_000)
})
onUnmounted(() => clearInterval(timer))
watch(() => props.id, load)

// p is the period picked now.
const p = computed(() => periods.value.find((x) => x.key === picked.value))
// net is ADA spent buying minus ADA received selling.
const net = computed(() => (p.value ? p.value.buy_ada - p.value.sell_ada : 0))
// buyShare is the buyers' part of the bar, in percent (50 when nothing traded).
const buyShare = computed(() => {
  const total = p.value ? p.value.buy_ada + p.value.sell_ada : 0
  return total > 0 ? (p.value.buy_ada / total) * 100 : 50
})
// pct shows a change like +8.54% or −2.49%.
const pct = (v) => `${v >= 0 ? '+' : '−'}${Math.abs(v).toFixed(2)}%`
</script>

<template>
  <section v-if="periods.length" class="box" aria-label="Trading activity">
    <div class="tabs" role="tablist">
      <button v-for="x in periods" :key="x.key" type="button" role="tab" :aria-selected="x.key === picked" @click="picked = x.key">
        <span class="k">{{ x.key }}</span>
        <span v-if="ready(x.key)" class="num" :class="x.change_pct > 0 ? 'up' : x.change_pct < 0 ? 'down' : 'muted'">{{ pct(x.change_pct) }}</span>
        <span v-else class="num muted" title="Still loading">{{ day.progress }}%</span>
      </button>
    </div>
    <dl v-if="p" class="grid num" :class="{ partial: !ready(p.key) }">
      <div><dt>Volume</dt><dd>{{ formatCompact(p.volume_ada) }}</dd></div>
      <div><dt>Buys</dt><dd class="up">{{ p.buys.toLocaleString('en-US') }}</dd></div>
      <div><dt>Sells</dt><dd class="down">{{ p.sells.toLocaleString('en-US') }}</dd></div>
      <div><dt>Net buy</dt><dd :class="net > 0 ? 'up' : net < 0 ? 'down' : ''">{{ net < 0 ? '−' : net > 0 ? '+' : '' }}{{ formatCompact(Math.abs(net)) }}</dd></div>
      <div><dt>Buy vol</dt><dd class="up">{{ formatCompact(p.buy_ada) }}</dd></div>
      <div><dt>Sell vol</dt><dd class="down">{{ formatCompact(p.sell_ada) }}</dd></div>
    </dl>
    <div class="bar" role="img" :aria-label="`Buyers ${Math.round(buyShare)}%, sellers ${Math.round(100 - buyShare)}%`">
      <span class="buy" :style="{ width: buyShare + '%' }"></span>
    </div>
  </section>
</template>

<style scoped>
.box { margin-top: 20px; border: 1px solid var(--line); border-radius: 12px; overflow: hidden; }
.tabs { display: grid; grid-template-columns: repeat(4, 1fr); border-bottom: 1px solid var(--line); }
.tabs button { display: flex; flex-direction: column; align-items: center; gap: 2px; padding: 10px 4px; border: 0; background: none; font: inherit; cursor: pointer; }
.tabs button + button { border-left: 1px solid var(--line); }
.tabs button[aria-selected="true"] { background: var(--chip); }
.tabs button:focus-visible { outline: 2px solid var(--fg); outline-offset: -2px; }
.k { font-size: 13px; color: var(--muted); }
.tabs .num { font-size: 14px; font-weight: 600; }
.grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px 8px; margin: 0; padding: 14px 16px 12px; }
.grid dt { font-size: 12px; color: var(--muted); }
.grid dd { margin: 2px 0 0; font-size: 14px; font-weight: 600; }
.partial { opacity: 0.4; }
.bar { height: 6px; margin: 0 16px 16px; border-radius: 3px; background: var(--bear); overflow: hidden; }
.buy { display: block; height: 100%; background: var(--bull); }
</style>
