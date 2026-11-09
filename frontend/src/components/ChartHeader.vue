<!-- The slim bar on top of the chart, like Birdeye's and TradingView's:
     candle size | candles or line | indicators                Price / FDV
     One bar for both chart styles, so changing the style keeps the size. -->
<script setup>
import CandleSizePicker from './CandleSizePicker.vue'
import IndicatorMenu from './IndicatorMenu.vue'

const props = defineProps({
  size: { type: String, required: true }, // candle size, e.g. '1h'
  type: { type: String, required: true }, // 'candles' or 'line'
  indicators: { type: Array, required: true }, // picked indicators
  unit: { type: String, required: true }, // 'price' or 'fdv'
  canFdv: { type: Boolean, default: false }, // false when the total supply is unknown
})
const emit = defineEmits(['update:size', 'update:type', 'update:indicators', 'update:unit'])

const STYLES = [
  { key: 'candles', label: 'Candles', icon: 'M6 3v3M6 14v3M4 6h4v8H4zM14 3v5M14 15v2M12 8h4v7h-4z' },
  { key: 'line', label: 'Line', icon: 'M2.5 14.5l4.5-5 3.5 3 7-8' },
]
</script>

<template>
  <div class="bar">
    <CandleSizePicker :model-value="size" @update:model-value="emit('update:size', $event)" />
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
    <div v-if="canFdv" class="group end" role="group" aria-label="Price or fully diluted value">
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
.slash { color: var(--disabled); }
.btn { display: inline-flex; align-items: center; height: 30px; padding: 0 8px; border: 0; border-radius: 4px; background: none; color: var(--subtle); font: inherit; font-weight: 500; cursor: pointer; }
.btn:hover { background: var(--btn-hover); color: var(--text); }
.btn[aria-pressed="true"] { color: var(--accent); }
.btn:focus-visible { outline: 2px solid var(--accent); outline-offset: -2px; }
.icon { width: 30px; justify-content: center; padding: 0; }
.icon svg { width: 18px; height: 18px; fill: none; stroke: currentColor; stroke-width: 1.4; stroke-linecap: round; stroke-linejoin: round; }
.sr-only { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0); white-space: nowrap; }
</style>
