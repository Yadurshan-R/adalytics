<!-- The chart on a token page, laid out like Birdeye's and TradingView's: one
     slim bar on top (ChartHeader) for the candle size, candles or line,
     indicators and Price / FDV, with the chart under it.
     Candles and line use the same candle sizes: the line joins each candle's
     close. Each size shows all the history loaded so far, so the visitor can
     scroll back in time.
     It tells the page the low and high of the last 24 hours (event "span").
     Until the picked size has enough history to look like a real chart (about
     20 candles), ChartWaiting covers it with how far the load is. -->
<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useTokens } from '../useTokens.js'
import { chartType, candleSize, indicators, chartUnit } from '../chartPrefs.js'
import { useUnits } from '../chart/useUnits.js'
import ChartHeader from './ChartHeader.vue'
import PriceChart from './PriceChart.vue'
import ChartWaiting from './ChartWaiting.vue'
import { loadWait } from '../chart/loadWait.js'

const props = defineProps({
  id: { type: String, required: true }, // the token's id
  supply: { type: Number, default: 0 }, // total supply, for Price / FDV
})
const emit = defineEmits(['span'])
const { history, reload } = useTokens()

const candles = ref([])
const error = ref('')
const loadedSize = ref('') // the size the candles shown were fetched at
const units = useUnits(props, candles)
let timer = null

// load fetches the candles at the picked size, and again every minute.
async function load() {
  try {
    const size = candleSize.value
    const res = await fetch(`/api/tokens/${encodeURIComponent(props.id)}/chart?type=candles&size=${size}`, { cache: 'no-store' })
    const data = await res.json()
    if (res.status === 409) return reload() // not loaded yet (the backend just started)
    if (!res.ok) throw new Error(data.error || `The backend answered ${res.status}`)
    candles.value = data.candles ?? []
    loadedSize.value = size
    error.value = ''
  } catch (e) {
    error.value = e instanceof TypeError ? 'Cannot reach the backend.' : e.message
  }
}

onMounted(() => {
  load()
  timer = setInterval(load, 60_000)
})
onUnmounted(() => clearInterval(timer))
watch([() => props.id, candleSize], load)

// wait is how far the history is from what the picked size needs (null once it is there);
// the candles are fetched again the moment it is.
const wait = computed(() => loadWait(candleSize.value, history.value))
watch(wait, (now, before) => before && !now && load())
// empty is true for a 1W or 1M chart with no candle at all (fetched, not just switching).
const empty = computed(() => ['1W', '1M'].includes(candleSize.value) && loadedSize.value === candleSize.value && !candles.value.length)

// The range bar on the page: the low and high of the last 24 hours.
watch(candles, () => {
  const day = Date.now() / 1000 - 86400
  const bars = candles.value.filter((c) => c.time >= day && c.low > 0)
  if (!bars.length) return emit('span', null)
  emit('span', { low: Math.min(...bars.map((b) => b.low)), high: Math.max(...bars.map((b) => b.high)), label: '24h Range' })
})
</script>

<template>
  <div class="card">
    <ChartHeader v-model:size="candleSize" v-model:type="chartType" v-model:indicators="indicators" v-model:unit="chartUnit"
                 :can-fdv="supply > 0" />
    <p v-if="error" class="notice">{{ error }}</p>
    <PriceChart :type="chartType" :candles="units.shown.value" :indicators="indicators" :fmt="units.fmt.value"
                :fit-key="`${id}-${candleSize}-${units.key.value}`" :token="`${id}-${units.key.value}`">
      <ChartWaiting v-if="!error && (wait || empty)" :size="candleSize" :wait="wait" />
    </PriceChart>
  </div>
</template>

<style scoped>
.card { border: 1px solid var(--line); border-radius: 12px; }
.notice { margin: 0; padding: 10px 14px; background: var(--danger-bg); color: var(--danger-fg); font-size: 14px; }
</style>
