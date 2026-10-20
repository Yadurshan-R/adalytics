<!-- The Large Trades page, laid out like Birdeye's Find Trades: a filter bar
     and one table of the big trades across every token, up to 30 days back.
     Who uses it: traders who want to see where big money is going without
     opening each token. A big trade is at least 1% of its pool (it moves the
     price about 2%) or ₳10K and more; the size icons (fish, dolphin, whale,
     like Birdeye's) sort them by amount. The backend filters, and sends 200
     trades at a time (Show more for the next 200). The filters are kept in
     the address, so a view can be shared. Refreshed every minute. -->
<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import TradeFilters from '../components/TradeFilters.vue'
import TradesTable from '../components/TradesTable.vue'
import { SIZES } from '../tradeSizes.js'

const route = useRoute()
const router = useRouter()
const q = route.query
const HOURS = [1, 4, 24, 168, 720]
const PAGE = 200 // trades per page, and per Show more (the backend's PageWhales)
const token = ref(String(q.token ?? ''))
const hours = ref(HOURS.includes(Number(q.hours)) ? Number(q.hours) : 24)
const size = ref(SIZES.some((s) => s.key === q.size) ? q.size : 'all')
const wallet = ref(String(q.wallet ?? ''))

const trades = ref([])
const tokens = ref([])
const ranges = ref([]) // how much of 7D and 30D is loaded
const more = ref(false)
const loaded = ref(false)
const loadingMore = ref(false)
const error = ref('')
const now = ref(Date.now())
let timer = null
let typing = null
let asked = 0 // counts the questions sent, so a late answer to an old one is dropped

// ask builds the address for one page of trades with the filters picked.
function ask(offset, limit) {
  const p = new URLSearchParams({ hours: hours.value, offset, limit })
  const s = SIZES.find((x) => x.key === size.value) // none for All
  if (s?.min) p.set('min', s.min)
  if (s?.max) p.set('max', s.max)
  if (token.value) p.set('token', token.value)
  if (wallet.value) p.set('wallet', wallet.value)
  return `/api/whales?${p}`
}
// load asks the backend for the newest trades again, as many as are shown
// (so a refresh keeps the pages already opened with Show more), or the
// next 200 when next is true; after a filter change (fresh), one page.
// Only the answer to the latest question is
// used, so a slow answer for old filters never shows under new ones.
async function load(next = false, fresh = false) {
  const mine = ++asked
  try {
    loadingMore.value = next
    const rows = next ? trades.value.length : fresh ? PAGE : Math.max(PAGE, Math.min(trades.value.length, 1000))
    const res = await fetch(ask(next ? rows : 0, next ? PAGE : rows), { cache: 'no-store' })
    if (!res.ok) throw new Error(`The backend answered ${res.status}`)
    const data = await res.json()
    if (mine !== asked) return
    trades.value = next ? addNew(trades.value, data.trades ?? []) : data.trades ?? []
    more.value = data.more
    tokens.value = data.tokens ?? []
    ranges.value = data.ranges ?? []
    error.value = ''
  } catch (e) {
    if (mine === asked) error.value = e instanceof TypeError ? 'Cannot reach the backend.' : e.message
  } finally {
    if (mine === asked) {
      loaded.value = true
      loadingMore.value = false
      now.value = Date.now()
    }
  }
}
// addNew adds the next page without repeating a trade: new trades arriving
// between two pages move the rest down, so the next page can start with a
// few already shown.
function addNew(shown, page) {
  const key = (t) => `${t.tx_hash}#${t.fill}:${t.id}`
  const seen = new Set(shown.map(key))
  return shown.concat(page.filter((t) => !seen.has(key(t))))
}
onMounted(() => {
  load()
  timer = setInterval(() => load(), 60_000)
})
onUnmounted(() => clearInterval(timer))

// A filter change loads again from the newest trade and goes into the
// address (leaving out the defaults); typing a wallet waits for a pause.
watch([token, hours, size, wallet], ([, , , w], [, , , was]) => {
  router.replace({ query: { token: token.value || undefined, hours: hours.value !== 24 ? hours.value : undefined,
    size: size.value !== 'all' ? size.value : undefined, wallet: wallet.value || undefined } })
  clearTimeout(typing)
  typing = setTimeout(() => load(false, true), w !== was ? 400 : 0)
})

const changed = computed(() => !!(token.value || wallet.value || hours.value !== 24 || size.value !== 'all'))
// reset puts every filter back to its default.
function reset() {
  token.value = ''; hours.value = 24; size.value = 'all'; wallet.value = ''
}
// pickWallet shows one wallet's trades; the same wallet again shows all.
const pickWallet = (w) => (wallet.value = wallet.value === w ? '' : w)
const span = computed(() => ({ 1: 'hour', 4: '4 hours', 24: '24 hours', 168: '7 days', 720: '30 days' })[hours.value])
</script>

<template>
  <section aria-labelledby="trades-title">
    <h1 id="trades-title" class="sr-only">Large Trades</h1>
    <TradeFilters v-model:token="token" v-model:hours="hours" v-model:size="size" v-model:wallet="wallet" :tokens="tokens" :ranges="ranges"
                  :changed="changed" @reset="reset" />
    <p v-if="error" class="muted note">{{ error }}</p>
    <template v-else>
      <TradesTable :rows="trades" :wallet="wallet" :now="now" :show-size="size === 'all'" @wallet="pickWallet" />
      <p v-if="loaded && !trades.length" class="muted note">No trades match these filters in the last {{ span }}.</p>
      <button v-if="more" type="button" class="more" :disabled="loadingMore" @click="load(true)">{{ loadingMore ? 'Loading…' : 'Show more' }}</button>
    </template>
  </section>
</template>

<style scoped>
.sr-only { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0); white-space: nowrap; }
.note { margin: 0; padding: 24px 0; }
.more { display: block; margin: 16px auto 0; height: 36px; padding: 0 20px; border: 1px solid var(--line); border-radius: 6px; background: none; color: var(--fg); font: inherit; font-size: 14px; font-weight: 500; cursor: pointer; }
.more:hover:not(:disabled) { background: var(--hover); }
</style>
