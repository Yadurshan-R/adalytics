// indicators.js lists the indicators in the Indicators menu. The maths is
// TradingView's own code (src/vendor/tradingview, copied unchanged from the
// Lightweight Charts repository); this file only names each indicator, gives
// it a colour and says where it goes. Nothing is asked from the backend.
import { calculateMovingAverageIndicatorValues } from '../vendor/tradingview/indicators/moving-average/moving-average-calculation.ts'
import { calculateMomentumIndicatorValues } from '../vendor/tradingview/indicators/momentum/momentum-calculation.ts'
import { calculatePercentChangeIndicatorValues } from '../vendor/tradingview/indicators/percent-change/percent-change-calculation.ts'
import { calculateAveragePriceIndicatorValues } from '../vendor/tradingview/indicators/average-price/average-price-calculation.ts'
import { calculateMedianPriceIndicatorValues } from '../vendor/tradingview/indicators/median-price/median-price-calculation.ts'
import { calculateWeightedCloseIndicatorValues } from '../vendor/tradingview/indicators/weighted-close/weighted-close-calculation.ts'

// INDICATORS is the list in the Indicators menu, in TradingView's names.
//   key, name, colour; need, the candles it needs for its first value;
//   panel: false for a line on the price, true for its own panel under the
//   chart; value: 'price' (written like a price), 'signed' (a price move,
//   up or down) or 'percent'; calc, TradingView's maths for the candles.
export const INDICATORS = [
  { key: 'ma20', name: 'MA 20', color: '#f59e0b', need: 20, panel: false, value: 'price',
    calc: (b) => calculateMovingAverageIndicatorValues(b, { length: 20, source: 'close' }) },
  { key: 'ma50', name: 'MA 50', color: '#ec407a', need: 50, panel: false, value: 'price',
    calc: (b) => calculateMovingAverageIndicatorValues(b, { length: 50, source: 'close' }) },
  { key: 'avg', name: 'Average price', color: '#26a69a', need: 1, panel: false, value: 'price',
    calc: (b) => calculateAveragePriceIndicatorValues(b, {}) },
  { key: 'median', name: 'Median price', color: '#42a5f5', need: 1, panel: false, value: 'price',
    calc: (b) => calculateMedianPriceIndicatorValues(b, {}) },
  { key: 'wclose', name: 'Weighted close', color: '#ab47bc', need: 1, panel: false, value: 'price',
    calc: (b) => calculateWeightedCloseIndicatorValues(b, {}) },
  { key: 'mom', name: 'Momentum 10', color: '#7e57c2', need: 11, panel: true, value: 'signed',
    calc: (b) => calculateMomentumIndicatorValues(b, { length: 10, source: 'close' }) },
  { key: 'pct', name: 'Percent change', color: '#00acc1', need: 2, panel: true, value: 'percent',
    calc: (b) => calculatePercentChangeIndicatorValues(b, {}) },
]

// lineOf works out an indicator's line for the candles, [{ time, value }],
// keeping only the candles that have a value (TradingView's code also returns
// the first candles, before there is enough history, with no value).
export function lineOf(ind, bars) {
  const candles = bars.map((b) => ({ time: b.time, open: b.open, high: b.high, low: b.low, close: b.close }))
  return ind.calc(candles).filter((p) => typeof p.value === 'number' && Number.isFinite(p.value))
}
