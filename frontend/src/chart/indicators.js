// indicators.js works out the chart indicators from the bars on screen, in the
// browser: nothing is asked from the backend. Each takes bars as
// [{ time, close, volume, high?, low? }] and returns [{ time, value }], with no
// value for the first bars where there is not enough history yet.

// INDICATORS is the list in the Indicators menu: key, name, colour, and whether
// it sits on the price (or in its own panel under the chart, like RSI).
export const INDICATORS = [
  { key: 'ma', name: 'MA 20', color: '#f59e0b', onPrice: true },
  { key: 'ema', name: 'EMA 50', color: '#8b5cf6', onPrice: true },
  { key: 'bb', name: 'Bollinger Bands 20 2', color: '#2196f3', onPrice: true },
  { key: 'vwap', name: 'VWAP', color: '#e91e63', onPrice: true },
  { key: 'rsi', name: 'RSI 14', color: '#7e57c2', onPrice: false },
]

// sma is the simple moving average: the average close of the last n bars.
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

// ema is the exponential moving average: like the average, but recent bars
// count more, so it turns sooner.
export function ema(bars, n) {
  const k = 2 / (n + 1)
  const out = []
  let value = null
  bars.forEach((b, i) => {
    if (i < n - 1) return
    if (value == null) value = bars.slice(0, n).reduce((s, x) => s + x.close, 0) / n
    else value = b.close * k + value * (1 - k)
    out.push({ time: b.time, value })
  })
  return out
}

// bollinger is the n-bar average with a band `width` standard deviations
// above and below it: { upper, middle, lower }, each [{ time, value }].
export function bollinger(bars, n, width) {
  const upper = []
  const middle = []
  const lower = []
  for (let i = n - 1; i < bars.length; i++) {
    const slice = bars.slice(i - n + 1, i + 1).map((b) => b.close)
    const mean = slice.reduce((s, v) => s + v, 0) / n
    const sd = Math.sqrt(slice.reduce((s, v) => s + (v - mean) ** 2, 0) / n)
    const time = bars[i].time
    upper.push({ time, value: mean + width * sd })
    middle.push({ time, value: mean })
    lower.push({ time, value: mean - width * sd })
  }
  return { upper, middle, lower }
}

// vwap is the volume weighted average price of each day so far: the average
// price paid, where big trades count more. It starts again every day (UTC).
export function vwap(bars) {
  const out = []
  let day = null
  let pv = 0
  let vol = 0
  for (const b of bars) {
    const d = Math.floor(b.time / 86400)
    if (d !== day) {
      day = d
      pv = 0
      vol = 0
    }
    const typical = b.high != null ? (b.high + b.low + b.close) / 3 : b.close
    pv += typical * (b.volume || 0)
    vol += b.volume || 0
    out.push({ time: b.time, value: vol > 0 ? pv / vol : typical })
  }
  return out
}

// rsi is the relative strength index (0 to 100, Wilder's smoothing): above 70
// the price has risen fast (overbought), below 30 it has fallen fast (oversold).
export function rsi(bars, n) {
  const out = []
  let gain = 0
  let loss = 0
  for (let i = 1; i < bars.length; i++) {
    const change = bars[i].close - bars[i - 1].close
    const up = Math.max(change, 0)
    const down = Math.max(-change, 0)
    if (i <= n) {
      gain += up / n
      loss += down / n
      if (i < n) continue
    } else {
      gain = (gain * (n - 1) + up) / n
      loss = (loss * (n - 1) + down) / n
    }
    out.push({ time: bars[i].time, value: loss === 0 ? 100 : 100 - 100 / (1 + gain / loss) })
  }
  return out
}
