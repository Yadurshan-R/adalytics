<!-- The activity box on a token page, laid out like Birdeye's: pick 5m, 1h,
     4h or 24h (each tab shows its price change), then the volume and the
     number of trades for that period, each split into buys, sells and net
     with a green/red bar. Refreshed every minute from the trades the backend
     already holds. Right after the backend starts, a period that is not
     fully loaded yet shows its numbers faded. -->
<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { formatCompact } from '../format.js'
import { useTokens } from '../useTokens.js'
import { HOURS, signedPct } from '../listSort.js'

const props = defineProps({ id: { type: String, required: true } }) // the token's id
const { day } = useTokens()
// ready is true once a period's history is fully loaded.
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
// rows are the two lines under the tabs: volume in ADA and number of trades.
const rows = computed(() => {
  if (!p.value) return []
  const { buy_ada: bv, sell_ada: sv, buys, sells } = p.value
  return [
    { label: `${picked.value} Vol`, total: formatCompact(bv + sv), buy: formatCompact(bv), sell: formatCompact(sv), net: bv - sv, netText: formatCompact(Math.abs(bv - sv)), share: share(bv, sv) },
    { label: `${picked.value} Txns`, total: (buys + sells).toLocaleString('en-US'), buy: buys.toLocaleString('en-US'), sell: sells.toLocaleString('en-US'), net: buys - sells, netText: Math.abs(buys - sells).toLocaleString('en-US'), share: share(buys, sells) },
  ]
})
// share is the buyers' part of a bar, in percent (50 when nothing traded).
const share = (b, s) => (b + s > 0 ? (b / (b + s)) * 100 : 50)
const tone = (v) => (v > 0.005 ? 'up' : v < -0.005 ? 'down' : 'muted')
</script>

<template>
  <section v-if="periods.length" class="box" aria-label="Trading activity">
    <div class="tabs" role="tablist">
      <button v-for="x in periods" :key="x.key" type="button" role="tab" :aria-selected="x.key === picked" @click="picked = x.key">
        <span class="k">{{ x.key }}</span>
        <span v-if="ready(x.key)" class="num" :class="tone(x.change_pct)">{{ signedPct(x.change_pct) }}</span>
        <span v-else class="num muted" title="Still loading">{{ day.progress }}%</span>
      </button>
    </div>
    <div v-for="r in rows" :key="r.label" class="row" :class="{ faded: !ready(picked) }">
      <div class="cols num">
        <div><span class="label">{{ r.label }}</span><b>{{ r.total }}</b></div>
        <div class="right"><span class="label">Buy</span><b class="up">{{ r.buy }}</b></div>
        <div class="right"><span class="label">Sell</span><b class="down">{{ r.sell }}</b></div>
        <div class="right"><span class="label">Net</span><b :class="tone(r.net)">{{ r.net < 0 ? '−' : r.net > 0 ? '+' : '' }}{{ r.netText }}</b></div>
      </div>
      <div class="bar" role="img" :aria-label="`Buyers ${Math.round(r.share)}%`"><span :style="{ width: r.share + '%' }"></span></div>
    </div>
  </section>
</template>

<style scoped>
.box { margin-top: 12px; border: 1px solid var(--line); border-radius: 8px; overflow: hidden; }
.tabs { display: grid; grid-template-columns: repeat(4, 1fr); border-bottom: 1px solid var(--line); }
.tabs button { display: flex; flex-direction: column; align-items: center; gap: 3px; padding: 10px 4px; border: 0; background: none; font: inherit; cursor: pointer; }
.tabs button:hover { background: var(--hover); }
.tabs button[aria-selected="true"] { background: var(--chip); }
.tabs button[aria-selected="true"] .k { color: var(--fg); font-weight: 600; }
.tabs button:focus-visible { outline: 2px solid #2962ff; outline-offset: -2px; }
.k { font-size: 13px; color: var(--muted); }
.tabs .num { font-size: 14px; font-weight: 600; }
.row { padding: 12px 14px; }
.row + .row { border-top: 1px solid var(--line); }
.cols { display: grid; grid-template-columns: 1.3fr 1fr 1fr 1fr; gap: 6px; }
.cols div { display: flex; flex-direction: column; gap: 3px; min-width: 0; }
.right { align-items: flex-end; }
.label { font-size: 12px; color: var(--muted); white-space: nowrap; }
.cols b { font-size: 14px; font-weight: 600; white-space: nowrap; }
.bar { height: 4px; margin-top: 10px; border-radius: 2px; background: var(--bear); overflow: hidden; }
.bar span { display: block; height: 100%; background: var(--bull); }
.faded { opacity: 0.4; }
</style>
