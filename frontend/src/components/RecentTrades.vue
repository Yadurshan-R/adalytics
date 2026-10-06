<!-- A token's 5 most recent trades (last 24 hours, newest first), refreshed every
     minute. Each row links to the transaction on Cardanoscan. -->
<script setup>
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { formatPrice, fullPrice, formatAmount, formatAgo } from '../format.js'

const props = defineProps({
  id: { type: String, required: true },
  ticker: { type: String, required: true },
})

const SHOWN = 5
const trades = ref([])
const loaded = ref(false)
const error = ref('')
const now = ref(Date.now())
let timer = null

async function load() {
  try {
    const res = await fetch(`/api/tokens/${encodeURIComponent(props.id)}/trades?limit=${SHOWN}`, { cache: 'no-store' })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error || `The backend answered ${res.status}`)
    trades.value = data.trades ?? []
    error.value = ''
  } catch (e) {
    error.value = e instanceof TypeError ? 'Cannot reach the backend.' : e.message
  } finally {
    loaded.value = true
    now.value = Date.now()
  }
}

onMounted(() => {
  load()
  timer = setInterval(load, 60_000)
})
onUnmounted(() => clearInterval(timer))
watch(() => props.id, () => {
  loaded.value = false
  trades.value = []
  load()
})

// ₳1,250 for normal amounts, ₳4.35 for small ones.
function ada(v) {
  return '₳' + v.toLocaleString('en-US', { maximumFractionDigits: v < 100 ? 2 : 0 })
}
</script>

<template>
  <section class="trades" aria-labelledby="trades-title">
    <h2 id="trades-title">Recent Trades</h2>
    <p v-if="error" class="muted note">{{ error }}</p>
    <p v-else-if="loaded && !trades.length" class="muted note">No trades in the last 24 hours.</p>

    <div v-else class="wrap">
      <table>
        <thead>
          <tr>
            <th class="left">Time</th>
            <th class="left">Type</th>
            <th class="hide-sm">Price</th>
            <th>₳ Amount</th>
            <th class="hide-sm">{{ ticker }} Amount</th>
            <th class="link-col"><span class="sr-only">Transaction</span></th>
          </tr>
        </thead>
        <tbody>
          <template v-if="!loaded">
            <tr v-for="n in SHOWN" :key="n" class="skeleton">
              <td class="left"><i style="width: 70px" /></td>
              <td class="left"><i style="width: 36px" /></td>
              <td class="hide-sm"><i style="width: 70px" /></td>
              <td><i style="width: 60px" /></td>
              <td class="hide-sm"><i style="width: 90px" /></td>
              <td class="link-col"></td>
            </tr>
          </template>
          <template v-else>
            <tr v-for="t in trades" :key="t.tx_hash + t.time">
              <td class="left muted" :title="new Date(t.time).toLocaleString()">{{ formatAgo(t.time, now) }}</td>
              <td class="left side" :class="t.side === 'buy' ? 'up' : 'down'">{{ t.side === 'buy' ? 'Buy' : 'Sell' }}</td>
              <td class="hide-sm num" :title="fullPrice(t.price_ada)">{{ formatPrice(t.price_ada) }}</td>
              <td class="num">{{ ada(t.amount_ada) }}</td>
              <td class="hide-sm num">{{ formatAmount(t.amount_token) }} {{ ticker }}</td>
              <td class="link-col">
                <a :href="`https://cardanoscan.io/transaction/${t.tx_hash}`" target="_blank" rel="noopener"
                   class="tx" :aria-label="`View this ${t.side} on Cardanoscan`" title="View on Cardanoscan">
                  <svg viewBox="0 0 16 16" aria-hidden="true"><path d="M9 2h5v5h-1.5V4.6L7.1 10 6 8.9l5.4-5.4H9V2ZM3 4h4v1.5H4.5v6h6V9H12v4H3V4Z" /></svg>
                </a>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>
  </section>
</template>

<style scoped>
.trades { margin-top: 32px; }
h2 { font-size: 18px; font-weight: 700; margin: 0 0 12px; }
.note { margin: 0; padding: 24px 0; border-top: 1px solid var(--line); }

.wrap { overflow-x: auto; }
table { width: 100%; border-collapse: collapse; }
th, td { padding: 12px 10px; text-align: right; white-space: nowrap; border-bottom: 1px solid var(--line); font-size: 14px; }
th { font-size: 13px; font-weight: 700; border-top: 1px solid var(--line); }
.left { text-align: left; }
.side { font-weight: 600; }

.link-col { width: 36px; padding-inline: 4px; }
.tx { display: inline-flex; padding: 4px; border-radius: 6px; color: var(--muted); }
.tx:hover { color: var(--fg); background: var(--chip); }
.tx:focus-visible { outline: 2px solid var(--fg); outline-offset: 1px; }
.tx svg { width: 14px; height: 14px; fill: currentColor; }
.sr-only { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0); white-space: nowrap; }

.skeleton i { display: inline-block; height: 14px; border-radius: 6px; background: linear-gradient(90deg, var(--line), #f7f9fb, var(--line)); background-size: 200% 100%; animation: shimmer 1.4s infinite; }
@keyframes shimmer { to { background-position: -200% 0; } }
@media (prefers-reduced-motion: reduce) { .skeleton i { animation: none; } }

@media (max-width: 700px) {
  .hide-sm { display: none; }
  th, td { padding: 12px 6px; }
}
</style>
