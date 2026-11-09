// indicators.js works out the chart indicators from the bars on screen, in the
// browser: nothing is asked from the backend. Each takes bars as
// [{ time, open, high, low, close }] and returns [{ time, value }], with no
// value for the first bars where there is not enough history yet.

// INDICATORS is the list in the Indicators menu: the three most traders use.
//   key, name, colour; n, the number of bars it looks back;
//   panel: false for a line on the price, true for its own panel under the
//   chart (RSI, ATR); price: true when its value is a price (the legend
//   writes it like one), false for a score (RSI, 0 to 100).
export const INDICATORS = [
  { key: 'ma20', name: 'MA 20', color: '#f59e0b', n: 20, panel: false, price: true },
  { key: 'ma50', name: 'MA 50', color: '#ec407a', n: 50, panel: false, price: true },
  { key: 'rsi', name: 'RSI 14', color: '#7e57c2', n: 14, panel: true, price: false },
  { key: 'atr', name: 'ATR 14', color: '#00acc1', n: 14, panel: true, price: true },
]

// sma is the simple moving average: the average close of the last n bars.
// Above it the price is in an uptrend, below it in a downtrend.
export function sma(bars, n) {
  const out = []
  let sum = 0
  bars.forEach((b, i) => {
    sum += b.close
    if (i >= n) sum -= bars[i - n].close
    if (i >= n - 1) out.push({ time: b.time, value: sum / n })
  })
  return out
}

// wilder smooths values over n steps the way RSI and ATR do (J. Welles
// Wilder): the plain average of the first n, then each new value counts 1/n.
// It returns one smoothed value per input from the nth on.
function wilder(values, n) {
  const out = []
  let avg = 0
  values.forEach((v, i) => {
    if (i < n) {
      avg += v / n
      if (i === n - 1) out.push(avg)
    } else {
      avg = (avg * (n - 1) + v) / n
      out.push(avg)
    }
  })
  return out
}

// rsi is the relative strength index (0 to 100): how much of the recent
// movement was up. Above 70 the price has risen fast (overbought), below 30
// it has fallen fast (oversold).
export function rsi(bars, n) {
  const moves = bars.slice(1).map((b, i) => b.close - bars[i].close)
  const gains = wilder(moves.map((m) => Math.max(m, 0)), n)
  const losses = wilder(moves.map((m) => Math.max(-m, 0)), n)
  return gains.map((g, i) => ({
    time: bars[i + n].time,
    value: losses[i] === 0 ? 100 : 100 - 100 / (1 + g / losses[i]),
  }))
}

// trueRange is how far the price moved in one bar, gaps included: the largest
// of high − low, |high − previous close| and |low − previous close|.
export function trueRange(bar, prevClose) {
  if (prevClose == null) return bar.high - bar.low
  return Math.max(bar.high - bar.low, Math.abs(bar.high - prevClose), Math.abs(bar.low - prevClose))
}

// atr is the average true range: how much the price usually moves in one bar,
// in price. It says nothing about direction, only size: traders use it to set
// a stop-loss beyond the normal swings (for example 1.5 × ATR from the entry).
export function atr(bars, n) {
  const ranges = bars.map((b, i) => trueRange(b, i ? bars[i - 1].close : null))
  return wilder(ranges, n).map((value, i) => ({ time: bars[i + n - 1].time, value }))
}
