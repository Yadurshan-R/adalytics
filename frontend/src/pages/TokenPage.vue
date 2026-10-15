<!-- Token page: the token panel on the left (laid out like Birdeye's, with
     CoinMarketCap's info list under it), the chart and recent trades on the
     right.
     - TokenOverview: name, links, policy ID and the main numbers in boxes.
     - ActivityBox: 5m, 1h, 4h and 24h volume and trades, buys against sells.
     - TokenInfo: Minswap button, traders/holders/security, price performance,
       links and IDs, supply, pool, description.
     On phones the chart comes right after the price boxes, like Birdeye's app. -->
<script setup>
import { ref, computed } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useTokens } from '../useTokens.js'
import TokenOverview from '../components/TokenOverview.vue'
import ActivityBox from '../components/ActivityBox.vue'
import TokenInfo from '../components/TokenInfo.vue'
import ChartPanel from '../components/ChartPanel.vue'
import RecentTrades from '../components/RecentTrades.vue'
import LoadingProgress from '../components/LoadingProgress.vue'

// The route's :ticker is the token's unique id from the backend (usually its ticker in lower case).
const props = defineProps({ ticker: { type: String, required: true } })
const { tokens, history, day, loading } = useTokens()
// token is the token this page is about, found by the id in the address.
const token = computed(() => tokens.value.find((t) => t.id === props.ticker.toLowerCase()))

// Back arrow: return to the list as it was (same view, period and search) when
// the visitor came from it, otherwise open the list.
const router = useRouter()
// back goes to the previous page if it was the list, otherwise opens the list.
function back() {
  if (window.history.state?.back) router.back()
  else router.push('/')
}

// span is the low and high of the last 24 hours, from the chart (ChartPanel).
const span = ref(null)
</script>

<template>
  <LoadingProgress v-if="history.loading" :loading="history.loading" />
  <p v-else-if="loading" class="muted">Loading…</p>
  <p v-else-if="!token" class="muted">No token called {{ ticker }}. <RouterLink to="/" class="link">See all tokens</RouterLink></p>

  <div v-else class="layout">
    <div class="top">
      <nav class="crumb">
        <button type="button" class="back" aria-label="Back to all tokens" @click="back">
          <svg viewBox="0 0 20 20" aria-hidden="true"><path d="M12.7 4.3 7 10l5.7 5.7-1.1 1.1L4.8 10l6.8-6.8z" /></svg>
        </button>
        <RouterLink to="/" class="link">Tokens</RouterLink>
        <span class="sep" aria-hidden="true">›</span>
        <span class="muted">{{ token.name }}</span>
      </nav>
      <TokenOverview :token="token" :day-ready="day.ready" />
    </div>
    <div class="rest">
      <ActivityBox :id="token.id" />
      <TokenInfo :token="token" :span="span" />
    </div>

    <section class="chart-col">
      <ChartPanel :id="token.id" :supply="token.total_supply || 0" @span="span = $event" />
      <RecentTrades :id="token.id" :ticker="token.ticker" />
    </section>
  </div>
</template>

<style scoped>
.layout { display: grid; grid-template-columns: 360px minmax(0, 1fr); grid-template-rows: auto 1fr; grid-template-areas: "top chart" "rest chart"; column-gap: 32px; align-items: start; }
.top { grid-area: top; }
.rest { grid-area: rest; }
.chart-col { grid-area: chart; min-width: 0; }
@media (max-width: 900px) {
  .layout { grid-template-columns: minmax(0, 1fr); grid-template-rows: none; grid-template-areas: "top" "chart" "rest"; row-gap: 20px; }
}
.link { color: var(--fg); }
.link:hover { text-decoration: underline; }
.crumb { display: flex; align-items: center; gap: 8px; font-size: 14px; margin-bottom: 18px; }
.back { display: inline-flex; align-items: center; justify-content: center; width: 30px; height: 30px; margin-left: -4px; border: 0; border-radius: 50%; background: var(--chip); color: var(--fg); cursor: pointer; }
.back:hover { background: #e3e8ef; }
.back:focus-visible { outline: 2px solid var(--fg); outline-offset: 2px; }
.back svg { width: 16px; height: 16px; fill: currentColor; }
.sep { color: var(--muted); }
</style>
