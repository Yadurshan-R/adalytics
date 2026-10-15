<!-- The list page, laid out like Birdeye's: every listed Cardano token (a
     Minswap pool with at least ₳10,000), quiet ones too, with price, FDV,
     volume, trades, liquidity and the 5m, 1h, 4h and 24h change.
     - Top Volume / Gainers / Losers, and the period (5m, 1h, 4h, 24h) that
       Volume, Txns, Gainers and Losers follow.
     - Click a column title to sort by it; click again to turn the order round.
     - The view, period, sort and search are kept in the address, so a view can
       be shared or reloaded (?view=gainers&period=1h&sort=liquidity&q=snek).
     The bar above the table is ListToolbar; the table is TokenTable. -->
<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useTokens } from '../useTokens.js'
import { formatTime } from '../format.js'
import { PERIODS, HOURS, COLUMNS, periodOf, sortRows } from '../listSort.js'
import ListToolbar from '../components/ListToolbar.vue'
import TokenTable from '../components/TokenTable.vue'
import LoadingProgress from '../components/LoadingProgress.vue'

const { tokens, history, day, error, updatedAt, loading } = useTokens()
const route = useRoute()
const router = useRouter()

const VIEWS = [
  { key: 'volume', label: 'Top Volume' },
  { key: 'gainers', label: 'Gainers' },
  { key: 'losers', label: 'Losers' },
]
const pick = (v, list, def) => (list.includes(v) ? v : def)
const view = ref(pick(route.query.view, VIEWS.map((v) => v.key), 'volume'))
const period = ref(pick(route.query.period, PERIODS, '24h'))
const query = ref(typeof route.query.q === 'string' ? route.query.q : '')
// sort is the column picked by the visitor ({ key, dir }), or null for the
// view's own order.
const sort = ref(COLUMNS.some((c) => c.key === route.query.sort) ? { key: route.query.sort, dir: route.query.dir === 'asc' ? 1 : -1 } : null)

// setSort sorts by a column: highest first, then lowest first on a second click.
function setSort(key) {
  sort.value = sort.value?.key === key && sort.value.dir === -1 ? { key, dir: 1 } : { key, dir: -1 }
}
// setView picks Top Volume, Gainers or Losers and drops the column sort.
function setView(v) {
  view.value = v
  sort.value = null
}

// Keep the view, period, sort and search in the address without adding history entries.
watch([view, period, sort, query], () => {
  const q = {}
  if (view.value !== 'volume') q.view = view.value
  if (period.value !== '24h') q.period = period.value
  if (sort.value) Object.assign(q, { sort: sort.value.key, dir: sort.value.dir === 1 ? 'asc' : 'desc' })
  if (query.value.trim()) q.q = query.value.trim()
  router.replace({ query: q })
})

// periodReady is true once the picked period's history is fully loaded.
const periodReady = computed(() => day.value.ready || day.value.hours >= HOURS[period.value])
const changeOf = (t) => periodOf(t, period.value).change_pct
// A change that rounds to 0.00% is not a gain or a loss.
const moved = (v) => Math.abs(v) >= 0.005

// shown is the list on screen: every listed token, filtered by the search box
// and the view, in the view's order or the column picked.
const shown = computed(() => {
  const q = query.value.trim().toLowerCase()
  let list = tokens.value.filter((t) => t.price_ada)
  if (q) list = list.filter((t) => t.name.toLowerCase().includes(q) || t.ticker.toLowerCase().includes(q) || t.policy_id.startsWith(q))
  if (view.value === 'gainers') list = list.filter((t) => moved(changeOf(t)) && changeOf(t) > 0)
  if (view.value === 'losers') list = list.filter((t) => moved(changeOf(t)) && changeOf(t) < 0)
  const own = { volume: { key: 'volume', dir: -1 }, gainers: { key: `c${period.value}`, dir: -1 }, losers: { key: `c${period.value}`, dir: 1 } }
  return sortRows(list, sort.value ?? own[view.value], period.value)
})

// emptyText explains why the list is empty.
const emptyText = computed(() => {
  if (query.value.trim()) return `No token matches “${query.value.trim()}”.`
  const span = { '5m': '5 minutes', '1h': 'hour', '4h': '4 hours', '24h': '24 hours' }[period.value]
  if (view.value === 'gainers') return `No token is up in the last ${span}.`
  if (view.value === 'losers') return `No token is down in the last ${span}.`
  return 'No tokens yet.'
})
</script>

<template>
  <section>
    <div class="heading">
      <h1>Cardano Tokens</h1>
      <p v-if="updatedAt && !error" class="live muted num">
        <span class="dot" aria-hidden="true"></span>
        {{ tokens.length }} tokens · updated {{ formatTime(updatedAt) }}
      </p>
    </div>

    <ListToolbar :views="VIEWS" :view="view" :periods="PERIODS" v-model:period="period" v-model:query="query" @update:view="setView" />

    <p v-if="error" class="notice" role="alert">{{ error }}</p>
    <LoadingProgress v-if="history.loading && !error" :loading="history.loading" />
    <template v-else>
      <div v-if="!day.ready" class="day" role="status">
        <span>Loading the last 24 hours <b class="num">{{ day.progress }}%</b> · prices are live, longer periods are filling in</span>
        <span class="bar"><span :style="{ width: Math.max(day.progress, 2) + '%' }"></span></span>
      </div>
      <TokenTable :rows="shown" :period="period" :sort="sort" :view="view" :loading="loading && !error"
                  :hours="day.ready ? 24 : day.hours" :period-ready="periodReady" :empty-text="emptyText" @sort="setSort" />
    </template>
  </section>
</template>

<style scoped>
.heading { display: flex; align-items: baseline; justify-content: space-between; gap: 16px; flex-wrap: wrap; margin-bottom: 14px; }
h1 { font-size: 20px; font-weight: 700; margin: 0; letter-spacing: -0.01em; }
.live { display: inline-flex; align-items: center; gap: 8px; font-size: 13px; margin: 0; }
.dot { width: 7px; height: 7px; border-radius: 50%; background: var(--bull); }
.day { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; margin: 0 0 10px; font-size: 13px; color: var(--muted); }
.day b { color: var(--fg); font-weight: 600; }
.bar { flex: 1; min-width: 120px; max-width: 240px; height: 4px; border-radius: 2px; background: var(--chip); overflow: hidden; }
.bar span { display: block; height: 100%; background: #2962ff; transition: width 0.6s ease; }
@media (prefers-reduced-motion: reduce) { .bar span { transition: none; } }
.notice { margin: 0 0 16px; padding: 12px 16px; border-radius: 8px; background: #fdecee; color: #a0202c; font-weight: 500; }
</style>
