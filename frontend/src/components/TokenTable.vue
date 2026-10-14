<!-- The token table: rank, token, price, 1h / 24h / 7d change, volume, FDV and
     a small price line. Shows grey placeholder rows while loading, and a short
     message when nothing matches. A click on a row opens the token's page. -->
<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { formatPrice, fullPrice, formatADA, formatCompact } from '../format.js'
import Change from './Change.vue'
import Sparkline from './Sparkline.vue'

const props = defineProps({
  rows: { type: Array, required: true }, // the tokens to show, in order
  ranks: { type: Object, default: null }, // token id → place in Top Volume; null numbers the rows 1, 2, 3
  loading: { type: Boolean, default: false },
  week: { type: Boolean, default: false }, // the list shows 7 day numbers
  weekReady: { type: Boolean, default: false }, // 7 days are loaded, so the 7d column shows
  emptyText: { type: String, default: '' },
})

const router = useRouter()
// open goes to a token's page.
const open = (t) => router.push({ name: 'token', params: { ticker: t.id } })
// lineOf is a token's small price line for the chosen period.
const lineOf = (t) => (props.week ? t.sparkline_7d : t.sparkline) || []
// lineLabel is the title of the small price line column.
const lineLabel = computed(() => (props.week ? 'Last 7 Days' : 'Last 24h'))
</script>

<template>
  <div class="table-wrap">
    <table>
      <thead>
        <tr>
          <th class="rank">#</th>
          <th class="left">Token</th>
          <th>Price</th>
          <th class="hide-sm">1h</th>
          <th :class="{ 'hide-sm': week }">24h</th>
          <th v-if="weekReady" :class="{ 'hide-sm': !week }">7d</th>
          <th class="hide-sm">24h Volume</th>
          <th class="hide-md">FDV</th>
          <th class="hide-sm spark-col">{{ lineLabel }}</th>
        </tr>
      </thead>

      <tbody v-if="loading">
        <tr v-for="n in 10" :key="n" class="skeleton">
          <td class="rank"><i style="width: 12px" /></td>
          <td class="left"><i style="width: 140px" /></td>
          <td><i style="width: 70px" /></td>
          <td class="hide-sm"><i style="width: 44px" /></td>
          <td><i style="width: 44px" /></td>
          <td class="hide-sm"><i style="width: 90px" /></td>
          <td class="hide-md"><i style="width: 70px" /></td>
          <td class="hide-sm spark-col"><i style="width: 130px; height: 30px" /></td>
        </tr>
      </tbody>

      <tbody v-else>
        <tr v-for="(t, i) in rows" :key="t.id" class="row" tabindex="0"
            @click="open(t)" @keydown.enter="open(t)">
          <td class="rank muted num">{{ ranks ? ranks[t.id] : i + 1 }}</td>
          <td class="left">
            <span class="token" :title="`${t.name} (${t.ticker})`">
              <img v-if="t.has_logo" :src="`/api/tokens/${t.id}/logo`" alt="" class="logo" width="24" height="24" loading="lazy" />
              <span class="name">{{ t.name }}</span>
              <span class="ticker muted">{{ t.ticker }}</span>
            </span>
          </td>
          <td class="num price" :title="fullPrice(t.price_ada)">{{ formatPrice(t.price_ada) }}</td>
          <td class="hide-sm"><Change :value="t.change_1h" /></td>
          <td :class="{ 'hide-sm': week }"><Change :value="t.change_24h" /></td>
          <td v-if="weekReady" :class="{ 'hide-sm': !week }"><Change :value="t.change_7d" /></td>
          <td class="hide-sm num">{{ formatADA(t.volume_24h_ada) }}</td>
          <td class="hide-md num" :title="t.fdv_ada ? formatADA(t.fdv_ada) : ''">{{ t.fdv_ada ? formatCompact(t.fdv_ada) : '' }}</td>
          <td class="hide-sm spark-col"><Sparkline :values="lineOf(t)" :width="130" /></td>
        </tr>
        <tr v-if="!rows.length">
          <td :colspan="weekReady ? 9 : 8" class="empty muted">{{ emptyText }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.table-wrap { overflow-x: auto; }
table { width: 100%; border-collapse: collapse; }
th, td { padding: 16px 10px; text-align: right; white-space: nowrap; border-bottom: 1px solid var(--line); }
th { font-size: 13px; font-weight: 700; color: var(--fg); padding-block: 14px; border-top: 1px solid var(--line); }
td { font-size: 14px; font-weight: 400; }
.left { text-align: left; }
.rank { width: 40px; text-align: left; }

.row { cursor: pointer; }
.row:hover, .row:focus-visible { background: var(--hover); outline: none; }

.token { display: inline-flex; align-items: center; gap: 8px; max-width: 300px; }
.logo { width: 24px; height: 24px; border-radius: 50%; object-fit: cover; flex: none; }
.name { font-weight: 600; min-width: 0; overflow: hidden; text-overflow: ellipsis; }
.ticker { font-weight: 400; flex: none; }
.price { font-weight: 400; }

.spark-col { width: 150px; }
.empty { text-align: center; padding-block: 40px; }

.skeleton i { display: inline-block; height: 14px; border-radius: 6px; background: linear-gradient(90deg, var(--line), #f7f9fb, var(--line)); background-size: 200% 100%; animation: shimmer 1.4s infinite; }
@keyframes shimmer { to { background-position: -200% 0; } }
@media (prefers-reduced-motion: reduce) { .skeleton i { animation: none; } }

@media (max-width: 1000px) { .hide-md { display: none; } }
@media (max-width: 700px) {
  .hide-sm, .ticker { display: none; }
  th, td { padding: 14px 6px; }
  .rank { width: 24px; }
  .token { max-width: 160px; }
}
</style>
