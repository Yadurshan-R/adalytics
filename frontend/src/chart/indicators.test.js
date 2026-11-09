// Tests for indicators.js: the moving average, RSI and ATR give the textbook
// values on small hand-checked examples.
import { describe, it, expect } from 'vitest'
import { INDICATORS, sma, rsi, atr, trueRange, minBars } from './indicators.js'

// bar makes one bar at time t (open is not used by these indicators).
const bar = (t, high, low, close) => ({ time: t, open: close, high, low, close })

describe('INDICATORS', () => {
  it('lists the three indicators traders use: MA 20, MA 50, RSI 14, ATR 14', () => {
    expect(INDICATORS.map((i) => i.name)).toEqual(['MA 20', 'MA 50', 'RSI 14', 'ATR 14'])
  })
})

describe('sma', () => {
  it('averages the last n closes, from the nth bar on', () => {
    const bars = [1, 2, 3, 4, 5].map((c, i) => bar(i, c, c, c))
    expect(sma(bars, 3)).toEqual([{ time: 2, value: 2 }, { time: 3, value: 3 }, { time: 4, value: 4 }])
  })
})

describe('rsi', () => {
  it('is 100 when every move was up, 0 when every move was down', () => {
    const up = [1, 2, 3, 4, 5, 6].map((c, i) => bar(i, c, c, c))
    const down = [6, 5, 4, 3, 2, 1].map((c, i) => bar(i, c, c, c))
    expect(rsi(up, 3).map((p) => p.value)).toEqual([100, 100, 100])
    expect(rsi(down, 3).map((p) => p.value)).toEqual([0, 0, 0])
  })

  it('is 50 when the ups and downs are the same size', () => {
    const bars = [10, 11, 10, 11, 10].map((c, i) => bar(i, c, c, c))
    expect(rsi(bars, 2)[0]).toEqual({ time: 2, value: 50 })
  })
})

describe('trueRange', () => {
  it('counts a gap from the previous close', () => {
    expect(trueRange(bar(1, 12, 11, 11.5), 10)).toBe(2) // gap up: high − previous close
    expect(trueRange(bar(1, 9, 8, 8.5), 10)).toBe(2) // gap down: previous close − low
    expect(trueRange(bar(1, 12, 9, 10), 10)).toBe(3) // no gap: high − low
    expect(trueRange(bar(0, 12, 9, 10), null)).toBe(3) // the first bar
  })
})

describe('atr', () => {
  it('averages the first n true ranges, then smooths each new one by 1/n', () => {
    const bars = [bar(0, 11, 9, 10), bar(1, 12, 10, 11), bar(2, 13, 11, 12), bar(3, 18, 12, 16)]
    // true ranges: 2, 2, 2, 6 → ATR 2 at bar 2, then (2 × 2 + 6) / 3 = 3.333 at bar 3
    const out = atr(bars, 3)
    expect(out[0]).toEqual({ time: 2, value: 2 })
    expect(out[1].time).toBe(3)
    expect(out[1].value).toBeCloseTo(10 / 3, 10)
  })

  it('is 0 for bars where the price never moved (a token nobody traded)', () => {
    const flat = [1, 2, 3, 4].map((t) => bar(t, 5, 5, 5))
    expect(atr(flat, 3).every((p) => p.value === 0)).toBe(true)
  })
})

describe('too few bars', () => {
  // A 1W chart has about 5 candles in 30 days, a 1D chart 30: fewer than some
  // indicators need. They must give no value at all, not a wrong one or NaN.
  const five = [1, 2, 3, 4, 5].map((c, i) => bar(i, c + 1, c - 1, c))
  it('gives no value with fewer bars than needed', () => {
    expect(sma(five, 20)).toEqual([])
    expect(rsi(five, 14)).toEqual([])
    expect(atr(five, 14)).toEqual([])
    expect(sma([], 20)).toEqual([])
    expect(rsi([], 14)).toEqual([])
  })

  it('gives the first value exactly at minBars', () => {
    const many = Array.from({ length: 60 }, (_, i) => bar(i, i + 2, i, i + 1))
    for (const ind of INDICATORS) {
      const f = { ma20: sma, ma50: sma, rsi, atr }[ind.key]
      expect(f(many.slice(0, minBars(ind) - 1), ind.n), ind.name).toEqual([])
      expect(f(many.slice(0, minBars(ind)), ind.n), ind.name).toHaveLength(1)
    }
  })
})
