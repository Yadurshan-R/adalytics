// Tests for localTime.js: intraday candles on the visitor's clock (the new day
// at 00:00 in Colombo, UTC+05:30, not at 05:30), and 1D, 1W, 1M on UTC days
// written as dates only.
import { describe, it, expect } from 'vitest'
import { localShift, chartShift, toLocal, tickMarkFor, crosshairFor } from './localTime.js'

const COLOMBO = { getTimezoneOffset: () => -330 }
const NEW_YORK = { getTimezoneOffset: () => 300 }
const utc = (d, h, m = 0) => Date.UTC(2026, 10, d, h, m) / 1000 // November 2026

describe('local time', () => {
  it('turns the browser offset into seconds to add', () => {
    expect(localShift(COLOMBO)).toBe(5.5 * 3600)
    expect(localShift(NEW_YORK)).toBe(-5 * 3600)
  })

  it('moves intraday candles only; 1D, 1W and 1M stay on UTC days', () => {
    expect(chartShift('15m', COLOMBO)).toBe(5.5 * 3600)
    for (const s of ['1D', '1W', '1M']) expect(chartShift(s, NEW_YORK)).toBe(0) // 9 Nov stays 9 Nov
    const c = { time: utc(9, 0), open: 1 }
    expect(toLocal([c], 5.5 * 3600)).toEqual([{ time: utc(9, 5, 30), open: 1 }])
  })

  it('writes intraday times as the local clock, the new day at local midnight', () => {
    const [c] = toLocal([{ time: utc(9, 0, 45) }], 5.5 * 3600) // 00:45 UTC = 06:15 in Colombo
    expect(tickMarkFor('15m')(c.time, 4)).toBe('06:15')
    expect(crosshairFor('15m')(c.time)).toBe('Mon 9 Nov, 06:15')
    const [mid] = toLocal([{ time: utc(8, 18, 30) }], 5.5 * 3600) // 00:00 on 9 Nov in Colombo
    expect(mid.time).toBe(utc(9, 0))
    expect(tickMarkFor('15m')(mid.time, 2)).toBe('9 Nov')
  })

  it('writes 1D, 1W and 1M as dates, never a clock time', () => {
    const nov1 = utc(1, 0), nov2 = utc(2, 0), nov9 = utc(9, 0)
    expect(tickMarkFor('1M')(nov1, 2)).toBe('Nov')
    expect(tickMarkFor('1M')(nov1, 4)).toBe('Nov')
    expect(tickMarkFor('1W')(nov2, 4)).toBe('2 Nov')
    expect(tickMarkFor('1D')(nov9, 4)).toBe('9 Nov')
    expect(tickMarkFor('1D')(nov9, 0)).toBe('2026')
    expect(crosshairFor('1M')(nov1)).toBe('November 2026')
    expect(crosshairFor('1W')(nov2)).toBe('Week of 2 Nov')
    expect(crosshairFor('1D')(nov9)).toBe('Mon 9 Nov')
  })
})
