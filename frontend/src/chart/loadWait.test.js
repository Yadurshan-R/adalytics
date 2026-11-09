import { describe, it, expect } from 'vitest'
import { loadWait, needHours } from './loadWait.js'

const loading = (hours) => ({ days: 30, loaded_hours: hours, done: false })

describe('loadWait', () => {
  it('needs about 20 candles, never more than the span or the history', () => {
    expect(needHours('5m', 30)).toBeCloseTo(100 / 60) // 20 × 5 minutes
    expect(needHours('1h', 30)).toBe(20)
    expect(needHours('4h', 30)).toBe(80)
    expect(needHours('1D', 30)).toBe(480) // 20 days
    expect(needHours('1W', 30)).toBe(720) // the whole 30 days
    expect(needHours('1D', 7)).toBe(168) // only 7 days are kept
  })

  it('shows 1h right after the first day while bigger sizes still load', () => {
    expect(loadWait('1h', loading(24))).toBeNull()
    expect(loadWait('4h', loading(24))).toEqual({ pct: 30 }) // 24 of 80 hours
    expect(loadWait('1W', loading(360))).toEqual({ pct: 50 })
  })

  it('is null when the whole history is loaded or there is no progress yet', () => {
    expect(loadWait('1M', { days: 30, loaded_hours: 700, done: true })).toBeNull()
    expect(loadWait('1h', { days: 0, done: false })).toBeNull()
  })

  it('never says 100% while something is still missing', () => {
    expect(loadWait('1D', loading(479.9)).pct).toBe(99)
  })
})
