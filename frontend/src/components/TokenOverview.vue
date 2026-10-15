<!-- The top of the token page's left column, laid out like Birdeye's token
     panel: logo, ticker and name, links and the short policy ID with a copy
     button, then the main numbers in boxes (price in ADA and USD, liquidity,
     FDV and 24h volume). -->
<script setup>
import { ref } from 'vue'
import { formatPrice, fullPrice, formatUSD, formatCompact } from '../format.js'
import { signedPct } from '../listSort.js'

const props = defineProps({
  token: { type: Object, required: true },
  dayReady: { type: Boolean, default: true }, // the 24h change is complete
})
const copied = ref(false)
// copy puts the policy ID on the clipboard and shows a tick for 1.5 seconds.
function copy() {
  navigator.clipboard?.writeText(props.token.policy_id)
  copied.value = true
  setTimeout(() => (copied.value = false), 1500)
}
</script>

<template>
  <div class="overview">
    <div class="head">
      <img v-if="token.has_logo" :src="`/api/tokens/${token.id}/logo`" alt="" class="logo" width="48" height="48" />
      <span v-else class="logo blank" aria-hidden="true">{{ token.ticker.slice(0, 1) }}</span>
      <div>
        <h1>{{ token.ticker }}</h1>
        <p class="muted">{{ token.name }}</p>
      </div>
    </div>

    <div class="links">
      <a v-if="token.website" :href="token.website" target="_blank" rel="noopener nofollow" class="icon" title="Website">
        <svg viewBox="0 0 20 20" aria-hidden="true"><circle cx="10" cy="10" r="7.5" /><path d="M2.5 10h15M10 2.5c2.2 2.3 2.2 12.7 0 15M10 2.5c-2.2 2.3-2.2 12.7 0 15" /></svg>
      </a>
      <a :href="`https://cardanoscan.io/token/${token.policy_id}${token.asset_name}`" target="_blank" rel="noopener" class="icon" title="Cardanoscan">
        <svg viewBox="0 0 20 20" aria-hidden="true"><circle cx="8.5" cy="8.5" r="5.5" /><path d="M12.5 12.5 17 17" /></svg>
      </a>
      <button type="button" class="policy" :title="copied ? 'Copied' : 'Copy policy ID'" @click="copy">
        <span class="mono">{{ token.policy_id.slice(0, 6) }}…{{ token.policy_id.slice(-6) }}</span>
        <svg v-if="copied" viewBox="0 0 16 16" aria-hidden="true"><path d="M3 8.5l3 3 7-7" /></svg>
        <svg v-else viewBox="0 0 16 16" aria-hidden="true"><rect x="5.5" y="5.5" width="8" height="8" rx="1.5" /><path d="M10.5 3.5v-.5A1.5 1.5 0 0 0 9 1.5H4A1.5 1.5 0 0 0 2.5 3v5A1.5 1.5 0 0 0 4 9.5h.5" /></svg>
      </button>
    </div>

    <div class="boxes two">
      <div class="box">
        <span class="label">Price</span>
        <span class="value num" :title="fullPrice(token.price_ada)">{{ formatPrice(token.price_ada) }}</span>
        <span class="num chg" :class="[token.change_24h > 0 ? 'up' : token.change_24h < 0 ? 'down' : 'muted', { faded: !dayReady }]">{{ signedPct(token.change_24h) }} 24h</span>
      </div>
      <div class="box">
        <span class="label">Price USD</span>
        <span class="value num">{{ token.price_usd ? formatUSD(token.price_usd) : '—' }}</span>
        <span class="chg muted">via USDM</span>
      </div>
    </div>
    <div class="boxes three">
      <div class="box"><span class="label">Liquidity</span><span class="value num">{{ formatCompact(token.reserve_ada * 2) }}</span></div>
      <div class="box"><span class="label">FDV</span><span class="value num">{{ token.fdv_ada ? formatCompact(token.fdv_ada) : '—' }}</span></div>
      <div class="box"><span class="label">24h Vol</span><span class="value num" :class="{ faded: !dayReady }">{{ formatCompact(token.volume_24h_ada) }}</span></div>
    </div>
  </div>
</template>

<style scoped>
.head { display: flex; align-items: center; gap: 14px; }
.logo { width: 48px; height: 48px; border-radius: 50%; object-fit: cover; flex: none; }
.blank { display: inline-flex; align-items: center; justify-content: center; background: var(--chip); color: var(--muted); font-size: 20px; font-weight: 700; }
h1 { margin: 0; font-size: 22px; font-weight: 700; line-height: 1.2; }
.head p { margin: 2px 0 0; font-size: 14px; }
.links { display: flex; align-items: center; gap: 4px; margin: 14px 0 12px; }
.icon { display: inline-flex; padding: 5px; border-radius: 6px; color: var(--muted); }
.icon:hover { background: var(--chip); color: var(--fg); }
.icon svg { width: 18px; height: 18px; fill: none; stroke: currentColor; stroke-width: 1.4; stroke-linecap: round; }
.policy { display: inline-flex; align-items: center; gap: 6px; margin-left: auto; padding: 4px 6px; border: 0; border-radius: 6px; background: none; color: var(--muted); font: inherit; font-size: 13px; cursor: pointer; }
.policy:hover { background: var(--chip); color: var(--fg); }
.policy svg { width: 14px; height: 14px; fill: none; stroke: currentColor; stroke-width: 1.3; stroke-linecap: round; stroke-linejoin: round; }
.mono { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
.boxes { display: grid; gap: 8px; margin-top: 8px; }
.two { grid-template-columns: repeat(2, 1fr); }
.three { grid-template-columns: repeat(3, 1fr); }
.box { display: flex; flex-direction: column; align-items: center; gap: 4px; padding: 12px 6px; border: 1px solid var(--line); border-radius: 8px; text-align: center; min-width: 0; }
.label { font-size: 12px; color: var(--muted); }
.value { font-size: 17px; font-weight: 600; white-space: nowrap; }
.chg { font-size: 12px; font-weight: 500; }
.faded { opacity: 0.4; }
</style>
