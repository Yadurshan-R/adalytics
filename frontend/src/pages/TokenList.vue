<!-- The list page: every token with live price, changes, volume and a 24h line.
     Top Volume / Top Gainers / Top Losers views and a search box, kept in the
     address (?view=gainers&q=snek) so a view can be shared or reloaded. -->
<script setup>
import { computed, ref, watch, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useTokens } from '../useTokens.js'
import { formatPrice, fullPrice, formatADA, formatCompact, formatTime } from '../format.js'
import Change from '../components/Change.vue'
import Sparkline from '../components/Sparkline.vue'
import LoadingProgress from '../components/LoadingProgress.vue'

const { tokens, history, error, updatedAt, loading } = useTokens()

// The 7d column and "Last 7 Days" lines appear once 7 days of history are loaded.
const week = computed(() => tokens.value.some((t) => t.change_7d != null))
const sparkLabel = computed(() => (history.value.sparkline_days === 7 ? 'Last 7 Days' : 'Last 24h'))
const route = useRoute()
const router = useRouter()

const views = [
  { key: 'volume', label: 'Top Volume' },
  { key: 'gainers', label: 'Top Gainers' },
  { key: 'losers', label: 'Top Losers' },
]
const view = ref(views.some((v) => v.key === route.query.view) ? route.query.view : 'volume')
const query = ref(typeof route.query.q === 'string' ? route.query.q : '')

// Keep the view and search in the address without adding history entries.
watch([view, query], () => {
  const q = {}
  if (view.value !== 'volume') q.view = view.value
  if (query.value.trim()) q.q = query.value.trim()
  router.replace({ query: q })
})

// A change that rounds to 0.0% is not a gain or a loss.
const moved = (v) => Math.abs(v) >= 0.05

const shown = computed(() => {
  let list = tokens.value.filter((t) => t.price_ada)
  if (view.value === 'gainers') {
    list = list.filter((t) => t.change_24h > 0 && moved(t.change_24h)).sort((a, b) => b.change_24h - a.change_24h)
  } else if (view.value === 'losers') {
    list = list.filter((t) => t.change_24h < 0 && moved(t.change_24h)).sort((a, b) => a.change_24h - b.change_24h)
  }
  const q = query.value.trim().toLowerCase()
  if (q) {
    list = list.filter((t) =>
      t.name.toLowerCase().includes(q) || t.ticker.toLowerCase().includes(q) || t.policy_id.startsWith(q))
  }
  return list
})

const emptyText = computed(() => {
  if (query.value.trim()) return `No token matches “${query.value.trim()}”.`
  if (view.value === 'gainers') return 'No token is up in the last 24 hours.'
  if (view.value === 'losers') return 'No token is down in the last 24 hours.'
  return 'No tokens yet.'
})

const open = (t) => router.push({ name: 'token', params: { ticker: t.id } })

// Press "/" anywhere on the page to jump to the search box, Esc to clear it.
const search = ref(null)
function onKey(e) {
  if (e.key === '/' && document.activeElement !== search.value) {
    e.preventDefault()
    search.value?.focus()
  }
}
onMounted(() => window.addEventListener('keydown', onKey))
onUnmounted(() => window.removeEventListener('keydown', onKey))
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

    <div class="toolbar">
      <div class="views" role="group" aria-label="Order the list">
        <button v-for="v in views" :key="v.key" type="button" :aria-pressed="view === v.key" @click="view = v.key">
          {{ v.label }}
        </button>
      </div>
      <label class="search">
        <svg viewBox="0 0 20 20" aria-hidden="true"><path d="M8.5 3a5.5 5.5 0 1 0 3.4 9.8l3.6 3.7 1.1-1.1-3.7-3.6A5.5 5.5 0 0 0 8.5 3Zm0 1.5a4 4 0 1 1 0 8 4 4 0 0 1 0-8Z" /></svg>
        <input ref="search" v-model="query" type="search" placeholder="Search token" aria-label="Search token by name or ticker"
               autocomplete="off" spellcheck="false" @keydown.esc="query = ''" />
        <kbd v-if="!query" aria-hidden="true">/</kbd>
      </label>
    </div>

    <p v-if="error" class="notice" role="alert">{{ error }}</p>
    <LoadingProgress v-if="history.loading && !error" :loading="history.loading" />

    <div v-else class="table-wrap">
      <table>
        <thead>
          <tr>
            <th class="rank">#</th>
            <th class="left">Token</th>
            <th>Price</th>
            <th class="hide-sm">1h</th>
            <th>24h</th>
            <th v-if="week" class="hide-sm">7d</th>
            <th class="hide-sm">24h Volume</th>
            <th class="hide-md">FDV</th>
            <th class="hide-sm spark-col">{{ sparkLabel }}</th>
          </tr>
        </thead>

        <tbody v-if="loading && !error">
          <tr v-for="n in 10" :key="n" class="skeleton">
            <td class="rank"><i style="width: 12px" /></td>
            <td class="left"><i style="width: 140px" /></td>
            <td><i style="width: 70px" /></td>
            <td class="hide-sm"><i style="width: 44px" /></td>
            <td><i style="width: 44px" /></td>
            <td class="hide-sm"><i style="width: 90px" /></td>
            <td class="hide-md"><i style="width: 70px" /></td>
            <td class="hide-sm spark-col"><i style="width: 130px; height: 30px" /></td>
          </tr>
        </tbody>

        <tbody v-else>
          <tr v-for="(t, i) in shown" :key="t.id" class="row" tabindex="0"
              @click="open(t)" @keydown.enter="open(t)">
            <td class="rank muted num">{{ view === 'volume' ? t.rank : i + 1 }}</td>
            <td class="left">
              <span class="token" :title="`${t.name} (${t.ticker})`">
                <img v-if="t.has_logo" :src="`/api/tokens/${t.id}/logo`" alt="" class="logo" width="24" height="24" loading="lazy" />
                <span class="name">{{ t.name }}</span>
                <span class="ticker muted">{{ t.ticker }}</span>
              </span>
            </td>
            <td class="num price" :title="fullPrice(t.price_ada)">{{ formatPrice(t.price_ada) }}</td>
            <td class="hide-sm"><Change :value="t.change_1h" /></td>
            <td><Change :value="t.change_24h" /></td>
            <td v-if="week" class="hide-sm"><Change :value="t.change_7d" /></td>
            <td class="hide-sm num">{{ formatADA(t.volume_24h_ada) }}</td>
            <td class="hide-md num" :title="t.fdv_ada ? formatADA(t.fdv_ada) : ''">{{ t.fdv_ada ? formatCompact(t.fdv_ada) : '' }}</td>
            <td class="hide-sm spark-col"><Sparkline :values="t.sparkline || []" :width="130" /></td>
          </tr>
          <tr v-if="!shown.length">
            <td :colspan="week ? 9 : 8" class="empty muted">{{ emptyText }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>

<style scoped>
.heading { display: flex; align-items: flex-end; justify-content: space-between; gap: 16px; flex-wrap: wrap; margin-bottom: 20px; }
h1 { font-size: 24px; font-weight: 700; margin: 0; letter-spacing: -0.01em; text-wrap: balance; }
.live { display: inline-flex; align-items: center; gap: 8px; font-size: 13px; margin: 0; }
.dot { width: 8px; height: 8px; border-radius: 50%; background: var(--bull); box-shadow: 0 0 0 4px rgba(22, 199, 132, 0.15); }

.toolbar { display: flex; align-items: center; justify-content: space-between; gap: 12px; flex-wrap: wrap; margin-bottom: 16px; }
.views { display: inline-flex; background: var(--chip); border-radius: 10px; padding: 4px; gap: 2px; }
.views button { font: inherit; font-size: 14px; font-weight: 600; color: var(--muted); background: none; border: 0; padding: 7px 14px; border-radius: 8px; cursor: pointer; white-space: nowrap; }
.views button:hover { color: var(--fg); }
.views button[aria-pressed="true"] { background: var(--bg); color: var(--fg); box-shadow: 0 1px 3px rgba(13, 20, 33, 0.12); }
.views button:focus-visible { outline: 2px solid var(--fg); outline-offset: 1px; }

.search { position: relative; display: flex; align-items: center; width: 260px; }
.search svg { position: absolute; left: 12px; width: 16px; height: 16px; fill: var(--muted); pointer-events: none; }
.search input { width: 100%; font: inherit; font-size: 14px; color: var(--fg); background: var(--chip); border: 1px solid transparent; border-radius: 10px; padding: 9px 36px; outline: none; }
.search input::placeholder { color: var(--muted); }
.search input:focus { background: var(--bg); border-color: #cfd6e4; }
.search input::-webkit-search-cancel-button { display: none; }
.search kbd { position: absolute; right: 10px; font: inherit; font-size: 12px; color: var(--muted); border: 1px solid #d5dbe5; border-radius: 6px; padding: 0 6px; line-height: 18px; }

.notice { margin: 0 0 16px; padding: 12px 16px; border-radius: 10px; background: #fdecee; color: #a0202c; font-weight: 500; }

.table-wrap { overflow-x: auto; }
table { width: 100%; border-collapse: collapse; }
th, td { padding: 16px 10px; text-align: right; white-space: nowrap; border-bottom: 1px solid var(--line); }
th { font-size: 13px; font-weight: 700; color: var(--fg); padding-block: 14px; border-top: 1px solid var(--line); }
td { font-size: 14px; font-weight: 400; }
.left { text-align: left; }
.rank { width: 40px; text-align: left; }

.row { cursor: pointer; }
.row:hover, .row:focus-visible { background: var(--hover); outline: none; }

.token { display: inline-flex; align-items: center; gap: 8px; max-width: 300px; }
.logo { width: 24px; height: 24px; border-radius: 50%; object-fit: cover; flex: none; }
.name { font-weight: 600; min-width: 0; overflow: hidden; text-overflow: ellipsis; }
.ticker { font-weight: 400; flex: none; }
.price { font-weight: 400; }

.spark-col { width: 150px; }
.empty { text-align: center; padding-block: 40px; }

.skeleton i { display: inline-block; height: 14px; border-radius: 6px; background: linear-gradient(90deg, var(--line), #f7f9fb, var(--line)); background-size: 200% 100%; animation: shimmer 1.4s infinite; }
@keyframes shimmer { to { background-position: -200% 0; } }
@media (prefers-reduced-motion: reduce) { .skeleton i { animation: none; } }

@media (max-width: 1000px) { .hide-md { display: none; } }
@media (max-width: 700px) {
  .hide-sm, .ticker { display: none; }
  th, td { padding: 14px 6px; }
  .rank { width: 24px; }
  h1 { font-size: 22px; }
  .toolbar { flex-direction: column; align-items: stretch; }
  .views { display: flex; }
  .views button { flex: 1; padding-inline: 6px; }
  .search { width: 100%; }
  .search kbd { display: none; }
  .token { max-width: 160px; }
}
</style>
