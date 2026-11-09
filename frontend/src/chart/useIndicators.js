// useIndicators puts the indicators picked in the Indicators menu on the chart:
// the moving averages as lines on the price, RSI and ATR each in their own
// small panel under the chart. It also gives the legend each indicator's
// value at the bar under the mouse.
import { ref } from 'vue'
import { LineSeries, LineStyle } from 'lightweight-charts'
import { INDICATORS, sma, rsi, atr } from './indicators.js'

// panelHeight is an indicator panel's height in pixels: smaller on a phone.
const panelHeight = () => (window.innerWidth <= 700 ? 70 : 110)

// SCORE is how RSI is written on its axis: one decimal (0 to 100).
const SCORE = { type: 'price', precision: 1, minMove: 0.1 }

// RSI_LEVELS are the lines in the RSI panel, as TradingView draws them:
// overbought (70) and oversold (30) dashed, the middle (50) dotted and fainter.
const RSI_LEVELS = [
  { price: 70, color: '#787b86', lineStyle: LineStyle.Dashed },
  { price: 50, color: 'rgba(120, 123, 134, 0.5)', lineStyle: LineStyle.Dotted },
  { price: 30, color: '#787b86', lineStyle: LineStyle.Dashed },
]

// getChart() returns the chart. Prices use priceFormat, like the price itself.
export function useIndicators(getChart, priceFormat) {
  const series = {} // indicator key → its line on the chart
  const lookup = ref({}) // indicator key → { time: value } for the legend
  const shown = ref(true) // false while the toolbar hides the indicators

  // create adds an indicator's line: on the price (panel 0), or in a new
  // panel under the others. The chart removes a panel by itself when its
  // line goes, and moves the panels below it up.
  function create(ind) {
    const chart = getChart()
    const panel = ind.panel ? chart.panes().length : 0
    const s = chart.addSeries(LineSeries, {
      color: ind.color, lineWidth: 1.5, priceLineVisible: false, lastValueVisible: false,
      crosshairMarkerVisible: false, priceFormat: ind.price ? priceFormat : SCORE, visible: shown.value,
    }, panel)
    if (ind.key === 'rsi') {
      for (const level of RSI_LEVELS) s.createPriceLine({ ...level, lineWidth: 1, axisLabelVisible: false })
      // The RSI axis always shows 0 to 100, so the 70 / 50 / 30 lines stay in view.
      s.applyOptions({ autoscaleInfoProvider: () => ({ priceRange: { minValue: 0, maxValue: 100 } }) })
    }
    if (ind.panel) chart.panes()[panel]?.setHeight(panelHeight())
    return s
  }

  // dataOf works out an indicator's line from the bars.
  function dataOf(ind, bars) {
    if (ind.key === 'rsi') return rsi(bars, ind.n)
    if (ind.key === 'atr') return atr(bars, ind.n)
    return sma(bars, ind.n)
  }

  // set shows the picked indicators (keys) for these bars and removes the rest.
  function set(keys, bars) {
    const chart = getChart()
    if (!chart) return
    const next = {}
    for (const ind of INDICATORS) {
      if (!keys.includes(ind.key) || !bars.length) {
        if (series[ind.key]) chart.removeSeries(series[ind.key])
        delete series[ind.key]
        continue
      }
      series[ind.key] ??= create(ind)
      const data = dataOf(ind, bars)
      series[ind.key].setData(data)
      next[ind.key] = Object.fromEntries(data.map((p) => [p.time, p.value]))
    }
    lookup.value = next
  }

  // at is what the legend shows for the bar at this time: [{ name, color, value, price }].
  function at(time) {
    if (!shown.value) return []
    return INDICATORS.filter((ind) => lookup.value[ind.key]).map((ind) => ({
      name: ind.name, color: ind.color, value: lookup.value[ind.key][time], price: ind.price,
    }))
  }

  // show hides (false) or shows (true) every indicator, lines and legend;
  // indicators added while hidden stay hidden.
  function show(on) {
    shown.value = on
    for (const s of Object.values(series)) s.applyOptions({ visible: on })
  }

  return { set, at, show }
}
