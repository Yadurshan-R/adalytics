<!-- Token page, CoinGecko-style: details on the left, live chart on the right. -->
<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { RouterLink } from 'vue-router'
import { useTokens } from '../useTokens.js'
import { formatPrice, fullPrice, formatUSD, formatCompact, formatAmount, formatAgo } from '../format.js'
import Change from '../components/Change.vue'
import PriceChart from '../components/PriceChart.vue'
import RecentTrades from '../components/RecentTrades.vue'

// The route's :ticker is the token's unique id from the backend (usually its ticker in lower case).
const props = defineProps({ ticker: { type: String, required: true } })
const { tokens, loading } = useTokens()
const token = computed(() => tokens.value.find((t) => t.id === props.ticker.toLowerCase()))

// Chart data: refreshed every minute, like the list.
const range = ref('24H')
const ranges = ['24H'] // more appear here as the backend gets longer history
const points = ref([])
const chartError = ref('')
let timer = null

async function loadChart() {
  try {
    const res = await fetch(`/api/tokens/${encodeURIComponent(props.ticker)}/chart?range=${range.value}`, { cache: 'no-store' })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error || `The backend answered ${res.status}`)
    points.value = data.points ?? []
    chartError.value = ''
  } catch (e) {
    chartError.value = e instanceof TypeError ? 'Cannot reach the backend.' : e.message
  }
}

onMounted(() => {
  loadChart()
  timer = setInterval(loadChart, 60_000)
})
onUnmounted(() => clearInterval(timer))
watch(() => props.ticker, loadChart)

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

// 24h low / high from the chart points.
const low = computed(() => (points.value.length ? Math.min(...points.value.map((p) => p.price)) : null))
const high = computed(() => (points.value.length ? Math.max(...points.value.map((p) => p.price)) : null))
const rangePos = computed(() => {
  if (!token.value || low.value == null || high.value === low.value) return 50
  return Math.min(100, Math.max(0, ((token.value.price_ada - low.value) / (high.value - low.value)) * 100))
})
</script>

<template>
  <p v-if="loading" class="muted">Loading…</p>
  <p v-else-if="!token" class="muted">No token called {{ ticker }}. <RouterLink to="/" class="link">See all tokens</RouterLink></p>

  <div v-else class="layout">
    <aside class="info">
      <nav class="crumb">
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

      <div v-if="low != null" class="range">
        <div class="bar"><span :style="{ width: rangePos + '%' }"></span></div>
        <div class="range-labels num">
          <span>{{ formatPrice(low) }}</span>
          <span class="label">24h Range</span>
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
        <div class="ranges" role="group" aria-label="Time range">
          <button v-for="r in ranges" :key="r" type="button" :aria-pressed="r === range" @click="range = r; loadChart()">{{ r }}</button>
        </div>
      </div>
      <p v-if="chartError" class="notice">{{ chartError }}</p>
      <PriceChart :points="points" />
      <RecentTrades :id="token.id" :ticker="token.ticker" />
    </section>
  </div>
</template>

<style scoped>
.layout { display: grid; grid-template-columns: 340px minmax(0, 1fr); gap: 40px; }
@media (max-width: 900px) { .layout { grid-template-columns: minmax(0, 1fr); gap: 24px; } }

.link { color: var(--fg); }
.link:hover { text-decoration: underline; }
.crumb { display: flex; gap: 8px; font-size: 14px; margin-bottom: 16px; }
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

.toolbar { display: flex; margin-bottom: 12px; }
.ranges { display: inline-flex; background: var(--chip); border-radius: 10px; padding: 4px; gap: 2px; }
.ranges button { font: inherit; font-size: 14px; font-weight: 600; color: var(--muted); background: none; border: 0; padding: 6px 12px; border-radius: 8px; cursor: pointer; }
.ranges button[aria-pressed="true"] { background: var(--bg); color: var(--fg); box-shadow: 0 1px 3px rgba(13, 20, 33, 0.12); }
.ranges button:focus-visible { outline: 2px solid var(--fg); outline-offset: 1px; }

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
