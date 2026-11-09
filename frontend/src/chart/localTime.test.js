// Tests for localTime.js and the time axis: the new day starts at local
// midnight (00:00 in Colombo, UTC+05:30), not at midnight UTC (05:30).
import { describe, it, expect } from 'vitest'
import { localShift, toLocal, tickMark, crosshairTime } from './localTime.js'

const COLOMBO = 5.5 * 3600 // seconds ahead of UTC
const utc = (d, h, m = 0) => Date.UTC(2026, 10, d, h, m) / 1000 // November 2026

describe('local time', () => {
  it('turns the browser offset into seconds to add', () => {
    expect(localShift({ getTimezoneOffset: () => -330 })).toBe(COLOMBO) // Colombo
    expect(localShift({ getTimezoneOffset: () => 300 })).toBe(-5 * 3600) // New York
  })

  it('moves every candle by the same shift and keeps the rest', () => {
    const c = { time: utc(9, 0), open: 1, close: 2 }
    expect(toLocal([c], COLOMBO)).toEqual([{ time: utc(9, 5, 30), open: 1, close: 2 }])
  })

  it('writes the moved times as the local clock', () => {
    const [c] = toLocal([{ time: utc(9, 0, 45) }], COLOMBO) // 00:45 UTC = 06:15 in Colombo
    expect(tickMark(c.time, 4)).toBe('06:15')
    expect(crosshairTime(c.time)).toBe('Mon 9 Nov, 06:15')
  })

  it('puts the new day at local midnight', () => {
    // 18:30 UTC on 8 Nov is 00:00 on 9 Nov in Colombo: the day label goes there.
    const [c] = toLocal([{ time: utc(8, 18, 30) }], COLOMBO)
    expect(c.time).toBe(utc(9, 0)) // midnight, so the library marks a new day
    expect(tickMark(c.time, 2)).toBe('9 Nov')
  })
})
