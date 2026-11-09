<!-- The loading box over the chart, like the loading states of TradingView and
     DexScreener: until the picked candle size has enough history to look like
     a real chart (loadWait), it covers the chart with "Loading 1 day candles
     35%", so a chart with one or two candles never looks broken. With the whole history loaded and still no candle (a
     31-day month on its 31st day needs more than the 30 days kept) it says
     there is no full month yet. -->
<script setup>
import { computed } from 'vue'
import { SIZES } from '../chart/loadWait.js'

const props = defineProps({
  size: { type: String, required: true }, // the picked candle size, e.g. '1h'
  wait: { type: Object, default: null }, // { pct } from loadWait, or null
})

const name = computed(() => SIZES[props.size]?.name ?? props.size)
</script>

<template>
  <div class="waiting" role="status" aria-live="polite">
    <div class="box">
      <template v-if="wait">
        <p class="title">Loading {{ name }} candles</p>
        <div class="bar" role="progressbar" :aria-valuenow="wait.pct" aria-valuemin="0" aria-valuemax="100">
          <span :style="{ width: `${wait.pct}%` }" />
        </div>
        <p class="pct num">{{ wait.pct }}%</p>
      </template>
      <p v-else class="title">No full {{ size === '1M' ? 'month' : 'week' }} yet</p>
    </div>
  </div>
</template>

<style scoped>
/* Covers the chart and its legend with the page colour; the toolbar stays usable. */
.waiting { position: absolute; inset: 0; z-index: 5; display: grid; place-items: center; background: var(--bg); }
.box { width: min(300px, 80%); text-align: center; }
.title { margin: 0 0 12px; font-size: 14px; font-weight: 600; color: var(--text); }
.bar { height: 4px; border-radius: 2px; background: var(--line); overflow: hidden; }
.bar span { display: block; height: 100%; background: var(--accent); border-radius: 2px; transition: width 0.4s ease; }
.pct { margin: 8px 0 0; font-size: 13px; color: var(--accent); font-weight: 600; }
</style>
