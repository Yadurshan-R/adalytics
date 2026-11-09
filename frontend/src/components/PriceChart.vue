<!--
  The price chart (TradingView Lightweight Charts), as candlesticks or a line.
  Both draw the same candles: the line joins each candle's close.
  - Candles: green when a candle closed higher than it opened, red when lower.
  - Line: a plain blue line with nothing filled under it, like CoinMarketCap's.
  - The line at the top left (ChartLegend) shows the bar under the mouse.
  - Volume bars underneath: green or red under candles, light grey under the line.
  - A drawing toolbar on the left (DrawToolbar), like Birdeye's; the drawings
    themselves come from lightweight-charts-drawing (see useDrawTools).
-->
<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { createChart, AreaSeries, CandlestickSeries, HistogramSeries, LineStyle } from 'lightweight-charts'
import { formatPrice } from '../format.js'
import { theme } from '../theme.js'
import { chartColors, chartOptions, themeOptions, rgba } from '../chart/chartOptions.js'
import { useDrawTools } from '../chart/useDrawTools.js'
import { useIndicators } from '../chart/useIndicators.js'
import { zoom, resetView } from '../chart/view.js'
import { priceStep } from '../chart/priceStep.js'
import { indicators as picked } from '../chartPrefs.js'
import DrawToolbar from './DrawToolbar.vue'
import ChartTextBox from './ChartTextBox.vue'
import ChartLegend from './ChartLegend.vue'

const props = defineProps({
  type: { type: String, default: 'candles' }, // 'candles' or 'line'
  candles: { type: Array, default: () => [] }, // [{ time, open, high, low, close, volume }], oldest first
  indicators: { type: Array, default: () => [] }, // picked indicators, e.g. ['ma20', 'atr']
  fmt: { type: Function, default: formatPrice }, // writes a value: a price, or a compact FDV
  fitKey: { type: String, default: '' }, // the view is fitted again only when this changes (token, size, unit, range)
  span: { type: Number, default: 0 }, // seconds to show when a time range is picked; 0 = the latest candles
  token: { type: String, default: '' }, // drawings are cleared when the token changes
})

let C = {} // the chart colours of the current theme (chartColors)

const el = ref(null)
const shown = ref(null) // the bar the legend shows
const base = ref(0) // line: the close before the bar shown, for the legend's change

let chart, area, candleSeries, volume
let bars = [] // the candles on the chart
let fitted = null // the fitKey the view was last fitted for

// isCandles reports whether the chart shows candles instead of a line.
const isCandles = () => props.type === 'candles'

const draw = useDrawTools({ chart: () => chart, series: () => (isCandles() ? candleSeries : area), bars: () => bars })
const { tool, magnet, stay, locked, hidden, count, selection, textBox } = draw
const hiddenInd = ref(false) // the toolbar's "Hide indicators"

// hide switches the drawings or the indicators between hidden and shown;
// 'all' hides both, or shows both when both are hidden.
function hide(which) {
  const both = !(hidden.value && hiddenInd.value)
  const drawings = which === 'all' ? both : which === 'drawings' ? !hidden.value : hidden.value
  const inds = which === 'all' ? both : which === 'indicators' ? !hiddenInd.value : hiddenInd.value
  draw.setHidden(drawings)
  hiddenInd.value = inds
  ind.show(!inds)
}

// remove removes the drawings, the indicators (unpicks them) or both.
function remove(which) {
  if (which !== 'indicators') draw.removeAll()
  if (which !== 'drawings') picked.value = []
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
  volume.setData(bars.map((c) => ({ time: c.time, value: c.volume, color: rgba(c.close >= c.open ? C.up : C.down, 0.5) })))
  showBar(-1)
  // The price step fits the token's price, so the drawings count sensible ticks.
  priceFormat.minMove = priceStep(bars.reduce((top, c) => Math.max(top, c.high), 0))
  for (const s of [candleSeries, area]) s.applyOptions({ priceFormat: { ...priceFormat } })
  draw.attach()
  ind.set(props.indicators, bars)
  // Keep the visitor's zoom when the data refreshes each minute.
  // Candles start with the latest ones shown, about 14px each so they are wide
  // enough to read (about 60 on a laptop), like TradingView; the line shows
  // about four times as many, so it reads as a trend. Drag for older ones.
  if (fitted !== props.fitKey && bars.length) {
    const n = bars.length
    const show = Math.max(20, Math.round(el.value.clientWidth / (line ? 3.5 : 14)))
    const last = bars[n - 1].time
    if (props.span) chart.timeScale().setVisibleRange({ from: Math.max(bars[0].time, last - props.span), to: last })
    else if (n > show) chart.timeScale().setVisibleLogicalRange({ from: n - show, to: n - 1 })
    else chart.timeScale().fitContent()
    fitted = props.fitKey
  }
}

// The axis writes values the way the fmt prop says (a price or an FDV).
const priceFormat = { type: 'custom', minMove: 1e-12, formatter: (v) => props.fmt(v) }
const ind = useIndicators(() => chart, priceFormat)

// applyTheme repaints the chart in the current theme's colours.
function applyTheme() {
  if (!chart) return
  C = chartColors()
  chart.applyOptions(themeOptions(C))
  area.applyOptions({ lineColor: C.line, priceLineColor: C.line })
  candleSeries.applyOptions({ upColor: C.up, downColor: C.down, borderUpColor: C.up, borderDownColor: C.down, wickUpColor: C.up, wickDownColor: C.down })
  load()
}

onMounted(() => {
  C = chartColors()
  chart = createChart(el.value, chartOptions(C))

  area = chart.addSeries(AreaSeries, {
    lineColor: C.line, topColor: 'rgba(0,0,0,0)', bottomColor: 'rgba(0,0,0,0)', lineWidth: 2, // just the line, no fill under it
    priceLineVisible: false, priceLineStyle: LineStyle.Dotted, priceLineColor: C.line, lastValueVisible: false, crosshairMarkerRadius: 4, priceFormat,
  })
  candleSeries = chart.addSeries(CandlestickSeries, {
    upColor: C.up, downColor: C.down, borderUpColor: C.up, borderDownColor: C.down, wickUpColor: C.up, wickDownColor: C.down,
    priceLineVisible: true, priceLineStyle: LineStyle.Dotted, lastValueVisible: true, priceFormat,
  })
  for (const s of [area, candleSeries]) s.priceScale().applyOptions({ scaleMargins: { top: 0.08, bottom: 0.28 } })
  volume = chart.addSeries(HistogramSeries, { priceScaleId: 'vol', priceLineVisible: false, lastValueVisible: false })
  chart.priceScale('vol').applyOptions({ scaleMargins: { top: 0.76, bottom: 0 }, visible: false })

  chart.subscribeCrosshairMove((param) => {
    if (!param.point || param.logical == null || param.point.x < 0 || param.point.y < 0) return showBar(-1)
    showBar(Math.max(0, Math.min(bars.length - 1, Math.round(param.logical))))
  })

  draw.install(el.value)
  load()
})

watch(() => [props.type, props.candles, props.indicators], load)
watch(() => props.token, () => draw.removeAll())
// The light / dark switch repaints the chart; the drawings follow on the next redraw.
watch(theme, applyTheme)

onBeforeUnmount(() => {
  draw.uninstall()
  chart && chart.remove()
})
</script>

<template>
  <div class="frame">
    <DrawToolbar :tool="tool" :magnet="magnet" :stay="stay" :locked="locked" :hidden="hidden" :hidden-ind="hiddenInd"
                 :count="count" :ind-count="indicators.length" @pick="draw.pick" @arm="draw.pick($event, false)"
                 @magnet="draw.setMagnet" @stay="draw.setStay" @lock="draw.setLocked" @hide="hide" @remove="remove"
                 @zoom="zoom(chart, $event)" @reset="resetView(chart)" />
    <div class="wrap">
      <div ref="el" class="chart"></div>
      <ChartTextBox v-if="textBox" :key="textBox.id" :box="textBox" @submit="draw.submitText" @cancel="draw.cancelText" />
      <ChartLegend :type="type" :bar="shown" :base="base" :fmt="fmt" :extras="shown ? ind.at(shown.time) : []" />
      <button v-if="selection.length" type="button" class="remove" title="Delete drawing (Del)" @click="draw.removeSelected">
        <svg viewBox="0 0 20 20" aria-hidden="true"><path d="M3.5 5.5h13M8 5.5V3.5h4v2M5.5 5.5l1 11h7l1-11M8.5 8.5v5M11.5 8.5v5" /></svg>
        Delete
      </button>
    </div>
  </div>
</template>

<style scoped>
.frame { display: flex; height: 470px; border-radius: 0 0 12px 12px; overflow: hidden; } /* a little smaller, like Birdeye */
.wrap { position: relative; flex: 1; min-width: 0; }
@media (max-width: 700px) { .frame { flex-direction: column; height: auto; } }
.remove { position: absolute; z-index: 6; top: 40px; left: 50%; transform: translateX(-50%); display: inline-flex; align-items: center; gap: 6px; padding: 6px 12px; border: 0; border-radius: 6px; background: var(--popup); color: var(--text); font: 500 13px Inter, system-ui, sans-serif; cursor: pointer; box-shadow: var(--popup-shadow); }
.remove:hover { background: var(--btn-hover); color: var(--chart-down); }
.remove svg { width: 18px; height: 18px; fill: none; stroke: currentColor; stroke-width: 1.3; stroke-linecap: round; stroke-linejoin: round; }
.chart { height: 100%; }
/* The TradingView logo in light grey. */
.chart :deep(a#tv-attr-logo) { --fill: var(--icon-off); }
@media (max-width: 700px) { .chart { height: 320px; } }
</style>
