<!-- The chart on a token page: Line or Candles, the buttons above it, and the
     loading box while a range's history is still coming in.
     - Line: 24H, 7D or 30D, CoinGecko style.
     - Candles: the candle size (15m, 1h, 4h, 1D), TradingView style. Each size
       loads all the history there is, so the visitor can scroll back in time.
     It tells the page the low and high for the range bar (event "span"). -->
<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useTokens } from '../useTokens.js'
import { chartType, chartRange, candleSize, indicators } from '../chartPrefs.js'
import PriceChart from './PriceChart.vue'
import IndicatorMenu from './IndicatorMenu.vue'

const props = defineProps({ id: { type: String, required: true } }) // the token's id
const emit = defineEmits(['span'])
const { history, reload } = useTokens()

const TYPES = [{ key: 'line', label: 'Line' }, { key: 'candles', label: 'Candles' }]
const SIZES = ['15m', '1h', '4h', '1D']
const isCandles = computed(() => chartType.value === 'candles')
const setType = (key) => (chartType.value = key)
const setSize = (key) => (candleSize.value = key)
const setRange = (key) => (chartRange.value = key)

// Range buttons for the line. Every range can be clicked; if its history is
// still loading, the loading percentage shows until the chart is ready.
const ranges = computed(() => (history.value.ranges?.length ? history.value.ranges : [{ key: '24H', progress: 100, ready: true }]))
// selected is the line range picked now (24H, 7D or 30D).
const selected = computed(() => ranges.value.find((r) => r.key === chartRange.value) ?? ranges.value[0])
// waiting is true while the picked line range is still loading. Candles never
// wait: they show the history loaded so far.
const waiting = computed(() => !isCandles.value && !selected.value.ready)
// view is what the chart shows now, e.g. "line-24H" or "candles-1h".
const view = computed(() => (isCandles.value ? `candles-${candleSize.value}` : `line-${selected.value.key}`))

const points = ref([])
const candles = ref([])
const error = ref('')
let timer = null
let progressTimer = null

// load fetches the chart, and again every minute.
async function load() {
  if (waiting.value) {
    points.value = []
    candles.value = []
    return
  }
  const q = isCandles.value ? `type=candles&size=${candleSize.value}` : `range=${selected.value.key}`
  try {
    const res = await fetch(`/api/tokens/${encodeURIComponent(props.id)}/chart?${q}`, { cache: 'no-store' })
    const data = await res.json()
    if (res.status === 409) return reload() // not loaded after all (the backend restarted)
    if (!res.ok) throw new Error(data.error || `The backend answered ${res.status}`)
    points.value = data.points ?? []
    candles.value = data.candles ?? []
    error.value = ''
  } catch (e) {
    error.value = e instanceof TypeError ? 'Cannot reach the backend.' : e.message
  }
}

onMounted(() => {
  load()
  timer = setInterval(load, 60_000)
})
onUnmounted(() => {
  clearInterval(timer)
  clearInterval(progressTimer)
})
watch([() => props.id, view], load)

// While a loading range is picked, ask the backend for the progress every 3
// seconds (it reads its own memory, no Koios request), and draw the chart as
// soon as the range is ready.
watch(waiting, (now, before) => {
  clearInterval(progressTimer)
  if (now) progressTimer = setInterval(reload, 3000)
  else if (before) load()
}, { immediate: true })

// The range bar on the page: the low and high of the line's range, or of the
// last 24 hours of candles.
watch([points, candles, waiting], () => {
  if (waiting.value) return emit('span', null)
  const day = Date.now() / 1000 - 86400
  const bars = isCandles.value ? candles.value.filter((c) => c.time >= day) : points.value
  const lows = bars.map((b) => b.low ?? b.price).filter((v) => v > 0)
  const highs = bars.map((b) => b.high ?? b.price).filter((v) => v > 0)
  if (!lows.length) return emit('span', null)
  const label = isCandles.value ? '24h' : { '24H': '24h', '7D': '7d', '30D': '30d' }[selected.value.key]
  emit('span', { low: Math.min(...lows), high: Math.max(...highs), label: label + ' Range' })
})
</script>

<template>
  <div class="toolbar">
    <div class="left">
      <div class="pick" role="group" aria-label="Chart type">
        <button v-for="t in TYPES" :key="t.key" type="button" :aria-pressed="t.key === chartType" @click="setType(t.key)">{{ t.label }}</button>
      </div>
      <IndicatorMenu v-model="indicators" />
    </div>
    <div v-if="isCandles" class="pick" role="group" aria-label="Candle size">
      <button v-for="s in SIZES" :key="s" type="button" :aria-pressed="s === candleSize" @click="setSize(s)">{{ s }}</button>
    </div>
    <div v-else class="pick" role="group" aria-label="Time range">
      <button v-for="r in ranges" :key="r.key" type="button" :aria-pressed="r.key === selected.key" @click="setRange(r.key)">{{ r.key }}</button>
    </div>
  </div>
  <p v-if="error" class="notice">{{ error }}</p>
  <div v-if="waiting" class="progress" role="progressbar" :aria-label="`${selected.key} loading`"
       :aria-valuenow="selected.progress" aria-valuemin="0" aria-valuemax="100">
    <span class="p-key">{{ selected.key }}</span>
    <span class="p-pct num">{{ selected.progress }}%</span>
    <span class="p-bar"><span :style="{ width: Math.max(selected.progress, 1) + '%' }"></span></span>
  </div>
  <PriceChart v-else :type="chartType" :points="points" :candles="candles" :indicators="indicators" :fit-key="`${id}-${view}`" :token="id" />
</template>

<style scoped>
.toolbar { display: flex; justify-content: space-between; gap: 12px; flex-wrap: wrap; margin-bottom: 12px; }
.left { display: flex; gap: 8px; }
.pick { display: inline-flex; background: var(--chip); border-radius: 10px; padding: 4px; gap: 2px; }
.pick button { font: inherit; font-size: 14px; font-weight: 600; color: var(--muted); background: none; border: 0; padding: 6px 12px; border-radius: 8px; cursor: pointer; }
.pick button[aria-pressed="true"] { background: var(--bg); color: var(--fg); box-shadow: 0 1px 3px rgba(13, 20, 33, 0.12); }
.pick button:focus-visible { outline: 2px solid var(--fg); outline-offset: 1px; }
.progress { height: 520px; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 10px; border-radius: 12px; background: #fafbfd; }
@media (max-width: 700px) { .progress { height: 320px; } }
.p-key { font-size: 15px; font-weight: 600; color: var(--muted); }
.p-pct { font-size: 40px; font-weight: 700; letter-spacing: -0.02em; }
.p-bar { width: min(320px, 70%); height: 8px; border-radius: 4px; background: var(--chip); overflow: hidden; }
.p-bar span { display: block; height: 100%; border-radius: 4px; background: var(--bull); transition: width 0.6s ease; }
@media (prefers-reduced-motion: reduce) { .p-bar span { transition: none; } }
.notice { margin: 0 0 12px; padding: 10px 14px; border-radius: 10px; background: #fdecee; color: #a0202c; }
</style>
