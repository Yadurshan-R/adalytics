<!--
  The price chart (TradingView Lightweight Charts), as candlesticks or a line.
  Both draw the same candles: the line joins each candle's close.
  - Candles: green when a candle closed higher than it opened, red when lower.
  - Line: a plain blue line with nothing filled under it, like CoinMarketCap's.
  - The line at the top left (ChartLegend) shows the bar under the mouse.
  - Volume bars underneath: green or red under candles, light grey under the line.
  - A drawing toolbar on the left (ChartToolbar), with zoom and scroll.
-->
<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { createChart, AreaSeries, CandlestickSeries, HistogramSeries, LineStyle, CrosshairMode } from 'lightweight-charts'
import { formatPrice } from '../format.js'
import { useDrawings, SCROLL, SCALE } from '../chart/useDrawings.js'
import { useIndicators } from '../chart/useIndicators.js'
import ChartToolbar from './ChartToolbar.vue'
import ChartLegend from './ChartLegend.vue'

const props = defineProps({
  type: { type: String, default: 'candles' }, // 'candles' or 'line'
  candles: { type: Array, default: () => [] }, // [{ time, open, high, low, close, volume }], oldest first
  indicators: { type: Array, default: () => [] }, // picked indicators, e.g. ['ma', 'rsi']
  fmt: { type: Function, default: formatPrice }, // writes a value: a price, or a compact FDV
  fitKey: { type: String, default: '' }, // the view is fitted again only when this changes (token, size, unit)
  token: { type: String, default: '' }, // drawings are cleared when the token changes
})

const MUTED = '#58667e'
// TradingView's colours: green and red candles, a blue line.
const UP = '#089981'
const DOWN = '#f23645'
const BLUE = '#3861fb' // CoinMarketCap's chart blue
// rgba turns a colour like #16c784 into rgba() with transparency a (0 to 1).
const rgba = (hex, a) => {
  const n = parseInt(hex.slice(1), 16)
  return `rgba(${n >> 16},${(n >> 8) & 255},${n & 255},${a})`
}

const el = ref(null)
const shown = ref(null) // the bar the legend shows
const base = ref(0) // line: the close before the bar shown, for the legend's change

let chart, area, candleSeries, volume
let bars = [] // the candles on the chart
let barList = [] // [{ time, values }] for the drawing tools' magnet
let fitted = null // the fitKey the view was last fitted for

// isCandles reports whether the chart shows candles instead of a line.
const isCandles = () => props.type === 'candles'

const draw = useDrawings({ chart: () => chart, series: () => (isCandles() ? candleSeries : area), bars: () => barList })
const { tool, magnet, hidden, locked, stay, count, textBox, selected } = draw
// vFocus puts the cursor in the text box as soon as it appears.
const vFocus = { mounted: (el) => el.focus() }
// onKey: Esc stops the drawing being placed or unselects; Delete or Backspace
// removes the selected drawing (not while typing in a box).
function onKey(e) {
  if (e.key === 'Escape') {
    draw.cancel()
    draw.deselect()
  }
  if ((e.key === 'Delete' || e.key === 'Backspace') && selected.value != null && !/INPUT|TEXTAREA/.test(document.activeElement?.tagName)) {
    e.preventDefault()
    draw.removeSelected()
  }
}

// showBar puts bar i in the legend (-1: the latest bar).
function showBar(i) {
  if (!bars.length) return (shown.value = null)
  const k = i < 0 ? bars.length - 1 : i
  shown.value = bars[k]
  base.value = k > 0 ? bars[k - 1].close : 0
}

// load puts the candles into the chart, as candles or a line, and fits the view.
function load() {
  if (!chart) return
  const line = !isCandles()
  bars = props.candles.filter((c) => c.close > 0)
  candleSeries.setData(line ? [] : bars.map((c) => ({ time: c.time, open: c.open, high: c.high, low: c.low, close: c.close })))
  area.setData(line ? bars.map((c) => ({ time: c.time, value: c.close })) : [])
  // The latest price: a label on the axis and a dotted line across the chart, like Birdeye and CoinMarketCap.
  candleSeries.applyOptions({ lastValueVisible: !line, priceLineVisible: !line })
  area.applyOptions({ lastValueVisible: line, priceLineVisible: line })
  volume.setData(bars.map((c) => ({ time: c.time, value: c.volume, color: rgba(c.close >= c.open ? UP : DOWN, 0.5) })))
  showBar(-1)
  barList = bars.map((c) => ({ time: c.time, values: line ? [c.close] : [c.open, c.high, c.low, c.close] }))
  draw.attach()
  draw.setBars(barList.map((b) => b.time))
  ind.set(props.indicators, bars)
  // Keep the visitor's zoom when the data refreshes each minute.
  // Candles start with the latest ones shown, about 14px each so they are wide
  // enough to read (about 60 on a laptop), like TradingView; the line shows
  // about four times as many, so it reads as a trend. Drag for older ones.
  if (fitted !== props.fitKey && barList.length) {
    const n = bars.length
    const show = Math.max(20, Math.round(el.value.clientWidth / (line ? 3.5 : 14)))
    if (n > show) chart.timeScale().setVisibleLogicalRange({ from: n - show, to: n - 1 })
    else chart.timeScale().fitContent()
    fitted = props.fitKey
  }
}

// The axis writes values the way the fmt prop says (a price or an FDV).
const priceFormat = { type: 'custom', minMove: 1e-12, formatter: (v) => props.fmt(v) }
const ind = useIndicators(() => chart, priceFormat)

onMounted(() => {
  chart = createChart(el.value, {
    autoSize: true,
    layout: { background: { type: 'solid', color: '#ffffff' }, textColor: MUTED, fontFamily: 'Inter, system-ui, sans-serif', fontSize: 12 },
    grid: { vertLines: { color: '#f3f5f8' }, horzLines: { color: '#eff2f5' } },
    rightPriceScale: { borderVisible: false },
    timeScale: {
      borderVisible: false, fixLeftEdge: true, fixRightEdge: true,
      timeVisible: true, secondsVisible: false,
      // Hours like "16:00"; a new day shows as "5 Oct", like CoinGecko.
      tickMarkFormatter: (t, type) => {
        const d = new Date(t * 1000)
        if (type === 0) return String(d.getFullYear())
        const month = d.toLocaleString('en-US', { month: 'short' })
        if (type === 1) return month
        if (type === 2) return `${d.getDate()} ${month}`
        return d.toLocaleString('en-GB', { hour: '2-digit', minute: '2-digit', hour12: false })
      },
    },
    crosshair: {
      mode: CrosshairMode.Normal,
      vertLine: { style: LineStyle.Dashed, width: 1, color: MUTED, labelBackgroundColor: '#131722' },
      horzLine: { style: LineStyle.Dashed, width: 1, color: MUTED, labelBackgroundColor: '#131722' },
    },
    // The date under the mouse, on the time axis: "Tue 13 Oct, 10:00".
    localization: {
      timeFormatter: (t) => new Date(t * 1000).toLocaleString('en-GB', { weekday: 'short', day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit', hour12: false }),
    },
    handleScroll: SCROLL,
    handleScale: SCALE,
  })

  area = chart.addSeries(AreaSeries, {
    lineColor: BLUE, topColor: 'rgba(0,0,0,0)', bottomColor: 'rgba(0,0,0,0)', lineWidth: 2, // just the line, no fill under it
    priceLineVisible: false, priceLineStyle: LineStyle.Dotted, priceLineColor: BLUE, lastValueVisible: false, crosshairMarkerRadius: 4, priceFormat,
  })
  candleSeries = chart.addSeries(CandlestickSeries, {
    upColor: UP, downColor: DOWN, borderUpColor: UP, borderDownColor: DOWN, wickUpColor: UP, wickDownColor: DOWN,
    priceLineVisible: true, priceLineStyle: LineStyle.Dotted, lastValueVisible: true, priceFormat,
  })
  for (const s of [area, candleSeries]) s.priceScale().applyOptions({ scaleMargins: { top: 0.08, bottom: 0.28 } })
  volume = chart.addSeries(HistogramSeries, { priceScaleId: 'vol', priceLineVisible: false, lastValueVisible: false })
  chart.priceScale('vol').applyOptions({ scaleMargins: { top: 0.76, bottom: 0 }, visible: false })

  chart.subscribeCrosshairMove((param) => {
    if (!param.point || param.logical == null || param.point.x < 0 || param.point.y < 0) return showBar(-1)
    showBar(Math.max(0, Math.min(bars.length - 1, Math.round(param.logical))))
  })

  draw.install(chart, el.value)
  window.addEventListener('keydown', onKey)
  load()
})

watch(() => [props.type, props.candles, props.indicators], load)
watch(() => props.token, () => draw.clearAll())

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKey)
  draw.uninstall()
  chart && chart.remove()
})
</script>

<template>
  <div class="frame">
    <ChartToolbar v-model:tool="tool" v-model:magnet="magnet" v-model:hidden="hidden" v-model:locked="locked" v-model:stay="stay"
                  :count="count" @zoom="draw.zoom" @reset="draw.resetView" @clear="draw.clearAll" />
    <div class="wrap">
      <div ref="el" class="chart"></div>
      <input v-if="textBox" v-model="textBox.value" v-focus class="text-input" :style="{ left: textBox.x + 'px', top: textBox.y + 'px' }"
             maxlength="80" aria-label="Text on the chart" @keydown.enter="draw.submitText" @keydown.esc.stop="draw.cancel" @blur="draw.submitText" />
      <ChartLegend :type="type" :bar="shown" :base="base" :fmt="fmt" :extras="shown ? ind.at(shown.time) : []" />
      <button v-if="selected != null" type="button" class="remove" title="Delete drawing (Del)" @click="draw.removeSelected">
        <svg viewBox="0 0 20 20" aria-hidden="true"><path d="M3.5 5.5h13M8 5.5V3.5h4v2M5.5 5.5l1 11h7l1-11M8.5 8.5v5M11.5 8.5v5" /></svg>
        Delete
      </button>
    </div>
  </div>
</template>

<style scoped>
.frame { display: flex; height: 470px; border-radius: 0 0 12px 12px; overflow: hidden; } /* a little smaller, like Birdeye */
.wrap { position: relative; flex: 1; min-width: 0; }
.text-input { position: absolute; z-index: 6; transform: translateY(-50%); width: 180px; font: 600 14px Inter, system-ui, sans-serif; color: var(--fg); padding: 4px 8px; border: 1px solid #2962ff; border-radius: 6px; outline: none; background: #fff; }
@media (max-width: 700px) { .frame { flex-direction: column; height: auto; } }
.remove { position: absolute; z-index: 6; top: 40px; left: 50%; transform: translateX(-50%); display: inline-flex; align-items: center; gap: 6px; padding: 6px 12px; border: 0; border-radius: 6px; background: #fff; color: #131722; font: 500 13px Inter, system-ui, sans-serif; cursor: pointer; box-shadow: 0 2px 4px rgba(0, 0, 0, 0.12), 0 0 0 1px rgba(19, 23, 34, 0.06), 0 8px 24px rgba(0, 0, 0, 0.12); }
.remove:hover { background: #f0f3fa; color: #f23645; }
.remove svg { width: 18px; height: 18px; fill: none; stroke: currentColor; stroke-width: 1.3; stroke-linecap: round; stroke-linejoin: round; }
.chart { height: 100%; }
/* The TradingView logo in light grey. */
.chart :deep(a#tv-attr-logo) { --fill: #c5ccd8; }
@media (max-width: 700px) { .chart { height: 320px; } }
</style>
