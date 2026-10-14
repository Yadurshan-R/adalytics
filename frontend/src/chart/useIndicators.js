// useIndicators puts the indicators picked in the Indicators menu on the chart:
// lines on the price (MA, EMA, Bollinger Bands, VWAP) and RSI in its own small
// panel under the chart. It also gives the legend each indicator's value at
// the bar under the mouse.
import { ref } from 'vue'
import { LineSeries, LineStyle } from 'lightweight-charts'
import { INDICATORS, sma, ema, bollinger, vwap, rsi } from './indicators.js'

// rsiHeight is the RSI panel's height in pixels: smaller on a phone.
const rsiHeight = () => (window.innerWidth <= 700 ? 70 : 110)

// getChart() returns the chart. Prices use priceFormat, like the price itself.
export function useIndicators(getChart, priceFormat) {
  const series = {} // indicator key → its lines on the chart
  const lookup = ref({}) // indicator key → { time: value } for the legend

  // line adds one line for an indicator, on the price or in pane 1 (RSI).
  function line(color, pane, opts = {}) {
    const fmt = pane ? { type: 'price', precision: 1, minMove: 0.1 } : priceFormat
    return getChart().addSeries(LineSeries, {
      color, lineWidth: 1.5, priceLineVisible: false, lastValueVisible: false,
      crosshairMarkerVisible: false, priceFormat: fmt, ...opts,
    }, pane)
  }

  // create adds the lines one indicator needs.
  function create(ind) {
    if (ind.key === 'bb') {
      const band = { color: ind.color, lineWidth: 1, lineStyle: LineStyle.Solid }
      return [line(ind.color, 0, band), line(ind.color, 0, { lineWidth: 1, lineStyle: LineStyle.Dashed }), line(ind.color, 0, band)]
    }
    if (ind.key === 'rsi') {
      const s = line(ind.color, 1)
      s.createPriceLine({ price: 70, color: '#b2b5be', lineStyle: LineStyle.Dashed, lineWidth: 1, axisLabelVisible: false })
      s.createPriceLine({ price: 30, color: '#b2b5be', lineStyle: LineStyle.Dashed, lineWidth: 1, axisLabelVisible: false })
      getChart().panes()[1]?.setHeight(rsiHeight())
      return [s]
    }
    return [line(ind.color, 0)]
  }

  // dataOf works out an indicator's lines from the bars.
  function dataOf(key, bars) {
    if (key === 'ma') return [sma(bars, 20)]
    if (key === 'ema') return [ema(bars, 50)]
    if (key === 'bb') {
      const b = bollinger(bars, 20, 2)
      return [b.upper, b.middle, b.lower]
    }
    if (key === 'vwap') return [vwap(bars)]
    return [rsi(bars, 14)]
  }

  // set shows the picked indicators (keys) for these bars and removes the rest.
  // bars are [{ time, close, volume, high?, low? }].
  function set(keys, bars) {
    const chart = getChart()
    if (!chart) return
    const next = {}
    for (const ind of INDICATORS) {
      const on = keys.includes(ind.key) && bars.length > 0
      if (!on) {
        series[ind.key]?.forEach((s) => chart.removeSeries(s))
        delete series[ind.key]
        continue
      }
      series[ind.key] ??= create(ind)
      const data = dataOf(ind.key, bars)
      data.forEach((d, i) => series[ind.key][i].setData(d))
      next[ind.key] = Object.fromEntries(data[ind.key === 'bb' ? 1 : 0].map((p) => [p.time, p.value]))
    }
    lookup.value = next
  }

  // at is what the legend shows for the bar at this time: [{ name, color, value, price }].
  function at(time) {
    return INDICATORS.filter((ind) => lookup.value[ind.key]).map((ind) => ({
      name: ind.name, color: ind.color, value: lookup.value[ind.key][time], price: ind.onPrice,
    }))
  }

  return { set, at }
}
