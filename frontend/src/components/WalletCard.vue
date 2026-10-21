<!-- The wallet summary above Recent Trades, shown when a trader is picked:
     what the wallet did in this token over the last 24 hours (or 7 or 30
     days when opened from Top Traders, with the same totals as its row
     there), and which other tokens it traded in the last 24 hours (from the
     trades the backend already holds). -->
<script setup>
import { ref, computed, watch } from 'vue'
import { formatCompact, formatAgo } from '../format.js'

const props = defineProps({
  wallet: { type: String, required: true },
  id: { type: String, required: true }, // this token
  trader: { type: Object, default: null }, // its totals in this token, from the backend; null while loading
  hours: { type: Number, default: 24 }, // the time range of those totals
})
defineEmits(['close'])
const others = ref([])
const copied = ref(false)

// load asks the backend which other tokens this wallet traded.
async function load() {
  try {
    const res = await fetch(`/api/wallets/${encodeURIComponent(props.wallet)}`, { cache: 'no-store' })
    others.value = res.ok ? ((await res.json()).tokens ?? []).filter((t) => t.id !== props.id) : []
  } catch {
    others.value = []
  }
}
watch(() => props.wallet, load, { immediate: true })

// stats are the wallet's buys and sells in this token (zeros while loading).
const stats = computed(() => {
  const t = props.trader ?? {}
  return { buys: t.buys ?? 0, sells: t.sells ?? 0, buy: t.buy_ada ?? 0, sell: t.sell_ada ?? 0, net: (t.buy_ada ?? 0) - (t.sell_ada ?? 0) }
})
const span = computed(() => ({ 24: '24h', 168: '7D', 720: '30D' })[props.hours] ?? '24h')
// signed shows a net amount like +₳1.4K or −₳640.
const signed = (v) => (v > 0 ? '+' : v < 0 ? '−' : '') + formatCompact(Math.abs(v))
const tone = (v) => (v > 0 ? 'up' : v < 0 ? 'down' : '')

// copy puts the address on the clipboard and shows "Copied" for 1.5 seconds.
function copy() {
  navigator.clipboard?.writeText(props.wallet)
  copied.value = true
  setTimeout(() => (copied.value = false), 1500)
}
</script>

<template>
  <div class="card">
    <div class="head">
      <span class="label">Trader · {{ span }}</span>
      <span class="addr" :title="wallet">{{ wallet.slice(0, 16) }}…{{ wallet.slice(-6) }}</span>
      <button type="button" class="link" @click="copy">{{ copied ? 'Copied' : 'Copy' }}</button>
      <a class="link" :href="`https://cardanoscan.io/${wallet.startsWith('stake1') ? 'stakekey' : 'address'}/${wallet}`" target="_blank" rel="noopener">Cardanoscan ↗</a>
      <button type="button" class="link clear" @click="$emit('close')">Show all wallets</button>
    </div>
    <dl class="grid num" :class="{ wait: !trader }">
      <div><dt>Buys</dt><dd class="up">{{ stats.buys }}</dd></div>
      <div><dt>Sells</dt><dd class="down">{{ stats.sells }}</dd></div>
      <div><dt>Bought</dt><dd>{{ formatCompact(stats.buy) }}</dd></div>
      <div><dt>Sold</dt><dd>{{ formatCompact(stats.sell) }}</dd></div>
      <div><dt>Net</dt><dd :class="tone(stats.net)">{{ signed(stats.net) }}</dd></div>
    </dl>
    <p v-if="others.length" class="others">
      <span class="muted">Also traded in 24h:</span>
      <template v-for="(t, i) in others" :key="t.id">
        <RouterLink :to="`/token/${t.id}`" :title="`Last trade ${formatAgo(t.last)}`">{{ t.ticker }}</RouterLink>
        <span class="num" :class="tone(t.buy_ada - t.sell_ada)">{{ signed(t.buy_ada - t.sell_ada) }}</span><span v-if="i < others.length - 1" class="muted"> · </span>
      </template>
    </p>
  </div>
</template>

<style scoped>
.card { margin-bottom: 8px; padding: 12px 14px; border: 1px solid var(--line); border-radius: 8px; font-size: 13px; }
.head { display: flex; align-items: center; gap: 4px 12px; flex-wrap: wrap; }
.label { color: var(--muted); }
.addr { font: 13px ui-monospace, SFMono-Regular, Menlo, monospace; color: #2962ff; }
.link { padding: 0; border: 0; background: none; color: #6a6d78; font: inherit; font-weight: 500; text-decoration: none; cursor: pointer; }
.link:hover { color: #131722; text-decoration: underline; }
.clear { margin-left: auto; }
.grid { display: grid; grid-template-columns: repeat(5, 1fr); gap: 8px; margin: 10px 0 0; }
dt { font-size: 12px; color: var(--muted); }
dd { margin: 2px 0 0; font-size: 14px; font-weight: 600; }
.others { margin: 10px 0 0; line-height: 1.7; }
.others a { margin: 0 4px 0 6px; color: #131722; font-weight: 600; text-decoration: none; }
.others a:hover { text-decoration: underline; }
@media (max-width: 700px) { .grid { grid-template-columns: repeat(3, 1fr); } .clear { margin-left: 0; } }
.wait { opacity: 0.4; }
</style>
