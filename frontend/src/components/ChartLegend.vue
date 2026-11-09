<!-- The line at the top left of the chart, like TradingView's: open, high,
     low, close and the change for candles; the price and the change since the
     bar before for the line. It shows the bar under the mouse, or the
     latest bar when the mouse is off the chart. -->
<script setup>
import { computed } from 'vue'
import { formatPrice, formatADA } from '../format.js'

const props = defineProps({
  type: { type: String, default: 'candles' }, // 'candles' or 'line'
  bar: { type: Object, default: null }, // the candle shown
  base: { type: Number, default: 0 }, // line: the close of the bar before
  extras: { type: Array, default: () => [] }, // indicators: [{ name, color, value, price, need }]
  fmt: { type: Function, default: formatPrice }, // writes a price (or an FDV)
})

// change is the move in the chart's unit and percent: open to close for a
// candle, the close before to this close for the line.
const change = computed(() => {
  const b = props.bar
  if (!b) return null
  const from = props.type === 'candles' ? b.open : props.base
  const to = b.close
  if (!(from > 0)) return null
  return { diff: to - from, pct: (to / from - 1) * 100 }
})

const dir = computed(() => (change.value && change.value.diff < 0 ? 'down' : 'up'))
const sign = (v) => (v < 0 ? '−' : '+')
// volDir colours the volume like its bar under the chart (PriceChart): green
// when the candle closed at or above its open, red when below, like TradingView.
const volDir = computed(() => (props.bar && props.bar.close < props.bar.open ? 'vdown' : 'vup'))
</script>

<template>
  <div v-if="bar" class="legend num" :class="{ tv: type === 'candles' }">
    <template v-if="type === 'candles'">
      <span>O<b :class="dir">{{ fmt(bar.open) }}</b></span>
      <span>H<b :class="dir">{{ fmt(bar.high) }}</b></span>
      <span>L<b :class="dir">{{ fmt(bar.low) }}</b></span>
      <span>C<b :class="dir">{{ fmt(bar.close) }}</b></span>
    </template>
    <span v-else><b :class="dir">{{ fmt(bar.close) }}</b></span>
    <span v-if="change && change.diff === 0" :class="dir">0 (0.00%)</span><!-- no move, like TradingView -->
    <span v-else-if="change" :class="dir">
      {{ sign(change.diff) }}{{ fmt(Math.abs(change.diff)).slice(1) }} ({{ sign(change.pct) }}{{ Math.abs(change.pct).toFixed(2) }}%)
    </span>
    <span class="vol">Vol<b :class="volDir">{{ formatADA(bar.volume) }}</b></span>
    <template v-if="extras.length">
      <span class="break"></span>
      <span v-for="x in extras" :key="x.name" class="extra">
        {{ x.name }}<b v-if="x.need" class="need">needs {{ x.need }} candles</b>
        <b v-else :style="{ color: x.color }">{{ x.value == null ? '—' : x.price ? fmt(x.value) : x.value.toFixed(1) }}</b>
      </span>
    </template>
  </div>
</template>

<style scoped>
.legend {
  position: absolute; z-index: 4; top: 8px; left: 12px; right: 80px;
  display: flex; flex-wrap: wrap; gap: 4px 14px; pointer-events: none;
  font-size: 13px; color: var(--text); white-space: nowrap;
}
b { font-weight: 500; margin-left: 3px; }
.up { color: var(--bull); }
.down { color: var(--bear); }
.tv .up { color: var(--chart-up); }
.tv .down { color: var(--chart-down); }
.vol b.vup { color: var(--chart-up); }
.vol b.vdown { color: var(--chart-down); }
.break { flex-basis: 100%; height: 0; }
.need { color: var(--muted); font-weight: 400; }
.extra { color: var(--muted); }
@media (max-width: 700px) { .legend { font-size: 12px; gap: 2px 10px; } }
</style>
