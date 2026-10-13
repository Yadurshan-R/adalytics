<!--
  CoinGecko-style price chart (TradingView Lightweight Charts), as a line or as
  candlesticks.
  - Line: area line, green if the price is up over the range, red if down. On
    hover: dashed line + dot; the line stays bold up to the cursor and fades
    after it; the colour compares the hovered price with the first price.
  - Candles: green when a slot closed higher than it opened, red when lower.
  - The line at the top left (ChartLegend) shows the bar under the mouse.
  - Volume bars underneath (grey for the line, green/red for candles).
  - A drawing toolbar on the left (ChartToolbar), with zoom and scroll.
-->
<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { createChart, AreaSeries, CandlestickSeries, HistogramSeries, LineStyle, CrosshairMode } from 'lightweight-charts'
import { formatPrice } from '../format.js'
import { useDrawings, SCROLL, SCALE } from '../chart/useDrawings.js'
import ChartToolbar from './ChartToolbar.vue'
import ChartLegend from './ChartLegend.vue'

const props = defineProps({
  type: { type: String, default: 'line' }, // 'line' or 'candles'
  points: { type: Array, default: () => [] }, // line: [{ time, price, volume }]
  candles: { type: Array, default: () => [] }, // candles: [{ time, open, high, low, close, volume }]
  fitKey: { type: String, default: '' }, // the view is fitted again only when this changes (token, range, type)
  token: { type: String, default: '' }, // drawings are cleared when the token changes
})

const BULL = '#16c784'
const BEAR = '#ea3943'
const MUTED = '#58667e'
// Candles use TradingView's colours.
const UP = '#089981'
const DOWN = '#f23645'
// rgba turns a colour like #16c784 into rgba() with transparency a (0 to 1).
const rgba = (hex, a) => {
  const n = parseInt(hex.slice(1), 16)
  return `rgba(${n >> 16},${(n >> 8) & 255},${n & 255},${a})`
}

const el = ref(null)
const shown = ref(null) // the bar the legend shows
const base = ref(0) // line: the first price, for the legend's change

let chart, area, candleSeries, volume, lastLine
let rows = [] // line points
let bars = [] // candles
let lastIdx = null
let barList = [] // [{ time, values }] for the drawing tools' magnet
let fitted = null // the fitKey the view was last fitted for

// isCandles reports whether the chart shows candles instead of a line.
const isCandles = () => props.type === 'candles'

const draw = useDrawings({ chart: () => chart, series: () => (isCandles() ? candleSeries : area), bars: () => barList })
const { tool, magnet, hidden, count, textBox } = draw
// vFocus puts the cursor in the text box as soon as it appears.
const vFocus = { mounted: (el) => el.focus() }
// onKey: Esc stops the drawing being placed.
const onKey = (e) => e.key === 'Escape' && draw.cancel()

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

// load puts the current points or candles into the chart and fits the view.
function load() {
  if (!chart) return
  lastIdx = null
  if (isCandles()) {
    rows = []
    bars = props.candles.filter((c) => c.close > 0)
    area.setData([])
    lastLine.applyOptions({ axisLabelVisible: false })
    candleSeries.setData(bars.map((c) => ({ time: c.time, open: c.open, high: c.high, low: c.low, close: c.close })))
    volume.setData(bars.map((c) => ({ time: c.time, value: c.volume, color: rgba(c.close >= c.open ? UP : DOWN, 0.5) })))
    chart.applyOptions({ crosshair: { mode: CrosshairMode.Normal, vertLine: { color: MUTED }, horzLine: { visible: true, labelVisible: true } } })
  } else {
    bars = []
    rows = props.points.filter((p) => p.price > 0)
    candleSeries.setData([])
    volume.setData(rows.map((p) => ({ time: p.time, value: p.volume, color: '#e3e6eb' })))
    chart.applyOptions({ crosshair: { mode: CrosshairMode.Magnet, horzLine: { visible: false, labelVisible: false } } })
    if (rows.length) paint(-1)
    else area.setData([])
  }
  const list = isCandles() ? bars : rows
  shown.value = list[list.length - 1] || null
  base.value = rows[0]?.price || 0
  barList = isCandles()
    ? bars.map((c) => ({ time: c.time, values: [c.open, c.high, c.low, c.close] }))
    : rows.map((p) => ({ time: p.time, values: [p.price] }))
  draw.attach()
  draw.setBars(barList.map((b) => b.time))
  // Keep the visitor's zoom when the data refreshes each minute.
  // Candles start with the latest 60 shown, wide enough to read, like
  // TradingView; drag left for older ones.
  if (fitted !== props.fitKey && barList.length) {
    const n = bars.length
    if (n > 60) chart.timeScale().setVisibleLogicalRange({ from: n - 60, to: n - 1 })
    else chart.timeScale().fitContent()
    fitted = props.fitKey
  }
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
      vertLine: { style: LineStyle.Dashed, width: 1, labelBackgroundColor: '#131722' },
      horzLine: { style: LineStyle.Dashed, width: 1, color: MUTED, labelBackgroundColor: '#131722', visible: false, labelVisible: false },
    },
    // The date under the mouse, on the time axis: "Tue 13 Oct, 10:00".
    localization: {
      timeFormatter: (t) => new Date(t * 1000).toLocaleString('en-GB', { weekday: 'short', day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit', hour12: false }),
    },
    handleScroll: SCROLL,
    handleScale: SCALE,
  })

  area = chart.addSeries(AreaSeries, {
    lineWidth: 2, priceLineVisible: false, lastValueVisible: false,
    crosshairMarkerRadius: 5, crosshairMarkerBorderWidth: 3, priceFormat,
  })
  candleSeries = chart.addSeries(CandlestickSeries, {
    upColor: UP, downColor: DOWN, borderUpColor: UP, borderDownColor: DOWN, wickUpColor: UP, wickDownColor: DOWN,
    priceLineVisible: false, lastValueVisible: true, priceFormat,
  })
  for (const s of [area, candleSeries]) s.priceScale().applyOptions({ scaleMargins: { top: 0.08, bottom: 0.22 } })
  volume = chart.addSeries(HistogramSeries, { priceScaleId: 'vol', priceLineVisible: false, lastValueVisible: false })
  chart.priceScale('vol').applyOptions({ scaleMargins: { top: 0.84, bottom: 0 }, visible: false })
  lastLine = area.createPriceLine({ price: 0, lineVisible: false, axisLabelVisible: false, title: '' })

  chart.subscribeCrosshairMove((param) => {
    const list = isCandles() ? bars : rows
    if (!list.length) return
    if (!param.point || param.logical == null || param.point.x < 0 || param.point.y < 0) {
      shown.value = list[list.length - 1]
      paint(-1)
      return
    }
    const idx = Math.max(0, Math.min(list.length - 1, Math.round(param.logical)))
    shown.value = list[idx]
    if (!draw.busy.value) paint(idx)
  })

  draw.install(chart, el.value)
  window.addEventListener('keydown', onKey)
  load()
})

watch(() => [props.type, props.points, props.candles], load)
watch(() => props.token, () => draw.clearAll())

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKey)
  chart && chart.remove()
})
</script>

<template>
  <div class="frame">
    <ChartToolbar v-model:tool="tool" v-model:magnet="magnet" v-model:hidden="hidden" :count="count"
                  @zoom="draw.zoom" @reset="draw.resetView" @clear="draw.clearAll" />
    <div class="wrap">
      <div ref="el" class="chart"></div>
      <input v-if="textBox" v-model="textBox.value" v-focus class="text-input" :style="{ left: textBox.x + 'px', top: textBox.y + 'px' }"
             maxlength="80" aria-label="Text on the chart" @keydown.enter="draw.submitText" @keydown.esc.stop="draw.cancel" @blur="draw.submitText" />
      <ChartLegend :type="type" :bar="shown" :base="base" />
    </div>
  </div>
</template>

<style scoped>
.frame { display: flex; border: 1px solid var(--line); border-radius: 12px; overflow: hidden; }
.wrap { position: relative; flex: 1; min-width: 0; }
.text-input { position: absolute; z-index: 6; transform: translateY(-50%); width: 180px; font: 600 14px Inter, system-ui, sans-serif; color: var(--fg); padding: 4px 8px; border: 1px solid #2962ff; border-radius: 6px; outline: none; background: #fff; }
@media (max-width: 700px) { .frame { flex-direction: column; } }
.chart { height: 480px; }
/* The TradingView logo in light grey. */
.chart :deep(a#tv-attr-logo) { --fill: #c5ccd8; }
@media (max-width: 700px) { .chart { height: 320px; } }
</style>
