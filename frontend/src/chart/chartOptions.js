// chartOptions.js: how the price chart looks (PriceChart.vue): its colours
// from the theme, its axes and its crosshair, like TradingView's and Birdeye's.
import { LineStyle, CrosshairMode } from 'lightweight-charts'
import { color } from '../theme.js'
import { SCROLL, SCALE } from './view.js'

// chartColors reads the chart's colours from the current theme (style.css):
// TradingView's green and red candles, CoinMarketCap's blue line.
export const chartColors = () => ({
  bg: color('--chart-bg'), muted: color('--muted'), label: color('--chart-label'),
  gridV: color('--chart-grid-v'), gridH: color('--chart-grid-h'),
  up: color('--chart-up'), down: color('--chart-down'), line: color('--chart-line'),
})

// rgba turns a colour like #16c784 into rgba() with transparency a (0 to 1).
export function rgba(hex, a) {
  const n = parseInt(hex.slice(1), 16)
  return `rgba(${n >> 16},${(n >> 8) & 255},${n & 255},${a})`
}

// themeOptions are the chart options that change with light / dark (C: chartColors()).
export function themeOptions(C) {
  return {
    layout: { background: { type: 'solid', color: C.bg }, textColor: C.muted },
    grid: { vertLines: { color: C.gridV }, horzLines: { color: C.gridH } },
    crosshair: {
      vertLine: { color: C.muted, labelBackgroundColor: C.label },
      horzLine: { color: C.muted, labelBackgroundColor: C.label },
    },
  }
}

// tickMark writes the time axis: hours like "16:00"; a new day as "5 Oct",
// like CoinGecko; a new month or year as its name.
function tickMark(t, type) {
  const d = new Date(t * 1000)
  if (type === 0) return String(d.getFullYear())
  const month = d.toLocaleString('en-US', { month: 'short' })
  if (type === 1) return month
  if (type === 2) return `${d.getDate()} ${month}`
  return d.toLocaleString('en-GB', { hour: '2-digit', minute: '2-digit', hour12: false })
}

// chartOptions are the options the chart is created with (C: chartColors()).
export function chartOptions(C) {
  const theme = themeOptions(C)
  return {
    autoSize: true,
    layout: { ...theme.layout, fontFamily: 'Inter, system-ui, sans-serif', fontSize: 12 },
    grid: theme.grid,
    rightPriceScale: { borderVisible: false },
    // maxBarSpacing keeps a candle at most 40px wide, so a chart with only a
    // few candles (1W, 1M) shows them at a normal width, like Birdeye.
    timeScale: { borderVisible: false, fixLeftEdge: true, fixRightEdge: true, maxBarSpacing: 40, timeVisible: true, secondsVisible: false, tickMarkFormatter: tickMark },
    crosshair: {
      mode: CrosshairMode.Normal,
      vertLine: { ...theme.crosshair.vertLine, style: LineStyle.Dashed, width: 1 },
      horzLine: { ...theme.crosshair.horzLine, style: LineStyle.Dashed, width: 1 },
    },
    // The date under the mouse, on the time axis: "Tue 13 Oct, 10:00".
    localization: {
      timeFormatter: (t) => new Date(t * 1000).toLocaleString('en-GB', { weekday: 'short', day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit', hour12: false }),
    },
    handleScroll: SCROLL,
    handleScale: SCALE,
  }
}
