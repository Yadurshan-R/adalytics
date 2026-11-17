<!-- The lower part of the token page's left column:
     - a button that opens the token on Minswap (we never trade or hold funds);
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
.trade { display: block; padding: 10px; border-radius: 8px; background: var(--accent); color: #fff; font-size: 14px; font-weight: 600; text-align: center; text-decoration: none; }
.trade:hover { background: var(--accent-hover); }
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
.chip:hover { background: var(--chip-hover); }
.mono { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 12px; }
.about { margin: 0; color: var(--muted); line-height: 1.6; font-size: 14px; overflow-wrap: anywhere; }
</style>
