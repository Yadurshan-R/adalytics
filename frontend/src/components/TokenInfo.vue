<!-- The lower part of the token page's left column:
     - a button that opens the token on Minswap (we never trade or hold funds);
     - Token Info like Birdeye's: 24h traders, holders and security, which on
       Cardano means whether the minting policy is locked (no more tokens can
       ever be made) or still open;
     - price performance (24h low and high), then links, IDs, supply, pool and
       last trade as a plain list like CoinMarketCap's, and the description. -->
<script setup>
import { ref, computed } from 'vue'
import { formatPrice, formatCompact, formatAmount, formatAgo, short, host, createdDate } from '../format.js'

const props = defineProps({
  token: { type: Object, required: true },
  span: { type: Object, default: null }, // { low, high } of the last 24 hours, from the chart
})

// pos places the price now between the 24h low and high, 0 to 100.
const pos = computed(() => {
  const r = props.span
  if (!r || r.high === r.low) return 50
  return Math.min(100, Math.max(0, ((props.token.price_ada - r.low) / (r.high - r.low)) * 100))
})

// SECURITY explains each policy state for the Security box.
const SECURITY = {
  locked: { text: 'Locked', tone: 'up', tip: 'The minting policy is locked: no more tokens can ever be made.' },
  open: { text: 'Open', tone: 'warn', tip: 'The policy can still mint: the creator can make more tokens.' },
  script: { text: 'Smart contract', tone: 'muted', tip: 'Minting is controlled by a smart contract, so it cannot be read as locked or open.' },
  unknown: { text: 'Unknown', tone: 'muted', tip: 'Koios has no script for this policy.' },
}
const security = computed(() => SECURITY[props.token.policy_state] ?? { text: '…', tone: 'muted', tip: 'Checking the minting policy' })
// minswap opens the token's swap page on Minswap.
const minswap = computed(() => `https://app.minswap.org/swap?currencySymbolA=&tokenNameA=&currencySymbolB=${props.token.policy_id}&tokenNameB=${props.token.asset_name}`)

const copied = ref('')
// copy puts an ID on the clipboard and shows "Copied" for 1.5 seconds.
function copy(text, key) {
  navigator.clipboard?.writeText(text)
  copied.value = key
  setTimeout(() => (copied.value = ''), 1500)
}
</script>

<template>
  <div class="info">
    <a :href="minswap" target="_blank" rel="noopener" class="trade">Trade {{ token.ticker }} on Minswap ↗</a>

    <h2>Token Info</h2>
    <div class="cards">
      <div class="card"><b class="num">{{ token.traders_24h != null ? token.traders_24h.toLocaleString('en-US') : '…' }}</b><span>24h Traders</span></div>
      <div class="card"><b class="num">{{ token.holders ? formatAmount(token.holders) : '…' }}</b><span>Holders</span></div>
      <div class="card" :title="security.tip">
        <b :class="security.tone">
          <svg viewBox="0 0 16 16" aria-hidden="true"><path d="M8 1.5 2.5 3.5v4c0 3.3 2.3 5.9 5.5 7 3.2-1.1 5.5-3.7 5.5-7v-4z" /></svg>{{ security.text }}
        </b>
        <span>Policy</span>
      </div>
    </div>

    <h2>Price performance <span class="muted">24h</span></h2>
    <div v-if="span" class="perf num">
      <div class="ends"><span><i>Low</i>{{ formatPrice(span.low) }}</span><span class="right"><i>High</i>{{ formatPrice(span.high) }}</span></div>
      <div class="track"><span :style="{ left: pos + '%' }"></span></div>
    </div>

    <h2>Details</h2>
    <dl>
      <div v-if="token.website"><dt>Website</dt><dd><a :href="token.website" target="_blank" rel="noopener nofollow" class="chip">{{ host(token.website) }}</a></dd></div>
      <div><dt>Explorer</dt><dd><a :href="`https://cardanoscan.io/token/${token.policy_id}${token.asset_name}`" target="_blank" rel="noopener" class="chip">Cardanoscan</a></dd></div>
      <div>
        <dt>Policy ID</dt>
        <dd><button type="button" class="chip mono" @click="copy(token.policy_id, 'policy')">{{ copied === 'policy' ? 'Copied' : short(token.policy_id) }}</button></dd>
      </div>
      <div v-if="token.fingerprint">
        <dt>Fingerprint</dt>
        <dd><button type="button" class="chip mono" @click="copy(token.fingerprint, 'fp')">{{ copied === 'fp' ? 'Copied' : short(token.fingerprint) }}</button></dd>
      </div>
      <div v-if="token.total_supply"><dt>Total supply</dt><dd class="num">{{ formatAmount(token.total_supply) }} {{ token.ticker }}</dd></div>
      <div><dt>Pool reserves</dt><dd class="num">{{ formatCompact(token.reserve_ada) }} / {{ formatAmount(token.reserve_token) }} {{ token.ticker }}</dd></div>
      <div><dt>Last trade</dt><dd class="num" :title="token.last_swap ? new Date(token.last_swap).toLocaleString() : ''">{{ formatAgo(token.last_swap) }}</dd></div>
      <div v-if="token.created"><dt>Created</dt><dd>{{ createdDate(token.created) }}</dd></div>
    </dl>

    <template v-if="token.description">
      <h2>About {{ token.name }}</h2>
      <p class="about">{{ token.description }}</p>
    </template>
  </div>
</template>

<style scoped>
.info { margin-top: 12px; }
.trade { display: block; padding: 10px; border-radius: 8px; background: #2962ff; color: #fff; font-size: 14px; font-weight: 600; text-align: center; text-decoration: none; }
.trade:hover { background: #1e53e5; }
.cards { display: grid; grid-template-columns: repeat(3, 1fr); gap: 8px; }
.card { display: flex; flex-direction: column; align-items: center; gap: 4px; padding: 12px 4px; border: 1px solid var(--line); border-radius: 8px; }
.card b { display: inline-flex; align-items: center; gap: 4px; font-size: 16px; font-weight: 600; }
.card span { font-size: 12px; color: var(--muted); }
.card svg { width: 14px; height: 14px; fill: none; stroke: currentColor; stroke-width: 1.4; stroke-linejoin: round; }
.warn { color: #d97706; }
h2 { margin: 22px 0 10px; font-size: 15px; font-weight: 700; }
h2:first-child { margin-top: 0; }
h2 .muted { font-weight: 500; font-size: 13px; margin-left: 4px; }
.ends { display: flex; justify-content: space-between; font-size: 14px; font-weight: 600; }
.ends span { display: flex; flex-direction: column; gap: 2px; }
.ends i { font-style: normal; font-size: 12px; font-weight: 500; color: var(--muted); }
.right { align-items: flex-end; }
.track { position: relative; height: 4px; margin-top: 10px; border-radius: 2px; background: var(--chip); }
.track span { position: absolute; top: -4px; width: 3px; height: 12px; margin-left: -1.5px; border-radius: 2px; background: var(--muted); }
dl { margin: 0; }
dl div { display: flex; justify-content: space-between; align-items: center; gap: 12px; padding: 9px 0; font-size: 14px; }
dt { color: var(--muted); }
dd { margin: 0; font-weight: 500; text-align: right; }
.chip { display: inline-block; padding: 3px 10px; border: 0; border-radius: 6px; background: var(--chip); color: var(--fg); font: inherit; font-size: 13px; font-weight: 500; text-decoration: none; cursor: pointer; }
.chip:hover { background: #e3e8ef; }
.mono { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 12px; }
.about { margin: 0; color: var(--muted); line-height: 1.6; font-size: 14px; overflow-wrap: anywhere; }
</style>
