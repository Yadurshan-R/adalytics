<!-- The table on the Large Trades page, like Birdeye's Find Trades: time,
     buy or sell, the ADA value, both sides of the swap (the token and the
     ADA, signed from the trader's side), the share of the pool, the trader
     and a Cardanoscan link. A click on the token opens its page; a click on
     the trader shows only that wallet's trades (the page owns the filter). -->
<script setup>
import { useRouter } from 'vue-router'
import { formatAmount, formatAgo } from '../format.js'
import TraderCell from './TraderCell.vue'
import SizeTag from './SizeTag.vue'

defineProps({
  rows: { type: Array, default: () => [] },
  wallet: { type: String, default: '' }, // the wallet filtered on, or ''
  now: { type: Number, default: () => Date.now() },
  showSize: { type: Boolean, default: true }, // the BIG / WHALE tag; off when one size is picked
})
const emit = defineEmits(['wallet'])
const router = useRouter()

const open = (t) => router.push({ name: 'token', params: { ticker: t.id } })
const ada = (v) => '₳' + Math.round(v).toLocaleString('en-US')
// sign: a buy gets tokens (+) for ADA (−); a sell the other way round.
// rowKey is unique for every trade: one wallet can have two orders in one
// transaction (fill tells them apart), and one transaction can swap in two pools.
const rowKey = (t) => `${t.tx_hash}#${t.fill}:${t.id}`
const sign = (t, part) => ((t.side === 'buy') === (part === 'token') ? '+' : '−')
</script>

<template>
  <div class="wrap">
    <table role="table">
      <thead role="rowgroup">
        <tr role="row">
          <th role="columnheader" class="left time">Time</th>
          <th role="columnheader" class="left type">Type</th>
          <th role="columnheader" class="left hide-sm value-col">Value</th>
          <th role="columnheader" class="left amount-col">Amount</th>
          <th role="columnheader" class="left hide-sm pool-col" title="The trade's ADA against the pool's ADA">% of pool</th>
          <th role="columnheader" class="left trader-col">Trader</th>
          <th role="columnheader" class="link-col hide-sm"><span class="sr-only">Transaction</span></th>
        </tr>
      </thead>
      <tbody role="rowgroup">
        <tr role="row" v-for="t in rows" :key="rowKey(t)">
          <td role="cell" class="left muted" :title="new Date(t.time).toLocaleString()">{{ formatAgo(t.time, now) }}</td>
          <td role="cell" class="left"><span class="side" :class="t.side">{{ t.side === 'buy' ? 'Buy' : 'Sell' }}</span><span v-if="showSize" class="show-sm"><br /><SizeTag :ada="t.amount_ada" class="phone-tag" /></span></td>
          <td role="cell" class="left hide-sm num value value-col">
            {{ ada(t.amount_ada) }}<SizeTag v-if="showSize" :ada="t.amount_ada" />
          </td>
          <td role="cell" class="left amount-col">
            <button type="button" class="line token" :title="`Open ${t.ticker}`" @click="open(t)">
              <img v-if="t.has_logo" :src="`/api/tokens/${t.id}/logo`" alt="" class="logo" width="20" height="20" loading="lazy" />
              <span v-else class="logo blank" aria-hidden="true">{{ t.ticker.slice(0, 1) }}</span>
              <span class="num">{{ sign(t, 'token') }}{{ formatAmount(t.amount_token) }}</span><b>{{ t.ticker }}</b>
            </button>
            <span class="line">
              <span class="logo ada" aria-hidden="true">₳</span>
              <span class="num">{{ sign(t, 'ada') }}{{ Math.round(t.amount_ada).toLocaleString('en-US') }}</span><b>ADA</b>
            </span>
          </td>
          <td role="cell" class="left hide-sm num">{{ t.pool_share.toFixed(1) }}%</td>
          <td role="cell" class="left"><TraderCell :wallet="t.wallet" :on="!!wallet && t.wallet === wallet" @pick="emit('wallet', $event)" /></td>
          <td role="cell" class="link-col hide-sm">
            <a :href="`https://cardanoscan.io/transaction/${t.tx_hash}`" target="_blank" rel="noopener" class="tx"
               :aria-label="`View this ${t.ticker} ${t.side} on Cardanoscan`" title="View on Cardanoscan">
              <svg viewBox="0 0 16 16" aria-hidden="true"><path d="M9 2h5v5h-1.5V4.6L7.1 10 6 8.9l5.4-5.4H9V2ZM3 4h4v1.5H4.5v6h6V9H12v4H3V4Z" /></svg>
            </a>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.wrap { position: relative; overflow-x: auto; }
table { width: 100%; border-collapse: collapse; }
th, td { padding: 12px 10px; text-align: right; white-space: nowrap; border-bottom: 1px solid var(--line); font-size: 14px; }
th { font-size: 12px; font-weight: 500; color: var(--muted); }
.left { text-align: left; }
.side { font-size: 13px; font-weight: 600; }
.side.buy { color: var(--bull); }
.side.sell { color: var(--bear); }
.value { font-weight: 600; }
/* Value and Amount each start at their own left edge with room between,
   like Birdeye, so the two numbers do not read as one. */
/* On a computer the table is laid out as a grid: each column gets what it
   holds plus an equal share of the spare room, so the gaps between all
   columns are the same and nothing piles up at one end. The roles keep it
   a table for screen readers. */
@media (min-width: 1000px) {
  table { display: grid; grid-template-columns: repeat(6, auto) 44px; }
  thead, tbody, tr { display: contents; }
  th, td { display: flex; align-items: center; }
  td.amount-col { flex-direction: column; align-items: flex-start; justify-content: center; }
}
.line { display: flex; align-items: center; gap: 6px; padding: 0; border: 0; background: none; color: var(--fg); font: inherit; }
.line + .line { margin-top: 6px; }
.line b { font-weight: 500; color: #2962ff; }
.token { cursor: pointer; }
.token:hover b { text-decoration: underline; }
.logo { width: 20px; height: 20px; border-radius: 50%; object-fit: cover; flex: none; }
.blank, .ada { display: inline-flex; align-items: center; justify-content: center; background: var(--chip); color: var(--muted); font-size: 11px; font-weight: 700; }
.ada { background: #0033ad; color: #fff; }
.show-sm { display: none; }
.link-col { width: 36px; padding-inline: 4px; }
.tx { display: inline-flex; padding: 4px; border-radius: 6px; color: var(--muted); }
.tx:hover { color: var(--fg); background: var(--chip); }
.tx svg { width: 14px; height: 14px; fill: currentColor; }
.sr-only { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0); white-space: nowrap; }
@media (max-width: 700px) {
  .hide-sm { display: none; }
  .show-sm { display: inline; }
  .phone-tag { margin: 4px 0 0; }
  th, td { padding: 11px 3px; font-size: 13px; }
  .line { gap: 4px; }
}
</style>
