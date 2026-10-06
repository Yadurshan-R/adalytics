<!--
  CoinGecko-style price chart (TradingView Lightweight Charts).
  - Area line, green if the price is up over the range, red if down.
  - On hover: dashed line + dot; the line stays bold up to the cursor and fades
    after it; the colour compares the hovered price with the first price.
  - Grey volume bars underneath, tooltip with time, price and volume.
-->
<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { createChart, AreaSeries, HistogramSeries, LineStyle, CrosshairMode } from 'lightweight-charts'
import { formatPrice, formatADA, formatDateTime } from '../format.js'

const props = defineProps({
  points: { type: Array, default: () => [] }, // [{ time, price, volume }]
})

const BULL = '#16c784'
const BEAR = '#ea3943'
const rgba = (hex, a) => {
  const n = parseInt(hex.slice(1), 16)
  return `rgba(${n >> 16},${(n >> 8) & 255},${n & 255},${a})`
}

const el = ref(null)
const tip = ref({ show: false, x: 0, y: 0, time: 0, price: 0, volume: 0 })

let chart, area, volume, lastLine
let rows = []
let lastIdx = null

function paint(idx) {
  if (!rows.length || idx === lastIdx) return
  lastIdx = idx
  const first = rows[0].price
  const shown = idx < 0 ? rows[rows.length - 1].price : rows[idx].price
  const col = shown >= first ? BULL : BEAR
  const solid = { lineColor: col, topColor: rgba(col, 0.24), bottomColor: rgba(col, 0.02) }
  const faded = { lineColor: rgba(col, 0.3), topColor: rgba(col, 0.06), bottomColor: rgba(col, 0) }
  area.setData(rows.map((p, k) => ({ time: p.time, value: p.price, ...(idx < 0 || k <= idx ? solid : faded) })))
  area.applyOptions({ crosshairMarkerBackgroundColor: col, crosshairMarkerBorderColor: rgba(col, 0.25) })
  chart.applyOptions({ crosshair: { vertLine: { color: col } } })
  lastLine.applyOptions({ price: rows[rows.length - 1].price, color: col, axisLabelColor: col, axisLabelTextColor: '#ffffff' })
}

function load(points) {
  rows = points.filter((p) => p.price > 0)
  lastIdx = null
  if (!chart) return
  volume.setData(rows.map((p) => ({ time: p.time, value: p.volume, color: '#e3e6eb' })))
  if (rows.length) {
    paint(tip.value.show ? Math.min(tip.value.idx ?? -1, rows.length - 1) : -1)
    chart.timeScale().fitContent()
  } else {
    area.setData([])
  }
}

onMounted(() => {
  chart = createChart(el.value, {
    autoSize: true,
    layout: { background: { type: 'solid', color: '#ffffff' }, textColor: '#58667e', fontFamily: 'Inter, system-ui, sans-serif', fontSize: 12 },
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
    crosshairMarkerRadius: 5, crosshairMarkerBorderWidth: 3,
    priceFormat: { type: 'custom', minMove: 1e-12, formatter: formatPrice },
  })
  area.priceScale().applyOptions({ scaleMargins: { top: 0.08, bottom: 0.22 } })
  volume = chart.addSeries(HistogramSeries, { priceScaleId: 'vol', priceLineVisible: false, lastValueVisible: false })
  chart.priceScale('vol').applyOptions({ scaleMargins: { top: 0.84, bottom: 0 }, visible: false })
  lastLine = area.createPriceLine({ price: 0, lineVisible: false, axisLabelVisible: true, title: '' })

  chart.subscribeCrosshairMove((param) => {
    if (!rows.length || !param.point || param.logical == null || param.point.x < 0 || param.point.y < 0) {
      tip.value = { ...tip.value, show: false }
      paint(-1)
      return
    }
    const idx = Math.max(0, Math.min(rows.length - 1, Math.round(param.logical)))
    paint(idx)
    const p = rows[idx]
    const w = el.value.clientWidth
    const tw = 290 // tooltip width
    const x = param.point.x + 24 + tw > w - 70 ? param.point.x - 24 - tw : param.point.x + 24
    const y = Math.max(8, Math.min(el.value.clientHeight - 130, param.point.y - 50))
    tip.value = { show: true, idx, x: Math.max(0, x), y, time: p.time, price: p.price, volume: p.volume }
  })

  load(props.points)
})

watch(() => props.points, load)

onBeforeUnmount(() => chart && chart.remove())
</script>

<template>
  <div class="wrap">
    <div ref="el" class="chart"></div>
    <div v-show="tip.show" class="tip" :style="{ left: tip.x + 'px', top: tip.y + 'px' }">
      <div class="t">{{ formatDateTime(tip.time) }}</div>
      <div><span class="k">Price:</span> <span class="v num">{{ formatPrice(tip.price) }}</span></div>
      <div><span class="k">Vol:</span> <span class="v num">{{ formatADA(tip.volume) }}</span></div>
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
</style>
