<!-- A token's top traders, like the Top Traders tab on Birdeye and
     DexScreener: the wallets that traded this token most in the last 24
     hours, 7 days or 30 days, from the trades the backend already keeps (no
     extra Koios requests).
     - Volume ranks by ADA bought plus sold; Net ranks by net buying: ADA
       spent on buys minus ADA received from sells, the same Net as the
       wallet card (above zero, the wallet bought more than it sold).
     - Net is not profit: it does not know what the tokens are worth now.
     - 7 and 30 days are grey with "Loading 35%" until that history is in.
     - Clicking a wallet shows its trades over the same time range in the
       Trades tab, with the same totals as its row.
     - 10 rows at first, Show more adds 10 (the backend sends the top 50).
     - One token per tab: TradesSection gives each token its own. -->
<script setup>
import { ref, computed, watch, onActivated, onDeactivated } from 'vue'
import { formatAmount, formatAgo } from '../format.js'

const props = defineProps({
  id: { type: String, required: true },
  ticker: { type: String, required: true },
})
defineEmits(['wallet'])

const PAGE = 10
const TIMES = [{ hours: 24, label: '24H' }, { hours: 168, label: '7D', range: '7D' }, { hours: 720, label: '30D', range: '30D' }]
const SORTS = [{ key: 'volume', label: 'Volume' }, { key: 'net', label: 'Net' }]

const traders = ref([])
const ranges = ref([])
const loaded = ref(false)
const error = ref('')
const hours = ref(24)
const sort = ref('volume')
const limit = ref(PAGE)
const now = ref(Date.now())
let timer = null
let asked = 0 // drops an answer that comes after a newer question

// load asks the backend for this token's top traders and keeps any error to show.
async function load() {
  const n = ++asked
  try {
    const res = await fetch(`/api/tokens/${encodeURIComponent(props.id)}/traders?hours=${hours.value}&sort=${sort.value}`, { cache: 'no-store' })
    const data = await res.json()
    if (n !== asked) return
    if (!res.ok) throw new Error(data.error || `The backend answered ${res.status}`)
    traders.value = data.traders ?? []
    ranges.value = data.ranges ?? []
    error.value = ''
  } catch (e) {
    if (n === asked) error.value = e instanceof TypeError ? 'Cannot reach the backend.' : e.message
  } finally {
    if (n === asked) { loaded.value = true; now.value = Date.now() }
  }
}
// The tab is kept (KeepAlive in TradesSection), so its time range and sort
// stay when the visitor looks at a wallet's trades and comes back. It loads
// and refreshes every minute only while it is open.
onActivated(() => { load(); timer = setInterval(load, 60_000) })
onDeactivated(() => clearInterval(timer))
watch([hours, sort], () => { loaded.value = false; limit.value = PAGE; load() })

// waiting is how much of a time choice is loaded, or null when it can be picked.
function waiting(t) {
  if (!t.range) return null
  const r = ranges.value.find((x) => x.key === t.range)
  return r?.ready ? null : r?.progress ?? 0
}
const shown = computed(() => traders.value.slice(0, limit.value))
const net = (t) => t.buy_ada - t.sell_ada
// ada writes an ADA amount: ₳12,400, or a dash for nothing.
function ada(v) {
  return v ? '₳' + Math.round(v).toLocaleString('en-US') : '—'
}
function signed(v) {
  return (v > 0 ? '+' : v < 0 ? '−' : '') + ada(Math.abs(v)).replace('—', '₳0')
}
</script>

<template>
  <div class="bar">
    <slot name="title" />
    <div class="right">
    <div class="filters" role="group" aria-label="Time">
      <button v-for="t in TIMES" :key="t.hours" type="button" :aria-pressed="t.hours === hours" :disabled="waiting(t) !== null"
              :title="waiting(t) !== null ? `Loading ${waiting(t)}%` : null" @click="hours = t.hours">{{ t.label }}</button>
    </div>
    <div class="filters" role="group" aria-label="Rank by">
      <span class="muted by">Rank by</span>
      <button v-for="s in SORTS" :key="s.key" type="button" :aria-pressed="s.key === sort" @click="sort = s.key">{{ s.label }}</button>
    </div>
    </div>
  </div>
  <p v-if="error" class="muted note">{{ error }}</p>
  <div v-else class="wrap">
    <table>
      <thead>
        <tr>
          <th class="left rank">#</th>
          <th class="left">Trader</th>
          <th>Bought</th>
          <th>Sold</th>
          <th title="ADA spent buying minus ADA received selling. Above zero: the wallet bought more than it sold.">Net ₳</th>
          <th class="hide-sm">Trades</th>
          <th class="hide-sm">Last trade</th>
        </tr>
      </thead>
      <tbody :class="{ dim: !loaded }">
        <tr v-for="(t, i) in shown" :key="t.wallet">
          <td class="left rank muted">{{ i + 1 }}</td>
          <td class="left">
            <button type="button" class="trader" :title="`See ${t.wallet}'s trades`" @click="$emit('wallet', t.wallet, hours)">{{ t.wallet.slice(0, 6) }}<span class="mid">{{ t.wallet.slice(6, 10) }}</span>…{{ t.wallet.slice(-4) }}</button>
          </td>
          <td class="num">{{ ada(t.buy_ada) }}<small v-if="t.buy_token">{{ formatAmount(t.buy_token) }} {{ ticker }}</small></td>
          <td class="num">{{ ada(t.sell_ada) }}<small v-if="t.sell_token">{{ formatAmount(t.sell_token) }} {{ ticker }}</small></td>
          <td class="num" :class="net(t) > 0 ? 'up' : net(t) < 0 ? 'down' : ''">{{ signed(net(t)) }}</td>
          <td class="hide-sm num"><span class="up">{{ t.buys }}</span> <span class="muted">/</span> <span class="down">{{ t.sells }}</span></td>
          <td class="hide-sm muted" :title="new Date(t.last).toLocaleString()">{{ formatAgo(t.last, now) }}</td>
        </tr>
      </tbody>
    </table>
    <p v-if="loaded && !traders.length" class="muted note">No trades with a known wallet in this time.</p>
    <button v-if="traders.length > shown.length" type="button" class="more" @click="limit += PAGE">
      Show more <span class="muted">({{ shown.length }} of {{ traders.length }})</span>
    </button>
  </div>
</template>

<style scoped>
.bar { display: flex; justify-content: space-between; align-items: center; gap: 12px; flex-wrap: wrap; margin-bottom: 8px; }
.filters { display: flex; align-items: center; gap: 2px; font-size: 13px; }
.right { display: flex; gap: 16px; flex-wrap: wrap; }
.by { margin-right: 4px; font-size: 12px; }
.filters button { height: 28px; padding: 0 10px; border: 0; border-radius: 4px; background: none; color: #6a6d78; font: inherit; font-weight: 500; cursor: pointer; }
.filters button:hover:not(:disabled) { background: #f0f3fa; color: #131722; }
.filters button[aria-pressed="true"] { color: #2962ff; background: #f0f3fa; }
.filters button:disabled { opacity: 0.45; cursor: default; }
.note { margin: 0; padding: 24px 0; }
.wrap { overflow-x: auto; }
table { width: 100%; border-collapse: collapse; }
th, td { padding: 10px; text-align: right; white-space: nowrap; border-bottom: 1px solid var(--line); font-size: 14px; vertical-align: top; }
th { font-size: 12px; font-weight: 600; color: var(--muted); border-top: 1px solid var(--line); }
.left { text-align: left; }
.rank { width: 28px; padding-right: 0; }
small { display: block; margin-top: 2px; color: var(--muted); font-size: 12px; }
.up { color: var(--bull); }
.down { color: var(--bear); }
td.up, td.down { font-weight: 600; }
.dim { opacity: 0.5; }
.trader { padding: 2px 4px; margin: -2px -4px; border: 0; border-radius: 4px; background: none; color: #131722; font: 13px ui-monospace, SFMono-Regular, Menlo, monospace; cursor: pointer; }
.trader:hover { background: #f0f3fa; color: #2962ff; }
.more { display: block; width: 100%; margin-top: 12px; padding: 9px; border: 1px solid var(--line); border-radius: 8px; background: var(--bg); color: var(--fg); font: inherit; font-size: 13px; font-weight: 600; cursor: pointer; }
.more:hover { background: #f0f3fa; }
/* The columns fit the space the table has (the chart column is narrow on
   small laptops): Trades and Last trade go first, then the token amounts. */
.wrap { container-type: inline-size; }
@container (max-width: 640px) { .hide-sm { display: none; } }
@container (max-width: 440px) {
  th, td { padding: 10px 4px; font-size: 13px; }
  small, .mid { display: none; }
  .rank { width: 16px; }
}
</style>
