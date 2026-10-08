<!--
  CoinGecko-style price chart (TradingView Lightweight Charts), as a line or as
  candlesticks.
  - Line: area line, green if the price is up over the range, red if down. On
    hover: dashed line + dot; the line stays bold up to the cursor and fades
    after it; the colour compares the hovered price with the first price.
  - Candles: green when a slot closed higher than it opened, red when lower.
    The tooltip shows open, high, low, close and the change.
  - Volume bars underneath (grey for the line, green/red for candles).
-->
<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { createChart, AreaSeries, CandlestickSeries, HistogramSeries, LineStyle, CrosshairMode } from 'lightweight-charts'
import { formatPrice, formatADA, formatDateTime } from '../format.js'

const props = defineProps({
  type: { type: String, default: 'line' }, // 'line' or 'candles'
  points: { type: Array, default: () => [] }, // line: [{ time, price, volume }]
  candles: { type: Array, default: () => [] }, // candles: [{ time, open, high, low, close, volume }]
})

const BULL = '#16c784'
const BEAR = '#ea3943'
const MUTED = '#58667e'
const rgba = (hex, a) => {
  const n = parseInt(hex.slice(1), 16)
  return `rgba(${n >> 16},${(n >> 8) & 255},${n & 255},${a})`
}

const el = ref(null)
const tip = ref({ show: false, x: 0, y: 0, time: 0 })

let chart, area, candleSeries, volume, lastLine
let rows = [] // line points
let bars = [] // candles
let lastIdx = null

const isCandles = () => props.type === 'candles'

// Line mode: colour and bold/faded split for the hovered point (-1 = none).
function paint(idx) {
  if (isCandles() || !rows.length || idx === lastIdx) return
  lastIdx = idx
  const first = rows[0].price
  const shown = idx < 0 ? rows[rows.length - 1].price : rows[idx].price
  const col = shown >= first ? BULL : BEAR
  const solid = { lineColor: col, topColor: rgba(col, 0.24), bottomColor: rgba(col, 0.02) }
  const faded = { lineColor: rgba(col, 0.3), topColor: rgba(col, 0.06), bottomColor: rgba(col, 0) }
  area.setData(rows.map((p, k) => ({ time: p.time, value: p.price, ...(idx < 0 || k <= idx ? solid : faded) })))
  area.applyOptions({ crosshairMarkerBackgroundColor: col, crosshairMarkerBorderColor: rgba(col, 0.25) })
  chart.applyOptions({ crosshair: { vertLine: { color: col } } })
  lastLine.applyOptions({ price: rows[rows.length - 1].price, color: col, axisLabelColor: col, axisLabelTextColor: '#ffffff', axisLabelVisible: true })
}

function load() {
  if (!chart) return
  lastIdx = null
  if (isCandles()) {
    rows = []
    bars = props.candles.filter((c) => c.close > 0)
    area.setData([])
    lastLine.applyOptions({ axisLabelVisible: false })
    candleSeries.setData(bars.map((c) => ({ time: c.time, open: c.open, high: c.high, low: c.low, close: c.close })))
    volume.setData(bars.map((c) => ({ time: c.time, value: c.volume, color: rgba(c.close >= c.open ? BULL : BEAR, 0.35) })))
    chart.applyOptions({ crosshair: { mode: CrosshairMode.Normal, vertLine: { color: MUTED } } })
  } else {
    bars = []
    rows = props.points.filter((p) => p.price > 0)
    candleSeries.setData([])
    volume.setData(rows.map((p) => ({ time: p.time, value: p.volume, color: '#e3e6eb' })))
    chart.applyOptions({ crosshair: { mode: CrosshairMode.Magnet } })
    if (rows.length) paint(-1)
    else area.setData([])
  }
  chart.timeScale().fitContent()
}

const priceFormat = { type: 'custom', minMove: 1e-12, formatter: formatPrice }

onMounted(() => {
  chart = createChart(el.value, {
    autoSize: true,
    layout: { background: { type: 'solid', color: '#ffffff' }, textColor: MUTED, fontFamily: 'Inter, system-ui, sans-serif', fontSize: 12 },
    grid: { vertLines: { visible: false }, horzLines: { color: '#eff2f5' } },
    rightPriceScale: { borderVisible: false },
    timeScale: {
      borderVisible: false, fixLeftEdge: true, fixRightEdge: true,
      timeVisible: true, secondsVisible: false,
      // Hours like "16:00"; a new day shows as "5 Oct", like CoinGecko.
      tickMarkFormatter: (t, type) => {
        const d = new Date(t * 1000)
        if (type === 0) return String(d.getFullYear())
        if (type === 1) return d.toLocaleString('en-GB', { month: 'short' })
        if (type === 2) return d.toLocaleString('en-GB', { day: 'numeric', month: 'short' })
        return d.toLocaleString('en-GB', { hour: '2-digit', minute: '2-digit', hour12: false })
      },
    },
    crosshair: {
      mode: CrosshairMode.Magnet,
      vertLine: { style: LineStyle.Dashed, width: 1, labelVisible: false },
      horzLine: { visible: false, labelVisible: false },
    },
    handleScroll: false,
    handleScale: false,
  })

  area = chart.addSeries(AreaSeries, {
    lineWidth: 2, priceLineVisible: false, lastValueVisible: false,
    crosshairMarkerRadius: 5, crosshairMarkerBorderWidth: 3, priceFormat,
  })
  candleSeries = chart.addSeries(CandlestickSeries, {
    upColor: BULL, downColor: BEAR, wickUpColor: BULL, wickDownColor: BEAR, borderVisible: false,
    priceLineVisible: false, lastValueVisible: true, priceFormat,
  })
  for (const s of [area, candleSeries]) s.priceScale().applyOptions({ scaleMargins: { top: 0.08, bottom: 0.22 } })
  volume = chart.addSeries(HistogramSeries, { priceScaleId: 'vol', priceLineVisible: false, lastValueVisible: false })
  chart.priceScale('vol').applyOptions({ scaleMargins: { top: 0.84, bottom: 0 }, visible: false })
  lastLine = area.createPriceLine({ price: 0, lineVisible: false, axisLabelVisible: false, title: '' })

  chart.subscribeCrosshairMove((param) => {
    const list = isCandles() ? bars : rows
    if (!list.length || !param.point || param.logical == null || param.point.x < 0 || param.point.y < 0) {
      tip.value = { ...tip.value, show: false }
      paint(-1)
      return
    }
    const idx = Math.max(0, Math.min(list.length - 1, Math.round(param.logical)))
    paint(idx)
    const w = el.value.clientWidth
    const tw = 290 // tooltip width
    const x = param.point.x + 24 + tw > w - 70 ? param.point.x - 24 - tw : param.point.x + 24
    const y = Math.max(8, Math.min(el.value.clientHeight - 190, param.point.y - 50))
    tip.value = { show: true, x: Math.max(0, x), y, ...list[idx] }
  })

  load()
})

watch(() => [props.type, props.points, props.candles], load)

onBeforeUnmount(() => chart && chart.remove())

const pct = (c) => (c.open > 0 ? ((c.close / c.open - 1) * 100) : 0)
</script>

<template>
  <div class="wrap">
    <div ref="el" class="chart"></div>
    <div v-show="tip.show" class="tip" :style="{ left: tip.x + 'px', top: tip.y + 'px' }">
      <div class="t">{{ formatDateTime(tip.time) }}</div>
      <template v-if="type === 'candles'">
        <div class="ohlc num">
          <span class="k">Open</span><span class="v">{{ formatPrice(tip.open) }}</span>
          <span class="k">High</span><span class="v">{{ formatPrice(tip.high) }}</span>
          <span class="k">Low</span><span class="v">{{ formatPrice(tip.low) }}</span>
          <span class="k">Close</span><span class="v">{{ formatPrice(tip.close) }}</span>
          <span class="k">Change</span>
          <span class="v" :class="tip.close >= tip.open ? 'up' : 'down'">{{ pct(tip) >= 0 ? '+' : '' }}{{ pct(tip).toFixed(2) }}%</span>
          <span class="k">Vol</span><span class="v">{{ formatADA(tip.volume) }}</span>
        </div>
      </template>
      <template v-else>
        <div><span class="k">Price:</span> <span class="v num">{{ formatPrice(tip.price) }}</span></div>
        <div><span class="k">Vol:</span> <span class="v num">{{ formatADA(tip.volume) }}</span></div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.wrap { position: relative; }
.chart { height: 440px; }
@media (max-width: 700px) { .chart { height: 320px; } }
.tip {
  position: absolute; z-index: 5; pointer-events: none; width: 290px;
  background: #fff; border-radius: 12px; padding: 14px 18px; font-size: 14px; line-height: 1.7;
  box-shadow: 0 6px 24px rgba(13, 20, 33, 0.12), 0 0 0 1px rgba(13, 20, 33, 0.04);
}
.t { margin-bottom: 4px; white-space: nowrap; }
.k { color: var(--muted); }
.v { font-weight: 700; }
.ohlc { display: grid; grid-template-columns: auto 1fr; column-gap: 16px; }
.ohlc .v { text-align: right; }
</style>
