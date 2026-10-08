<!-- Token page, CoinGecko-style: details on the left, live chart on the right. -->
<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useTokens } from '../useTokens.js'
import { formatPrice, fullPrice, formatUSD, formatCompact, formatAmount, formatAgo } from '../format.js'
import Change from '../components/Change.vue'
import PriceChart from '../components/PriceChart.vue'
import RecentTrades from '../components/RecentTrades.vue'
import { chartType, chartRange } from '../chartPrefs.js'
import LoadingProgress from '../components/LoadingProgress.vue'

// The route's :ticker is the token's unique id from the backend (usually its ticker in lower case).
const props = defineProps({ ticker: { type: String, required: true } })
const { tokens, history, loading, reload } = useTokens()
const token = computed(() => tokens.value.find((t) => t.id === props.ticker.toLowerCase()))

// Back arrow: return to the list as it was (same view, toggle and search) when
// the visitor came from it, otherwise open the list.
const router = useRouter()
function back() {
  if (window.history.state?.back) router.back()
  else router.push('/')
}

// Range buttons. Every range can be clicked; if its history is still loading,
// the chart area shows the live loading percentage until the chart is ready.
const ranges = computed(() => (history.value.ranges?.length ? history.value.ranges : [{ key: '24H', progress: 100, ready: true }]))
const selected = computed(() => ranges.value.find((r) => r.key === chartRange.value) ?? ranges.value[0])
const waiting = computed(() => !selected.value.ready)
const types = [{ key: 'line', label: 'Line' }, { key: 'candles', label: 'Candles' }]
const setType = (key) => (chartType.value = key)
const setRange = (key) => (chartRange.value = key)

// Chart data: refreshed every minute, like the list.
const points = ref([])
const candles = ref([])
const chartError = ref('')
let timer = null

async function loadChart() {
  if (waiting.value) {
    points.value = []
    candles.value = []
    return
  }
  const url = `/api/tokens/${encodeURIComponent(props.ticker)}/chart?range=${selected.value.key}&type=${chartType.value}`
  try {
    const res = await fetch(url, { cache: 'no-store' })
    const data = await res.json()
    if (res.status === 409) {
      // Not loaded after all (for example the backend restarted): show the progress.
      reload()
      return
    }
    if (!res.ok) throw new Error(data.error || `The backend answered ${res.status}`)
    points.value = data.points ?? []
    candles.value = data.candles ?? []
    chartError.value = ''
  } catch (e) {
    chartError.value = e instanceof TypeError ? 'Cannot reach the backend.' : e.message
  }
}

onMounted(() => {
  loadChart()
  timer = setInterval(loadChart, 60_000)
})
onUnmounted(() => {
  clearInterval(timer)
  clearInterval(progressTimer)
})
watch([() => props.ticker, chartRange, chartType], loadChart)

// While a loading range is selected, ask the backend for the progress every 3
// seconds (it only reads its own memory; no Koios request), and draw the chart
// as soon as the range is ready.
let progressTimer = null
watch(waiting, (now, before) => {
  clearInterval(progressTimer)
  if (now) progressTimer = setInterval(reload, 3000)
  else if (before) loadChart()
}, { immediate: true })

// Info box helpers.
const short = (s) => (s && s.length > 16 ? `${s.slice(0, 8)}…${s.slice(-6)}` : s)
const host = (url) => { try { return new URL(url).hostname.replace(/^www\./, '') } catch { return url } }
const createdDate = (iso) => new Date(iso).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })

const copied = ref('')
let copiedTimer = null
async function copy(text, key) {
  try {
    await navigator.clipboard.writeText(text)
    copied.value = key
    clearTimeout(copiedTimer)
    copiedTimer = setTimeout(() => (copied.value = ''), 1500)
  } catch {
    copied.value = ''
  }
}

// Low / high over the chart's range, from the line points or the candles.
const lows = computed(() => (chartType.value === 'candles' ? candles.value.map((c) => c.low) : points.value.map((p) => p.price)).filter((v) => v > 0))
const highs = computed(() => (chartType.value === 'candles' ? candles.value.map((c) => c.high) : points.value.map((p) => p.price)).filter((v) => v > 0))
const low = computed(() => (lows.value.length ? Math.min(...lows.value) : null))
const high = computed(() => (highs.value.length ? Math.max(...highs.value) : null))
const rangeLabel = computed(() => ({ '24H': '24h', '7D': '7d', '30D': '30d' })[chartRange.value] + ' Range')
const rangePos = computed(() => {
  if (!token.value || low.value == null || high.value === low.value) return 50
  return Math.min(100, Math.max(0, ((token.value.price_ada - low.value) / (high.value - low.value)) * 100))
})
</script>

<template>
  <LoadingProgress v-if="history.loading" :loading="history.loading" />
  <p v-else-if="loading" class="muted">Loading…</p>
  <p v-else-if="!token" class="muted">No token called {{ ticker }}. <RouterLink to="/" class="link">See all tokens</RouterLink></p>

  <div v-else class="layout">
    <aside class="info">
      <nav class="crumb">
        <button type="button" class="back" aria-label="Back to all tokens" @click="back">
          <svg viewBox="0 0 20 20" aria-hidden="true"><path d="M12.7 4.3 7 10l5.7 5.7-1.1 1.1L4.8 10l6.8-6.8z" /></svg>
        </button>
        <RouterLink to="/" class="link">Tokens</RouterLink>
        <span class="sep" aria-hidden="true">›</span>
        <span class="muted">{{ token.name }} Price</span>
      </nav>

      <h1 class="title">
        <img v-if="token.has_logo" :src="`/api/tokens/${token.id}/logo`" alt="" class="logo" width="32" height="32" />
        <span class="name">{{ token.name }}</span>
        <span class="muted sub">{{ token.ticker }} Price</span>
      </h1>

      <div class="price-row">
        <span class="price num" :title="fullPrice(token.price_ada)">{{ formatPrice(token.price_ada) }}</span>
        <span class="chg"><Change :value="token.change_24h" /> <span :class="token.change_24h >= 0 ? 'up' : 'down'">(24h)</span></span>
      </div>
      <p v-if="token.price_usd" class="usd muted num" :title="'$' + token.price_usd.toPrecision(6)">≈ {{ formatUSD(token.price_usd) }}</p>

      <div v-if="low != null && !waiting" class="range">
        <div class="bar"><span :style="{ width: rangePos + '%' }"></span></div>
        <div class="range-labels num">
          <span>{{ formatPrice(low) }}</span>
          <span class="label">{{ rangeLabel }}</span>
          <span>{{ formatPrice(high) }}</span>
        </div>
      </div>

      <dl class="stats">
        <div v-if="token.fdv_ada" class="stat"><dt>Fully Diluted Valuation</dt><dd class="num">{{ formatCompact(token.fdv_ada) }}</dd></div>
        <div class="stat"><dt>24 Hour Trading Vol</dt><dd class="num">{{ formatCompact(token.volume_24h_ada) }}</dd></div>
        <template v-if="token.trades_24h">
          <div class="stat"><dt>24h Trades</dt><dd class="num">{{ token.trades_24h.toLocaleString('en-US') }}</dd></div>
          <div class="stat">
            <dt>24h Buys / Sells</dt>
            <dd class="num"><span class="up">{{ token.buys_24h.toLocaleString('en-US') }}</span> / <span class="down">{{ token.sells_24h.toLocaleString('en-US') }}</span></dd>
          </div>
          <div class="stat">
            <dt>24h Buy / Sell Vol</dt>
            <dd class="num"><span class="up">{{ formatCompact(token.buy_volume_24h_ada) }}</span> / <span class="down">{{ formatCompact(token.sell_volume_24h_ada) }}</span></dd>
          </div>
        </template>
        <div v-if="token.total_supply" class="stat"><dt>Total Supply</dt><dd class="num">{{ formatAmount(token.total_supply) }} {{ token.ticker }}</dd></div>
        <div v-if="token.holders" class="stat"><dt>Holders</dt><dd class="num">{{ token.holders.toLocaleString('en-US') }}</dd></div>
        <div class="stat"><dt>Pool Liquidity</dt><dd class="num">{{ formatCompact(token.reserve_ada * 2) }}</dd></div>
        <div class="stat"><dt>Pool Reserves</dt><dd class="num">{{ formatCompact(token.reserve_ada) }} / {{ Math.round(token.reserve_token).toLocaleString('en-US') }} {{ token.ticker }}</dd></div>
        <div class="stat"><dt>Last Trade</dt><dd class="num" :title="token.last_swap ? new Date(token.last_swap).toLocaleString() : ''">{{ formatAgo(token.last_swap) }}</dd></div>
      </dl>

      <h2 class="info-title">Info</h2>
      <dl class="stats info-list">
        <div v-if="token.website" class="stat">
          <dt>Website</dt>
          <dd><a :href="token.website" target="_blank" rel="noopener nofollow" class="pill">{{ host(token.website) }}</a></dd>
        </div>
        <div class="stat">
          <dt>Policy ID</dt>
          <dd class="id">
            <code :title="token.policy_id">{{ short(token.policy_id) }}</code>
            <button type="button" class="copy" :aria-label="'Copy policy ID'" @click="copy(token.policy_id, 'policy')">{{ copied === 'policy' ? 'Copied' : 'Copy' }}</button>
          </dd>
        </div>
        <div v-if="token.fingerprint" class="stat">
          <dt>Fingerprint</dt>
          <dd class="id">
            <code :title="token.fingerprint">{{ short(token.fingerprint) }}</code>
            <button type="button" class="copy" :aria-label="'Copy fingerprint'" @click="copy(token.fingerprint, 'fingerprint')">{{ copied === 'fingerprint' ? 'Copied' : 'Copy' }}</button>
          </dd>
        </div>
        <div v-if="token.created" class="stat"><dt>Created</dt><dd>{{ createdDate(token.created) }}</dd></div>
        <div class="stat">
          <dt>Explorer</dt>
          <dd><a :href="`https://cardanoscan.io/token/${token.policy_id}${token.asset_name}`" target="_blank" rel="noopener" class="pill">Cardanoscan</a></dd>
        </div>
      </dl>
      <p v-if="token.description" class="about">{{ token.description }}</p>
    </aside>

    <section class="chart-col">
      <div class="toolbar">
        <div class="ranges" role="group" aria-label="Chart type">
          <button v-for="t in types" :key="t.key" type="button" :aria-pressed="t.key === chartType" @click="setType(t.key)">{{ t.label }}</button>
        </div>
        <div class="ranges" role="group" aria-label="Time range">
          <button v-for="r in ranges" :key="r.key" type="button" :aria-pressed="r.key === selected.key" @click="setRange(r.key)">{{ r.key }}</button>
        </div>
      </div>
      <p v-if="chartError" class="notice">{{ chartError }}</p>
      <div v-if="waiting" class="progress" role="progressbar" :aria-label="`${selected.key} loading`"
           :aria-valuenow="selected.progress" aria-valuemin="0" aria-valuemax="100">
        <span class="p-key">{{ selected.key }}</span>
        <span class="p-pct num">{{ selected.progress }}%</span>
        <span class="p-bar"><span :style="{ width: Math.max(selected.progress, 1) + '%' }"></span></span>
      </div>
      <PriceChart v-else :type="chartType" :points="points" :candles="candles" />
      <RecentTrades :id="token.id" :ticker="token.ticker" />
    </section>
  </div>
</template>

<style scoped>
.layout { display: grid; grid-template-columns: 340px minmax(0, 1fr); gap: 40px; }
@media (max-width: 900px) { .layout { grid-template-columns: minmax(0, 1fr); gap: 24px; } }

.link { color: var(--fg); }
.link:hover { text-decoration: underline; }
.crumb { display: flex; align-items: center; gap: 8px; font-size: 14px; margin-bottom: 16px; }
.back { display: inline-flex; align-items: center; justify-content: center; width: 32px; height: 32px; margin-left: -6px; border: 0; border-radius: 50%; background: var(--chip); color: var(--fg); cursor: pointer; }
.back:hover { background: #e3e8ef; }
.back:focus-visible { outline: 2px solid var(--fg); outline-offset: 2px; }
.back svg { width: 18px; height: 18px; fill: currentColor; }
.sep { color: var(--muted); }

.title { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; margin: 0; font-size: 22px; }
.name { font-weight: 700; }
.logo { width: 32px; height: 32px; border-radius: 50%; object-fit: cover; }
.sub { font-size: 16px; font-weight: 400; }

.price-row { display: flex; align-items: baseline; gap: 10px; flex-wrap: wrap; margin-top: 10px; }
.price { font-size: 36px; font-weight: 700; letter-spacing: -0.01em; }
.chg { display: inline-flex; align-items: baseline; gap: 4px; font-size: 17px; font-weight: 600; }

.range { margin-top: 22px; }
.bar { height: 8px; border-radius: 4px; background: var(--chip); overflow: hidden; }
.bar span { display: block; height: 100%; border-radius: 4px; background: linear-gradient(90deg, #f6d32d, var(--bull)); }
.range-labels { display: flex; justify-content: space-between; margin-top: 8px; font-size: 14px; font-weight: 600; }
.range-labels .label { font-weight: 500; }

.stats { margin: 20px 0 0; }
.stat { display: flex; justify-content: space-between; gap: 12px; padding-block: 14px; border-bottom: 1px solid var(--line); font-size: 14px; }
.stat dt { color: var(--muted); }
.stat dd { margin: 0; font-weight: 600; text-align: right; }

.toolbar { display: flex; justify-content: space-between; gap: 12px; flex-wrap: wrap; margin-bottom: 12px; }
.ranges { display: inline-flex; background: var(--chip); border-radius: 10px; padding: 4px; gap: 2px; }
.ranges button { font: inherit; font-size: 14px; font-weight: 600; color: var(--muted); background: none; border: 0; padding: 6px 12px; border-radius: 8px; cursor: pointer; }
.ranges button[aria-pressed="true"] { background: var(--bg); color: var(--fg); box-shadow: 0 1px 3px rgba(13, 20, 33, 0.12); }
.ranges button:focus-visible { outline: 2px solid var(--fg); outline-offset: 1px; }
.progress { height: 440px; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 10px; border-radius: 12px; background: #fafbfd; }
@media (max-width: 700px) { .progress { height: 320px; } }
.p-key { font-size: 15px; font-weight: 600; color: var(--muted); }
.p-pct { font-size: 40px; font-weight: 700; letter-spacing: -0.02em; }
.p-bar { width: min(320px, 70%); height: 8px; border-radius: 4px; background: var(--chip); overflow: hidden; }
.p-bar span { display: block; height: 100%; border-radius: 4px; background: var(--bull); transition: width 0.6s ease; }
@media (prefers-reduced-motion: reduce) { .p-bar span { transition: none; } }

.usd { margin: 4px 0 0; font-size: 15px; }

.info-title { font-size: 18px; font-weight: 700; margin: 32px 0 4px; }
.info-list { margin-top: 0; }
.stat dd a { color: var(--fg); }
.pill { display: inline-block; padding: 3px 10px; border-radius: 8px; background: var(--chip); font-weight: 600; font-size: 13px; }
.pill:hover { background: #e3e8ef; }
.id { display: inline-flex; align-items: center; gap: 8px; }
.id code { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 13px; font-weight: 500; }
.copy { font: inherit; font-size: 12px; font-weight: 600; color: var(--muted); background: var(--chip); border: 0; border-radius: 6px; padding: 3px 8px; cursor: pointer; min-width: 56px; }
.copy:hover { color: var(--fg); }
.copy:focus-visible, .pill:focus-visible { outline: 2px solid var(--fg); outline-offset: 1px; }
.about { margin: 16px 0 0; color: var(--muted); line-height: 1.6; font-size: 14px; overflow-wrap: anywhere; }

.notice { margin: 0 0 12px; padding: 10px 14px; border-radius: 10px; background: #fdecee; color: #a0202c; }
</style>
