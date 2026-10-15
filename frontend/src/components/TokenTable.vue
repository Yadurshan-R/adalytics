<!-- The token table, laid out like Birdeye's: rank, token (ticker, name, age
     and short policy ID with a copy button), price, FDV, volume and trades for
     the period picked, liquidity, and the 5m, 1h, 4h and 24h change. Every
     column title sorts. Numbers whose history is still loading are faded.
     Grey placeholder rows while loading; a click on a row opens the token. -->
<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { formatPrice, fullPrice, formatCompact } from '../format.js'
import { COLUMNS, HOURS, periodOf, signedPct } from '../listSort.js'

const props = defineProps({
  rows: { type: Array, required: true }, // the tokens to show, in order
  period: { type: String, required: true }, // '5m', '1h', '4h' or '24h'
  sort: { type: Object, default: null }, // the column picked { key, dir }, or null
  view: { type: String, default: 'volume' }, // decides the order when no column is picked
  loading: { type: Boolean, default: false },
  hours: { type: Number, default: 24 }, // hours of history loaded so far
  periodReady: { type: Boolean, default: true }, // the period picked is fully loaded
  emptyText: { type: String, default: '' },
})
const emit = defineEmits(['sort'])

const router = useRouter()
// open goes to a token's page.
const open = (t) => router.push({ name: 'token', params: { ticker: t.id } })

// sortedBy is the column the rows are in order of, and which way.
function sortedBy(key) {
  const own = { volume: ['volume', -1], gainers: [`c${props.period}`, -1], losers: [`c${props.period}`, 1] }[props.view]
  const [k, dir] = props.sort ? [props.sort.key, props.sort.dir] : own
  return k === key ? dir : 0
}
// title is a column's heading; Vol and Txns name the period.
const title = (c) => (c.period ? `${c.title} ${props.period}` : c.title)
// hidden hides a column on a phone: only price, volume and the period's change stay.
const hidden = (c) => ['fdv', 'txns', 'liquidity'].includes(c.key) || (c.change && c.change !== props.period)
// faded is true for numbers whose history is not fully loaded yet.
const faded = (c) => (c.change ? props.hours < HOURS[c.change] : c.period && !props.periodReady)

// cell writes one number in its column's style.
function cell(c, t) {
  const p = periodOf(t, props.period)
  switch (c.key) {
    case 'price': return formatPrice(t.price_ada)
    case 'fdv': return t.fdv_ada ? formatCompact(t.fdv_ada) : '—'
    case 'volume': return p.volume_ada ? formatCompact(p.volume_ada) : '—'
    case 'txns': return (p.buys + p.sells).toLocaleString('en-US')
    case 'liquidity': return formatCompact(t.reserve_ada * 2)
  }
  return signedPct(periodOf(t, c.change).change_pct)
}
// tone colours a change green or red.
const tone = (c, t) => (c.change ? (v => (v > 0.005 ? 'up' : v < -0.005 ? 'down' : 'muted'))(periodOf(t, c.change).change_pct) : '')

// age shows how long ago a token was created: 4d, 8mo, 2y.
function age(iso) {
  if (!iso) return ''
  const d = (Date.now() - new Date(iso).getTime()) / 86_400_000
  return d < 1 ? `${Math.max(1, Math.round(d * 24))}h` : d < 60 ? `${Math.round(d)}d` : d < 365 ? `${Math.round(d / 30)}mo` : `${Math.floor(d / 365)}y`
}
const copied = ref('')
// copy puts a policy ID on the clipboard without opening the row.
function copy(t) {
  navigator.clipboard?.writeText(t.policy_id)
  copied.value = t.id
  setTimeout(() => (copied.value = ''), 1500)
}
</script>

<template>
  <div class="table-wrap">
    <table>
      <thead>
        <tr>
          <th class="rank">#</th>
          <th class="left">Token</th>
          <th v-for="c in COLUMNS" :key="c.key" :class="{ 'hide-sm': hidden(c) }" :aria-sort="sortedBy(c.key) ? (sortedBy(c.key) > 0 ? 'ascending' : 'descending') : 'none'">
            <button type="button" class="sort" :class="{ on: sortedBy(c.key) }" @click="emit('sort', c.key)">
              {{ title(c) }}
              <svg viewBox="0 0 8 12" aria-hidden="true"><path :class="{ lit: sortedBy(c.key) > 0 }" d="M4 1l3 3.5H1z" /><path :class="{ lit: sortedBy(c.key) < 0 }" d="M4 11l3-3.5H1z" /></svg>
            </button>
          </th>
        </tr>
      </thead>
      <tbody v-if="loading">
        <tr v-for="n in 10" :key="n" class="skeleton">
          <td class="rank"><i style="width: 12px" /></td>
          <td class="left"><i style="width: 150px; height: 28px" /></td>
          <td v-for="c in COLUMNS" :key="c.key" :class="{ 'hide-sm': hidden(c) }"><i style="width: 56px" /></td>
        </tr>
      </tbody>
      <tbody v-else>
        <tr v-for="(t, i) in rows" :key="t.id" class="row" tabindex="0" @click="open(t)" @keydown.enter="open(t)">
          <td class="rank muted num">{{ i + 1 }}</td>
          <td class="left">
            <div class="token">
              <img v-if="t.has_logo" :src="`/api/tokens/${t.id}/logo`" alt="" class="logo" width="28" height="28" loading="lazy" />
              <span v-else class="logo blank" aria-hidden="true">{{ t.ticker.slice(0, 1) }}</span>
              <div class="names">
                <div><b>{{ t.ticker }}</b> <span class="muted name">{{ t.name }}</span></div>
                <div class="sub muted">
                  <span v-if="t.created" class="age">{{ age(t.created) }}</span>
                  <span class="mono">{{ t.policy_id.slice(0, 6) }}…{{ t.policy_id.slice(-4) }}</span>
                  <button type="button" class="copy" :title="copied === t.id ? 'Copied' : 'Copy policy ID'" @click.stop="copy(t)">
                    <svg v-if="copied === t.id" viewBox="0 0 16 16" aria-hidden="true"><path d="M3 8.5l3 3 7-7" /></svg>
                    <svg v-else viewBox="0 0 16 16" aria-hidden="true"><rect x="5.5" y="5.5" width="8" height="8" rx="1.5" /><path d="M10.5 3.5v-.5A1.5 1.5 0 0 0 9 1.5H4A1.5 1.5 0 0 0 2.5 3v5A1.5 1.5 0 0 0 4 9.5h.5" /></svg>
                  </button>
                </div>
              </div>
            </div>
          </td>
          <td v-for="c in COLUMNS" :key="c.key" class="num" :class="[tone(c, t), { 'hide-sm': hidden(c), faded: faded(c) }]"
              :title="c.key === 'price' ? fullPrice(t.price_ada) : null">{{ cell(c, t) }}</td>
        </tr>
        <tr v-if="!rows.length">
          <td :colspan="COLUMNS.length + 2" class="empty muted">{{ emptyText }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.table-wrap { overflow-x: auto; }
table { width: 100%; border-collapse: collapse; }
th, td { padding: 10px 10px; text-align: right; white-space: nowrap; border-bottom: 1px solid var(--line); }
th { padding-block: 8px; font-size: 12px; font-weight: 500; color: var(--muted); }
td { font-size: 14px; }
.left { text-align: left; }
.rank { width: 28px; text-align: left; }
.sort { display: inline-flex; align-items: center; gap: 4px; padding: 0; border: 0; background: none; color: inherit; font: inherit; cursor: pointer; }
.sort:hover, .sort.on { color: var(--fg); }
.sort svg { width: 7px; height: 11px; }
.sort path { fill: #c5ccd8; }
.sort path.lit { fill: #2962ff; }
.row { cursor: pointer; }
.row:hover, .row:focus-visible { background: var(--hover); outline: none; }
.token { display: flex; align-items: center; gap: 10px; }
.logo { width: 28px; height: 28px; border-radius: 50%; object-fit: cover; flex: none; }
.blank { display: inline-flex; align-items: center; justify-content: center; background: var(--chip); color: var(--muted); font-size: 12px; font-weight: 700; }
.names { line-height: 1.3; min-width: 0; }
.names b { font-weight: 600; }
.name { display: inline-block; max-width: 160px; overflow: hidden; text-overflow: ellipsis; vertical-align: bottom; font-size: 13px; }
.sub { display: flex; align-items: center; gap: 6px; font-size: 12px; }
.age { color: var(--bull); font-weight: 600; }
.mono { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
.copy { display: inline-flex; padding: 1px; border: 0; background: none; color: inherit; cursor: pointer; }
.copy:hover { color: var(--fg); }
.copy svg { width: 13px; height: 13px; fill: none; stroke: currentColor; stroke-width: 1.3; stroke-linecap: round; stroke-linejoin: round; }
.faded { opacity: 0.35; }
.empty { text-align: center; padding-block: 40px; }
.skeleton i { display: inline-block; height: 14px; border-radius: 6px; background: linear-gradient(90deg, var(--line), #f7f9fb, var(--line)); background-size: 200% 100%; animation: shimmer 1.4s infinite; }
@keyframes shimmer { to { background-position: -200% 0; } }
@media (prefers-reduced-motion: reduce) { .skeleton i { animation: none; } }
@media (max-width: 700px) {
  .hide-sm, .name, .sub .mono, .copy { display: none; }
  th, td { padding: 10px 5px; }
}
</style>
