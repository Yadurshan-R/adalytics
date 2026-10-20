<!-- The filter bar on the Large Trades page, laid out like Birdeye's Find
     Trades: the token search (TokenPicker), the Time and Size lists and the
     wallet box side by side (two by two on a phone), and Reset all beside
     the Wallet title. The page owns the values; this bar only shows and
     changes them (v-model on each). -->
<script setup>
import { SIZES } from '../tradeSizes.js'
import TokenPicker from './TokenPicker.vue'

const token = defineModel('token', { type: String, default: '' }) // a token id, or '' for all
const hours = defineModel('hours', { type: Number, default: 24 })
const size = defineModel('size', { type: String, default: 'all' }) // 'all' or a SIZES key
const wallet = defineModel('wallet', { type: String, default: '' })
const props = defineProps({
  tokens: { type: Array, default: () => [] }, // [{ id, ticker, name, has_logo }]
  ranges: { type: Array, default: () => [] }, // [{ key: '7D', ready, progress }] from the backend
  changed: { type: Boolean, default: false }, // a filter is not at its default
})
defineEmits(['reset'])

// TIMES are the time list's choices. 7 and 30 days need that much history:
// after a restart they are grey with "Loading 35%" until it is in (like the
// chart's list).
const TIMES = [{ hours: 1, label: 'Last 1 hour' }, { hours: 4, label: 'Last 4 hours' }, { hours: 24, label: 'Last 24 hours' },
  { hours: 168, label: 'Last 7 days', range: '7D' }, { hours: 720, label: 'Last 30 days', range: '30D' }]
// loading is how much of a time choice is loaded, or null when it can be picked.
function loading(t) {
  if (!t.range) return null
  const r = props.ranges.find((x) => x.key === t.range)
  return r?.ready ? null : r?.progress ?? 0
}
</script>

<template>
  <div class="filters">
    <div class="box" role="group" aria-label="Token">
      <span class="label">Token</span>
      <TokenPicker v-model="token" :tokens="tokens" />
    </div>
    <div class="box" role="group" aria-label="Time">
      <span class="label">Time</span>
      <select v-model.number="hours" aria-label="Time">
        <option v-for="t in TIMES" :key="t.hours" :value="t.hours" :disabled="loading(t) !== null">
          {{ t.label }}{{ loading(t) !== null ? ` · Loading ${loading(t)}%` : '' }}
        </option>
      </select>
    </div>
    <div class="box" role="group" aria-label="Size">
      <span class="label">Size</span>
      <select v-model="size" aria-label="Size">
        <option value="all">All sizes</option>
        <option v-for="s in SIZES" :key="s.key" :value="s.key">{{ s.icon }} {{ s.label }} · {{ s.range }}</option>
      </select>
    </div>
    <div class="box" role="group" aria-label="Wallet">
      <span class="label row">Wallet
        <button type="button" class="reset" :disabled="!changed" @click="$emit('reset')">
          <svg viewBox="0 0 16 16" aria-hidden="true"><path d="M13 8a5 5 0 1 1-1.5-3.6M13 2v3h-3" /></svg>Reset all
        </button>
      </span>
      <input v-model.trim="wallet" class="text" type="search" placeholder="stake1… or addr1…" aria-label="Wallet" spellcheck="false" />
    </div>
  </div>
</template>

<style scoped>
.filters { display: grid; grid-template-columns: 1fr 1fr 1fr 1.3fr; border: 1px solid var(--line); border-radius: 8px; margin-bottom: 16px; }
.box { min-width: 0; padding: 12px 16px 14px; border-left: 1px solid var(--line); }
.box:first-child { border-left: 0; }
.label { display: block; margin-bottom: 10px; font-size: 12px; font-weight: 600; color: var(--muted); }
.row { display: flex; justify-content: space-between; align-items: center; }
.reset { display: inline-flex; align-items: center; gap: 4px; margin: -4px -6px; padding: 2px 6px; border: 0; border-radius: 4px; background: none; color: #2962ff; font: inherit; font-size: 12px; font-weight: 500; cursor: pointer; }
.reset:hover:not(:disabled) { background: #f0f3fa; }
.reset:disabled { color: var(--muted); opacity: 0.5; cursor: default; }
.reset svg { width: 12px; height: 12px; fill: none; stroke: currentColor; stroke-width: 1.8; stroke-linecap: round; }
select, .text { width: 100%; height: 34px; padding: 0 10px; border: 1px solid var(--line); border-radius: 6px; background: var(--bg, #fff); color: var(--fg); font: inherit; font-size: 14px; }
.text { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 13px; }
select:focus, .text:focus { outline: 2px solid #2962ff33; border-color: #2962ff; }
@media (max-width: 860px) {
  .filters { grid-template-columns: 1fr 1fr; }
  .box { padding: 10px 12px 12px; }
  .box:nth-child(3) { border-left: 0; }
  .box:nth-child(n+3) { border-top: 1px solid var(--line); }
}
</style>
