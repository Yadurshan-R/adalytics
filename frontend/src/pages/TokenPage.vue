<!-- Token page, CoinGecko-style: details on the left, live chart on the right. -->
<script setup>
import { ref, computed } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useTokens } from '../useTokens.js'
import { formatPrice, fullPrice, formatUSD, formatCompact, formatAmount, formatAgo, short, host, createdDate } from '../format.js'
import Change from '../components/Change.vue'
import ChartPanel from '../components/ChartPanel.vue'
import RecentTrades from '../components/RecentTrades.vue'
import ActivityBox from '../components/ActivityBox.vue'
import LoadingProgress from '../components/LoadingProgress.vue'

// The route's :ticker is the token's unique id from the backend (usually its ticker in lower case).
const props = defineProps({ ticker: { type: String, required: true } })
const { tokens, history, day, loading } = useTokens()
// token is the token this page is about, found by the id in the address.
const token = computed(() => tokens.value.find((t) => t.id === props.ticker.toLowerCase()))

// Back arrow: return to the list as it was (same view, toggle and search) when
// the visitor came from it, otherwise open the list.
const router = useRouter()
// back goes to the previous page if it was the list, otherwise opens the list.
function back() {
  if (window.history.state?.back) router.back()
  else router.push('/')
}

const copied = ref('')
let copiedTimer = null
// copy puts text (a policy ID or fingerprint) on the clipboard and shows
// "Copied" next to it for 1.5 seconds.
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

// span is the low and high for the range bar, from the chart (ChartPanel).
const span = ref(null)
// rangePos places the current price on the low/high bar, from 0% to 100%.
const rangePos = computed(() => {
  const r = span.value
  if (!token.value || !r || r.high === r.low) return 50
  return Math.min(100, Math.max(0, ((token.value.price_ada - r.low) / (r.high - r.low)) * 100))
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
        <span class="chg" :class="{ partial: !day.ready }"><Change :value="token.change_24h" /> <span :class="token.change_24h >= 0 ? 'up' : 'down'">(24h)</span></span>
      </div>
      <p v-if="token.price_usd" class="usd muted num" :title="'$' + token.price_usd.toPrecision(6)">≈ {{ formatUSD(token.price_usd) }}</p>

      <div v-if="span" class="range">
        <div class="bar"><span :style="{ width: rangePos + '%' }"></span></div>
        <div class="range-labels num">
          <span>{{ formatPrice(span.low) }}</span>
          <span class="label">{{ span.label }}</span>
          <span>{{ formatPrice(span.high) }}</span>
        </div>
      </div>

      <ActivityBox :id="token.id" />

      <dl class="stats">
        <div v-if="token.fdv_ada" class="stat"><dt>Fully Diluted Valuation</dt><dd class="num">{{ formatCompact(token.fdv_ada) }}</dd></div>
        <div class="stat"><dt>24 Hour Trading Vol</dt><dd class="num" :class="{ partial: !day.ready }">{{ formatCompact(token.volume_24h_ada) }}</dd></div>
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
      <ChartPanel :id="token.id" :ticker="token.ticker" :supply="token.total_supply || 0" @span="span = $event" />
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

.usd { margin: 4px 0 0; font-size: 15px; }
.partial { opacity: 0.4; } /* the last 24 hours are still loading */

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

</style>
