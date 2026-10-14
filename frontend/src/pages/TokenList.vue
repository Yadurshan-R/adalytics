<!-- The list page: tokens traded in the chosen period (24h or 7d) with live
     price, changes, volume and a small line; quiet tokens only show up in search.
     Top Volume / Top Gainers / Top Losers views, a 24h / 7d switch and a search
     box, kept in the address (?view=gainers&period=7d&q=snek) so a view can be
     shared or reloaded. The switch decides both what gainers and losers are
     sorted by and what the small lines show, so the two always agree.
     The bar above the table is ListToolbar; the table is TokenTable. -->
<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useTokens } from '../useTokens.js'
import { formatTime } from '../format.js'
import ListToolbar from '../components/ListToolbar.vue'
import TokenTable from '../components/TokenTable.vue'
import LoadingProgress from '../components/LoadingProgress.vue'

const { tokens, history, error, updatedAt, loading } = useTokens()

const route = useRoute()
const router = useRouter()

const views = [
  { key: 'volume', label: 'Top Volume' },
  { key: 'gainers', label: 'Top Gainers' },
  { key: 'losers', label: 'Top Losers' },
]
const view = ref(views.some((v) => v.key === route.query.view) ? route.query.view : 'volume')
const query = ref(typeof route.query.q === 'string' ? route.query.q : '')

// 24h / 7d toggle. It is there when HISTORY_DAYS allows 7 days. While those days
// are still loading, 7d cannot be picked yet and shows its live percentage; the
// list stays on 24h meanwhile.
const period = ref(route.query.period === '7d' ? '7d' : '24h')
// weekRange is the backend's loading status of the 7 day history.
const weekRange = computed(() => history.value.ranges?.find((r) => r.key === '7D'))
// weekReady is true once 7 days of history are loaded.
const weekReady = computed(() => !!weekRange.value?.ready)
// week is true when the list shows 7 day numbers.
const week = computed(() => weekReady.value && period.value === '7d')
// setPeriod picks 24h or 7d; 7d only works once it is loaded.
const setPeriod = (p) => { if (p === '24h' || weekReady.value) period.value = p }
// changeOf is a token's price change for the chosen period.
const changeOf = (t) => (week.value ? t.change_7d ?? 0 : t.change_24h)
// periodText is the period in words, for messages like "No token is up in the last 24 hours".
const periodText = computed(() => (week.value ? 'last 7 days' : 'last 24 hours'))

// Keep the view, period and search in the address without adding history entries.
watch([view, period, query], () => {
  const q = {}
  if (view.value !== 'volume') q.view = view.value
  if (period.value !== '24h') q.period = period.value
  if (query.value.trim()) q.q = query.value.trim()
  router.replace({ query: q })
})

// A change that rounds to 0.0% is not a gain or a loss.
const moved = (v) => Math.abs(v) >= 0.05

// The top lists only show tokens traded in the chosen period: on 24h, tokens
// with trading volume today; on 7d, tokens whose Minswap pool changed in the
// last 7 days. Quiet ones still come up when searched for.
const WEEK_MS = 7 * 24 * 60 * 60 * 1000
// active is true if the token traded in the chosen period.
const active = (t) => (week.value
  ? !!t.last_swap && Date.now() - new Date(t.last_swap).getTime() < WEEK_MS
  : t.volume_24h_ada > 0)

// Place in the Top Volume list (the backend's order without the quiet tokens);
// quiet tokens found by search have no place.
const place = computed(() => {
  const m = {}
  tokens.value.filter((t) => t.price_ada && active(t)).forEach((t, i) => (m[t.id] = i + 1))
  return m
})
// ranks numbers the rows by their Top Volume place, when searching or on Top
// Volume; gainers and losers are numbered 1, 2, 3 instead (null).
const ranks = computed(() => (query.value.trim() || view.value === 'volume' ? place.value : null))

// shown is the list on screen: the chosen view (volume, gainers or losers),
// filtered by the search box.
const shown = computed(() => {
  const q = query.value.trim().toLowerCase()
  let list = tokens.value.filter((t) => t.price_ada && (q || active(t)))
  if (view.value === 'gainers') {
    list = list.filter((t) => changeOf(t) > 0 && moved(changeOf(t))).sort((a, b) => changeOf(b) - changeOf(a))
  } else if (view.value === 'losers') {
    list = list.filter((t) => changeOf(t) < 0 && moved(changeOf(t))).sort((a, b) => changeOf(a) - changeOf(b))
  }
  if (q) {
    list = list.filter((t) =>
      t.name.toLowerCase().includes(q) || t.ticker.toLowerCase().includes(q) || t.policy_id.startsWith(q))
  }
  return list
})

// emptyText explains why the list is empty.
const emptyText = computed(() => {
  if (query.value.trim()) return `No token matches “${query.value.trim()}”.`
  if (view.value === 'gainers') return `No token is up in the ${periodText.value}.`
  if (view.value === 'losers') return `No token is down in the ${periodText.value}.`
  return 'No tokens yet.'
})
</script>

<template>
  <section>
    <div class="heading">
      <h1>Token Prices</h1>
      <p v-if="updatedAt && !error" class="live muted num">
        <span class="dot" aria-hidden="true"></span>
        Updated {{ formatTime(updatedAt) }}
      </p>
    </div>

    <ListToolbar v-model:view="view" v-model:query="query" :views="views" :week="week"
                 :week-range="weekRange" :week-ready="weekReady" @period="setPeriod" />

    <p v-if="error" class="notice" role="alert">{{ error }}</p>
    <LoadingProgress v-if="history.loading && !error" :loading="history.loading" />

    <TokenTable v-else :rows="shown" :ranks="ranks" :loading="loading && !error" :week="week"
                :week-ready="weekReady" :empty-text="emptyText" />
  </section>
</template>

<style scoped>
.heading { display: flex; align-items: flex-end; justify-content: space-between; gap: 16px; flex-wrap: wrap; margin-bottom: 20px; }
h1 { font-size: 24px; font-weight: 700; margin: 0; letter-spacing: -0.01em; text-wrap: balance; }
.live { display: inline-flex; align-items: center; gap: 8px; font-size: 13px; margin: 0; }
.dot { width: 8px; height: 8px; border-radius: 50%; background: var(--bull); box-shadow: 0 0 0 4px rgba(22, 199, 132, 0.15); }

.notice { margin: 0 0 16px; padding: 12px 16px; border-radius: 10px; background: #fdecee; color: #a0202c; font-weight: 500; }

@media (max-width: 700px) { h1 { font-size: 22px; } }
</style>
