// useIndicators puts the indicators picked in the Indicators menu on the chart:
// the price lines (moving averages, average, median, weighted close) on the
// price, momentum and percent change each in their own small panel under the
// chart. The maths is TradingView's (indicators.js). It also gives the legend
// each indicator's value at the bar under the mouse.
import { ref } from 'vue'
import { LineSeries, LineStyle } from 'lightweight-charts'
import { INDICATORS, lineOf } from './indicators.js'

// panelHeight is an indicator panel's height in pixels: smaller on a phone.
const panelHeight = () => (window.innerWidth <= 700 ? 70 : 110)

// PERCENT is how percent change is written on its axis: two decimals.
const PERCENT = { type: 'custom', minMove: 0.01, formatter: (v) => `${v.toFixed(2)}%` }

// ZERO is the dotted line at 0 in the momentum and percent change panels:
// above it the price went up, below it down.
const ZERO = { price: 0, color: 'rgba(120, 123, 134, 0.5)', lineStyle: LineStyle.Dotted, lineWidth: 1, axisLabelVisible: false }

// signedOf writes a price move with its sign (+₳0.0012, −₳0.0012), for
// momentum, using the way the price itself is written (priceFormat).
const signedOf = (priceFormat) => ({
  type: 'custom', minMove: 1e-12,
  formatter: (v) => (v > 0 ? '+' : v < 0 ? '−' : '') + (v ? priceFormat.formatter(Math.abs(v)) : '0'),
})

// getChart() returns the chart. Prices use priceFormat, like the price itself.
export function useIndicators(getChart, priceFormat) {
  const series = {} // indicator key → its line on the chart
  const lookup = ref({}) // indicator key → { values: { time: value } } or { need } for the legend
  const shown = ref(true) // false while the toolbar hides the indicators
  const FORMATS = { price: priceFormat, signed: signedOf(priceFormat), percent: PERCENT }

  // create adds an indicator's line: on the price (panel 0), or in a new
  // panel under the others. The chart removes a panel by itself when its
  // line goes, and moves the panels below it up.
  function create(ind) {
    const chart = getChart()
    const panel = ind.panel ? chart.panes().length : 0
    const s = chart.addSeries(LineSeries, {
      color: ind.color, lineWidth: 1.5, priceLineVisible: false, lastValueVisible: false,
      crosshairMarkerVisible: false, priceFormat: FORMATS[ind.value], visible: shown.value,
    }, panel)
    if (ind.panel) {
      s.createPriceLine(ZERO)
      chart.panes()[panel]?.setHeight(panelHeight())
    }
    return s
  }

  // set shows the picked indicators (keys) for these bars and removes the rest.
  // One picked but with too few bars for any value gets no line and no empty
  // panel; the legend says how many candles it needs instead.
  function set(keys, bars) {
    const chart = getChart()
    if (!chart) return
    const next = {}
    for (const ind of INDICATORS) {
      const data = keys.includes(ind.key) && bars.length ? lineOf(ind, bars) : null
      if (!data?.length) {
        if (series[ind.key]) chart.removeSeries(series[ind.key])
        delete series[ind.key]
        if (data) next[ind.key] = { need: ind.need }
        continue
      }
      series[ind.key] ??= create(ind)
      series[ind.key].setData(data)
      next[ind.key] = { values: Object.fromEntries(data.map((p) => [p.time, p.value])) }
    }
    lookup.value = next
  }

  // at is what the legend shows for the bar at this time:
  // [{ name, color, text, need }] (text: the value written like its axis, — when
  // the bar has none; need: the candles it needs, set only when there are fewer).
  function at(time) {
    if (!shown.value) return []
    return INDICATORS.filter((ind) => lookup.value[ind.key]).map((ind) => {
      const v = lookup.value[ind.key].values?.[time]
      return { name: ind.name, color: ind.color, text: v == null ? '—' : FORMATS[ind.value].formatter(v), need: lookup.value[ind.key].need }
    })
  }

  // show hides (false) or shows (true) every indicator, lines and legend;
  // indicators added while hidden stay hidden.
  function show(on) {
    shown.value = on
    for (const s of Object.values(series)) s.applyOptions({ visible: on })
  }

  return { set, at, show }
}
