<!-- The line at the top left of the chart, like TradingView's: open, high,
     low, close and the change for candles; the price and the change since the
     start of the range for the line. It shows the bar under the mouse, or the
     latest bar when the mouse is off the chart. -->
<script setup>
import { computed } from 'vue'
import { formatPrice, formatADA } from '../format.js'

const props = defineProps({
  type: { type: String, default: 'line' }, // 'line' or 'candles'
  bar: { type: Object, default: null }, // a candle or a line point
  base: { type: Number, default: 0 }, // line: the first price of the range
  extras: { type: Array, default: () => [] }, // indicators: [{ name, color, value, price }]
  fmt: { type: Function, default: formatPrice }, // writes a price (or an FDV), in ₳ or $
  volFmt: { type: Function, default: formatADA }, // writes the volume
})

// change is the move in the chart's unit and percent: open to close for a
// candle, first price to this price for the line.
const change = computed(() => {
  const b = props.bar
  if (!b) return null
  const from = props.type === 'candles' ? b.open : props.base
  const to = props.type === 'candles' ? b.close : b.price
  if (!(from > 0)) return null
  return { diff: to - from, pct: (to / from - 1) * 100 }
})

const dir = computed(() => (change.value && change.value.diff < 0 ? 'down' : 'up'))
const sign = (v) => (v < 0 ? '−' : '+')
</script>

<template>
  <div v-if="bar" class="legend num" :class="{ tv: type === 'candles' }">
    <template v-if="type === 'candles'">
      <span>O<b :class="dir">{{ fmt(bar.open) }}</b></span>
      <span>H<b :class="dir">{{ fmt(bar.high) }}</b></span>
      <span>L<b :class="dir">{{ fmt(bar.low) }}</b></span>
      <span>C<b :class="dir">{{ fmt(bar.close) }}</b></span>
    </template>
    <span v-else><b :class="dir">{{ fmt(bar.price) }}</b></span>
    <span v-if="change" :class="dir">
      {{ sign(change.diff) }}{{ fmt(Math.abs(change.diff)).slice(1) }} ({{ sign(change.pct) }}{{ Math.abs(change.pct).toFixed(2) }}%)
    </span>
    <span class="vol">Vol<b>{{ volFmt(bar.volume) }}</b></span>
    <template v-if="extras.length">
      <span class="break"></span>
      <span v-for="x in extras" :key="x.name" class="extra">
        {{ x.name }}<b :style="{ color: x.color }">{{ x.value == null ? '—' : x.price ? fmt(x.value) : x.value.toFixed(1) }}</b>
      </span>
    </template>
  </div>
</template>

<style scoped>
.legend {
  position: absolute; z-index: 4; top: 8px; left: 12px; right: 80px;
  display: flex; flex-wrap: wrap; gap: 4px 14px; pointer-events: none;
  font-size: 13px; color: #131722; white-space: nowrap;
}
b { font-weight: 500; margin-left: 3px; }
.up { color: #16c784; }
.down { color: #ea3943; }
.tv .up { color: #089981; }
.tv .down { color: #f23645; }
.vol b { color: #131722; }
.break { flex-basis: 100%; height: 0; }
.extra { color: #58667e; }
@media (max-width: 700px) { .legend { font-size: 12px; gap: 2px 10px; } }
</style>
