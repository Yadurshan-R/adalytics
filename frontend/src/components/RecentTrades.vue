<!-- A token's recent trades (last 24 hours, newest first), refreshed every
     minute, laid out like Birdeye's trades table:
     - each trade shows the wallet that made it (Trader); click it, or its
       filter icon, to see only that wallet's trades and its summary card;
     - All / Buys / Sells filter the list; trades of ₳10K and more carry
       the BIG or WHALE tag, as on Large Trades;
     - the title is a slot: the token page puts the Trades / Top Traders
       tabs there (TradesSection); showWallet opens one wallet's trades;
     - Buy and Sell are coloured tags, as on Large Trades;
     - 10 rows at first, Show more adds 10 (the backend keeps the last 50). -->
<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { formatPrice, fullPrice, formatAmount, formatAgo } from '../format.js'
import { sizeOf } from '../tradeSizes.js'
import WalletCard from './WalletCard.vue'
import TraderCell from './TraderCell.vue'
import SizeTag from './SizeTag.vue'

const props = defineProps({
  id: { type: String, required: true },
  ticker: { type: String, required: true },
})

const KEEP = 50 // the most the backend keeps per token (MaxTrades)
const PAGE = 10 // rows shown at first, and added by Show more
// SKELETON is the width and class of each cell in the loading rows.
const SKELETON = [{ w: 70, cls: 'left' }, { w: 36, cls: 'left' }, { w: 70, cls: 'hide-sm' }, { w: 60 }, { w: 90, cls: 'hide-sm' }, { w: 110, cls: 'left' }]
const FILTERS = [{ key: 'all', label: 'All' }, { key: 'buy', label: 'Buys' }, { key: 'sell', label: 'Sells' }]

const trades = ref([])
const loaded = ref(false)
const error = ref('')
const now = ref(Date.now())
const filter = ref('all')
const wallet = ref('') // the wallet picked, or '' for all
// hours, mine and trader are for the wallet picked: its time range (24, or
// 7 or 30 days when opened from Top Traders), its trades in this token over
// that range (the latest 200) and its totals (the same as its Top Traders
// row). They come from the backend, not from the 50 latest trades, which on
// a busy token cover less than an hour.
const hours = ref(24)
const mine = ref([])
const trader = ref(null)
const limit = ref(PAGE)
let timer = null
let asked = 0 // drops an answer that comes after a newer question

// get asks the backend for one list and gives its JSON, or throws its error.
async function get(url) {
  const res = await fetch(url, { cache: 'no-store' })
  const data = await res.json()
  if (!res.ok) throw new Error(data.error || `The backend answered ${res.status}`)
  return data
}
// load asks the backend for this token's latest trades (and the picked
// wallet's trades) and keeps any error to show.
async function load() {
  const n = ++asked
  const id = encodeURIComponent(props.id)
  const w = wallet.value
  try {
    const [recent, own] = await Promise.all([get(`/api/tokens/${id}/trades?limit=${KEEP}`),
      w ? get(`/api/tokens/${id}/trades?wallet=${encodeURIComponent(w)}&hours=${hours.value}`) : null])
    if (n !== asked) return
    trades.value = recent.trades ?? []
    mine.value = own?.trades ?? []
    trader.value = own?.trader ?? null
    error.value = ''
  } catch (e) {
    if (n === asked) error.value = e instanceof TypeError ? 'Cannot reach the backend.' : e.message
  } finally {
    if (n === asked) { loaded.value = true; now.value = Date.now() }
  }
}
onMounted(() => {
  load()
  timer = setInterval(load, 60_000)
})
onUnmounted(() => clearInterval(timer))
watch(() => props.id, () => {
  trades.value = []
  pick('', 24)
})

// isBig is true for ₳10K and more (the BIG or WHALE tag), shown in bold.
const isBig = (t) => !!sizeOf(t).tag
// matching is every trade that passes the filter: the latest 50, or the
// picked wallet's own trades.
const matching = computed(() => (wallet.value ? mine.value : trades.value).filter((t) => filter.value === 'all' || t.side === filter.value))
const shown = computed(() => matching.value.slice(0, limit.value))

// setFilter picks a filter and starts again at 10 rows.
function setFilter(key) {
  filter.value = key
  limit.value = PAGE
}
// pick shows one wallet's trades over some hours ('' for everyone's) and
// asks the backend for them.
function pick(w, h) {
  wallet.value = w
  hours.value = h
  mine.value = []
  trader.value = null
  limit.value = PAGE
  loaded.value = false
  load()
}
// setWallet picks a wallet from the table (the same one again: everyone).
const setWallet = (w) => pick(wallet.value === w ? '' : w, 24)
// showWallet shows a wallet's trades over its Top Traders time range.
function showWallet(w, h) {
  filter.value = 'all'
  pick(w, h)
}
defineExpose({ showWallet })

// emptyText explains why no trade is shown.
const emptyText = computed(() => {
  const what = { all: 'trades', buy: 'buys', sell: 'sells' }[filter.value]
  const span = { 24: '24 hours', 168: '7 days', 720: '30 days' }[wallet.value ? hours.value : 24]
  return `No ${what}${wallet.value ? ' by this wallet' : ''} in the last ${span}.`
})

// ₳1,250 for normal amounts, ₳4.35 for small ones.
function ada(v) {
  return '₳' + v.toLocaleString('en-US', { maximumFractionDigits: v < 100 ? 2 : 0 })
}
</script>

<template>
  <section class="trades">
    <div class="top">
      <slot name="title"><h2>Recent Trades</h2></slot>
      <div class="filters" role="group" aria-label="Show">
        <button v-for="f in FILTERS" :key="f.key" type="button" :aria-pressed="f.key === filter" @click="setFilter(f.key)">{{ f.label }}</button>
      </div>
    </div>
    <WalletCard v-if="wallet" :wallet="wallet" :id="id" :trader="trader" :hours="hours" @close="pick('', 24)" />
    <p v-if="error" class="muted note">{{ error }}</p>
    <div v-else class="wrap">
      <table>
        <thead>
          <tr>
            <th class="left">Time</th>
            <th class="left">Type</th>
            <th class="hide-sm">Price</th>
            <th>₳ Amount</th>
            <th class="hide-sm">{{ ticker }}</th>
            <th class="left">Trader</th>
            <th class="link-col"><span class="sr-only">Transaction</span></th>
          </tr>
        </thead>
        <tbody>
          <template v-if="!loaded">
            <tr v-for="n in 5" :key="n" class="skeleton">
              <td v-for="(c, k) in SKELETON" :key="k" :class="c.cls"><i :style="{ width: c.w + 'px' }" /></td>
              <td class="link-col"></td>
            </tr>
          </template>
          <tr v-for="t in shown" v-else :key="`${t.tx_hash}#${t.fill}`">
            <td class="left muted" :title="new Date(t.time).toLocaleString()">{{ formatAgo(t.time, now) }}</td>
            <td class="left"><span class="side" :class="t.side">{{ t.side === 'buy' ? 'Buy' : 'Sell' }}</span></td>
            <td class="hide-sm num" :title="fullPrice(t.price_ada)">{{ formatPrice(t.price_ada) }}</td>
            <td class="num" :class="{ big: isBig(t) }">
              {{ ada(t.amount_ada) }}<SizeTag :ada="t.amount_ada" />
            </td>
            <td class="hide-sm num">{{ formatAmount(t.amount_token) }}</td>
            <td class="left"><TraderCell :wallet="t.wallet" :on="!!wallet && t.wallet === wallet" @pick="setWallet" /></td>
            <td class="link-col">
              <a :href="`https://cardanoscan.io/transaction/${t.tx_hash}`" target="_blank" rel="noopener"
                 class="tx" :aria-label="`View this ${t.side} on Cardanoscan`" title="View on Cardanoscan">
                <svg viewBox="0 0 16 16" aria-hidden="true"><path d="M9 2h5v5h-1.5V4.6L7.1 10 6 8.9l5.4-5.4H9V2ZM3 4h4v1.5H4.5v6h6V9H12v4H3V4Z" /></svg>
              </a>
            </td>
          </tr>
        </tbody>
      </table>
      <p v-if="loaded && !shown.length" class="muted note">{{ emptyText }}</p>
      <p v-if="trader && trader.buys + trader.sells > mine.length && matching.length <= shown.length" class="muted cut">
        These are its latest {{ mine.length }} of {{ trader.buys + trader.sells }} trades; the totals above count them all.
      </p>
      <button v-if="matching.length > shown.length" type="button" class="more" @click="limit += PAGE">
        Show more <span class="muted">({{ shown.length }} of {{ matching.length }})</span>
      </button>
    </div>
  </section>
</template>

<style scoped>
.trades { margin-top: 32px; }
.top { display: flex; justify-content: space-between; align-items: center; gap: 12px; flex-wrap: wrap; margin-bottom: 8px; }
h2 { font-size: 18px; font-weight: 700; margin: 0; }
.filters { display: flex; gap: 2px; font-size: 13px; }
.filters button { height: 30px; padding: 0 10px; border: 0; border-radius: 4px; background: none; color: #6a6d78; font: inherit; font-weight: 500; cursor: pointer; }
.filters button:hover { background: #f0f3fa; color: #131722; }
.filters button[aria-pressed="true"] { color: #2962ff; }
.filters button:focus-visible { outline: 2px solid #2962ff; outline-offset: -2px; }
.note { margin: 0; padding: 24px 0; }
.cut { margin: 12px 0 0; font-size: 13px; }

.wrap { overflow-x: auto; }
table { width: 100%; border-collapse: collapse; }
th, td { padding: 11px 10px; text-align: right; white-space: nowrap; border-bottom: 1px solid var(--line); font-size: 14px; }
th { font-size: 12px; font-weight: 600; color: var(--muted); border-top: 1px solid var(--line); }
.left { text-align: left; }
.side { font-size: 13px; font-weight: 600; }
.side.buy { color: var(--bull); }
.side.sell { color: var(--bear); }
.big { font-weight: 600; }

.link-col { width: 36px; padding-inline: 4px; }
.tx { display: inline-flex; padding: 4px; border-radius: 6px; color: var(--muted); }
.tx:hover { color: var(--fg); background: var(--chip); }
.tx:focus-visible { outline: 2px solid var(--fg); outline-offset: 1px; }
.tx svg { width: 14px; height: 14px; fill: currentColor; }
.sr-only { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0); white-space: nowrap; }
.more { display: block; width: 100%; margin-top: 12px; padding: 9px; border: 1px solid var(--line); border-radius: 8px; background: var(--bg); color: var(--fg); font: inherit; font-size: 13px; font-weight: 600; cursor: pointer; }
.more:hover { background: #f0f3fa; }

.skeleton i { display: inline-block; height: 14px; border-radius: 6px; background: linear-gradient(90deg, var(--line), #f7f9fb, var(--line)); background-size: 200% 100%; animation: shimmer 1.4s infinite; }
@keyframes shimmer { to { background-position: -200% 0; } }
@media (prefers-reduced-motion: reduce) { .skeleton i { animation: none; } }

@media (max-width: 700px) {
  .hide-sm { display: none; }
  th, td { padding: 11px 4px; }
}
</style>
