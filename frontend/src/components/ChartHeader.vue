<!-- The slim bar on top of the chart, like Birdeye's and TradingView's:
     candle size | candles or line | indicators                Price / FDV
     One bar for both chart styles, so changing the style keeps the size.
     The candle size list also has 1 week and 30 days: each shows that span
     with a fitting candle size, once that much history is loaded. -->
<script setup>
import { computed } from 'vue'
import CandleSizePicker from './CandleSizePicker.vue'
import IndicatorMenu from './IndicatorMenu.vue'

const props = defineProps({
  size: { type: String, required: true }, // candle size, e.g. '1h'
  type: { type: String, required: true }, // 'candles' or 'line'
  indicators: { type: Array, required: true }, // picked indicators
  unit: { type: String, required: true }, // 'price' or 'fdv'
  range: { type: String, default: '' }, // '7D', '30D' or '' (see chartPrefs.js)
  history: { type: Object, required: true }, // loading progress from the backend
  canFdv: { type: Boolean, default: false }, // false when the total supply is unknown
})
const emit = defineEmits(['update:size', 'update:type', 'update:indicators', 'update:unit', 'update:range'])

// loaded is how much of the history is in, while older days still load.
const loaded = computed(() => {
  const h = props.history
  return h.done || !h.days ? 100 : Math.min(99, Math.floor(((h.loaded_hours || 0) / (h.days * 24)) * 100))
})

// RANGES are the time ranges in the candle size list and the size each uses.
const RANGES = [
  { key: '7D', label: '1 week', days: 7, size: '1h' },
  { key: '30D', label: '30 days', days: 30, size: '4h' },
]
// ready is how much of a range's history is loaded, 0 to 100, as the backend
// reports it.
function ready(r) {
  const h = props.history
  const x = h.ranges?.find((k) => k.key === r.key)
  if (x) return x.ready ? 100 : Math.min(99, x.progress)
  return h.done ? 100 : Math.min(100, Math.floor(((h.loaded_hours || 0) / (r.days * 24)) * 100))
}
// ranges are RANGES with how much of each is loaded, for the list.
const ranges = computed(() => RANGES.map((r) => ({ ...r, ready: ready(r) })))
// pickRange shows a range with its candle size; picking it again goes back.
function pickRange(r) {
  if (r.ready < 100) return
  if (props.range === r.key) return emit('update:range', '')
  emit('update:size', r.size)
  emit('update:range', r.key)
}
// pickSize is a candle size picked by hand, so no range is on any more.
function pickSize(size) {
  emit('update:size', size)
  emit('update:range', '')
}

const STYLES = [
  { key: 'candles', label: 'Candles', icon: 'M6 3v3M6 14v3M4 6h4v8H4zM14 3v5M14 15v2M12 8h4v7h-4z' },
  { key: 'line', label: 'Line', icon: 'M2.5 14.5l4.5-5 3.5 3 7-8' },
]
</script>

<template>
  <div class="bar">
    <CandleSizePicker :model-value="size" :ranges="ranges" :range="range" @update:model-value="pickSize" @range="pickRange" />
    <span class="sep"></span>
    <div class="group" role="group" aria-label="Chart style">
      <button v-for="s in STYLES" :key="s.key" type="button" class="btn icon" :aria-pressed="s.key === type" :title="s.label"
              @click="emit('update:type', s.key)">
        <svg viewBox="0 0 20 20" aria-hidden="true"><path :d="s.icon" /></svg>
        <span class="sr-only">{{ s.label }}</span>
      </button>
    </div>
    <span class="sep"></span>
    <IndicatorMenu :model-value="indicators" @update:model-value="emit('update:indicators', $event)" />
    <span v-if="loaded < 100" class="hist end" title="Older candles appear as the history loads">History {{ loaded }}%</span>
    <div v-if="canFdv" class="group" :class="{ end: loaded >= 100 }" role="group" aria-label="Price or fully diluted value">
      <button type="button" class="btn" :aria-pressed="unit !== 'fdv'" @click="emit('update:unit', 'price')">Price</button>
      <span class="slash">/</span>
      <button type="button" class="btn" :aria-pressed="unit === 'fdv'" title="Fully diluted value: price × total supply" @click="emit('update:unit', 'fdv')">FDV</button>
    </div>
  </div>
</template>

<style scoped>
.bar { display: flex; align-items: center; flex-wrap: wrap; gap: 2px; min-height: 44px; padding: 4px 8px; border-bottom: 1px solid var(--line); font-size: 13px; }
.group { display: flex; align-items: center; gap: 2px; }
.end { margin-left: auto; }
.sep { width: 1px; height: 20px; margin: 0 6px; background: var(--line); }
.slash { color: #b2b5be; }
.hist { margin-right: 8px; font-size: 12px; color: #9aa0ac; }
.btn { display: inline-flex; align-items: center; height: 30px; padding: 0 8px; border: 0; border-radius: 4px; background: none; color: #6a6d78; font: inherit; font-weight: 500; cursor: pointer; }
.btn:hover { background: #f0f3fa; color: #131722; }
.btn[aria-pressed="true"] { color: #2962ff; }
.btn:focus-visible { outline: 2px solid #2962ff; outline-offset: -2px; }
.icon { width: 30px; justify-content: center; padding: 0; }
.icon svg { width: 18px; height: 18px; fill: none; stroke: currentColor; stroke-width: 1.4; stroke-linecap: round; stroke-linejoin: round; }
.sr-only { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0); white-space: nowrap; }
</style>
