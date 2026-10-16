<!-- The chart on a token page, laid out like Birdeye's and TradingView's: one
     slim bar on top (ChartHeader) for the candle size, candles or line,
     indicators and Price / FDV, with the chart under it.
     Candles and line use the same candle sizes: the line joins each candle's
     close. Each size shows all the history loaded so far, so the visitor can
     scroll back in time.
     It tells the page the low and high of the last 24 hours (event "span"). -->
<script setup>
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { useTokens } from '../useTokens.js'
import { chartType, candleSize, indicators, chartUnit, chartRange } from '../chartPrefs.js'
import { useUnits } from '../chart/useUnits.js'
import ChartHeader from './ChartHeader.vue'
import PriceChart from './PriceChart.vue'

const props = defineProps({
  id: { type: String, required: true }, // the token's id
  supply: { type: Number, default: 0 }, // total supply, for Price / FDV
})
const emit = defineEmits(['span'])
const { history, reload } = useTokens()

// SPAN is how many seconds each time range shows.
const SPAN = { '7D': 7 * 86400, '30D': 30 * 86400 }

const candles = ref([])
const error = ref('')
const units = useUnits(props, candles)
let timer = null

// load fetches the candles at the picked size, and again every minute.
async function load() {
  try {
    const res = await fetch(`/api/tokens/${encodeURIComponent(props.id)}/chart?type=candles&size=${candleSize.value}`, { cache: 'no-store' })
    const data = await res.json()
    if (res.status === 409) return reload() // not loaded yet (the backend just started)
    if (!res.ok) throw new Error(data.error || `The backend answered ${res.status}`)
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
onUnmounted(() => clearInterval(timer))
watch([() => props.id, candleSize], load)

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
    <ChartHeader v-model:size="candleSize" v-model:type="chartType" v-model:indicators="indicators" v-model:unit="chartUnit" v-model:range="chartRange"
                 :history="history" :can-fdv="supply > 0" />
    <p v-if="error" class="notice">{{ error }}</p>
    <PriceChart :type="chartType" :candles="units.shown.value" :indicators="indicators" :fmt="units.fmt.value"
                :fit-key="`${id}-${candleSize}-${units.key.value}-${chartRange}`" :span="SPAN[chartRange] || 0" :token="`${id}-${units.key.value}`" />
  </div>
</template>

<style scoped>
.card { border: 1px solid var(--line); border-radius: 12px; }
.notice { margin: 0; padding: 10px 14px; background: #fdecee; color: #a0202c; font-size: 14px; }
</style>
