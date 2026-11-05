// Tests for priceStep: the price step fits the price, five significant digits.
import { describe, it, expect } from 'vitest'
import { priceStep, FALLBACK_STEP } from './priceStep.js'

describe('priceStep', () => {
  it('gives one unit of the fifth significant digit', () => {
    expect(priceStep(0.00465)).toBeCloseTo(1e-7, 20) // SNEK
    expect(priceStep(0.00000071)).toBeCloseTo(1e-11, 25) // HOSKY
    expect(priceStep(1.94)).toBeCloseTo(1e-4, 15) // LQ
    expect(priceStep(320_000_000)).toBe(10_000) // an FDV
  })

  it('counts sensible ticks for a real move', () => {
    // The move the measure tool showed as 329,707,313 ticks with the old step of 1e-12.
    expect(Math.round(0.00032971 / priceStep(0.0045))).toBe(3297)
  })

  it('falls back to a tiny step without a price', () => {
    for (const p of [0, -1, NaN, Infinity, undefined]) expect(priceStep(p)).toBe(FALLBACK_STEP)
  })
})
