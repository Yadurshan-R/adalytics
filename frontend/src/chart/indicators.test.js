// Tests for indicators.js: TradingView's indicator code (src/vendor/tradingview)
// gives the textbook numbers on our candles, and lineOf keeps only real values.
import { describe, it, expect } from 'vitest'
import { INDICATORS, lineOf } from './indicators.js'

// bar makes one candle at time t.
const bar = (t, open, high, low, close) => ({ time: t, open, high, low, close, volume: 1 })
// closes makes candles whose open, high and low equal the close.
const closes = (xs) => xs.map((c, i) => bar(1000 + i * 60, c, c, c, c))
const byKey = (k) => INDICATORS.find((i) => i.key === k)

describe('INDICATORS', () => {
  it('lists TradingView\'s single-token indicators, each with its maths', () => {
    expect(INDICATORS.map((i) => i.name)).toEqual([
      'MA 20', 'MA 50', 'Average price', 'Median price', 'Weighted close', 'Momentum 10', 'Percent change',
    ])
    for (const i of INDICATORS) expect(typeof i.calc).toBe('function')
  })
})

describe('lineOf with TradingView\'s maths', () => {
  it('MA 20 is the average close of the last 20 candles, from the 20th candle on', () => {
    const bars = closes(Array.from({ length: 25 }, (_, i) => i + 1)) // closes 1..25
    const line = lineOf(byKey('ma20'), bars)
    expect(line).toHaveLength(6)
    expect(line[0]).toEqual({ time: bars[19].time, value: 10.5 }) // (1 + … + 20) / 20
    expect(line[5].value).toBeCloseTo(15.5) // (6 + … + 25) / 20
  })

  it('Momentum 10 is the close now minus the close 10 candles ago', () => {
    const bars = closes([5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 20, 3])
    const line = lineOf(byKey('mom'), bars)
    expect(line.map((p) => p.value)).toEqual([15, -3]) // 20 − 5, 3 − 6
    expect(line[0].time).toBe(bars[10].time)
  })

  it('Percent change is the % move of the close from the candle before, newest candle included', () => {
    const bars = closes([100, 110, 99])
    const line = lineOf(byKey('pct'), bars)
    expect(line.map((p) => p.time)).toEqual([bars[1].time, bars[2].time])
    expect(line[0].value).toBeCloseTo(10) // 100 → 110
    expect(line[1].value).toBeCloseTo(-10) // 110 → 99
  })

  it('Average, median and weighted close use the candle\'s own prices, newest candle included', () => {
    const bars = [bar(1000, 1, 1, 1, 1), bar(1060, 10, 14, 6, 12)]
    expect(lineOf(byKey('avg'), bars).at(-1)).toEqual({ time: 1060, value: 10.5 }) // (10 + 14 + 6 + 12) / 4
    expect(lineOf(byKey('median'), bars).at(-1)).toEqual({ time: 1060, value: 10 }) // (14 + 6) / 2
    expect(lineOf(byKey('wclose'), bars).at(-1)).toEqual({ time: 1060, value: 11 }) // (14 + 6 + 2 × 12) / 4
  })

  it('gives no line, not NaN, when there are too few candles (1W, 1M)', () => {
    for (const k of ['ma20', 'ma50', 'mom']) expect(lineOf(byKey(k), closes([1, 2, 3]))).toEqual([])
    expect(lineOf(byKey('ma20'), [])).toEqual([])
  })
})
