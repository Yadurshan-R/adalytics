// While the backend still loads the history, a big candle size would show
// only one or two candles, which looks broken. loadWait tells the chart when
// the picked size has enough history to look like a real chart (about 20
// candles, as Birdeye and TradingView open with), so until then it can show
// "Loading 1 day candles 35%". After that, older candles fill in by themselves.

// MIN_CANDLES is how many candles a size needs before it is shown.
export const MIN_CANDLES = 20

// SIZES gives each candle size its length in hours, how far back it goes in
// days (as CandleSizes in the backend's internal/market/candles.go), and its
// name in words. 1W and 1M wait for their whole span: 30 days hold only a
// few of them.
export const SIZES = {
  '5m': { hours: 5 / 60, days: 2, name: '5 minute' },
  '15m': { hours: 0.25, days: 7, name: '15 minute' },
  '30m': { hours: 0.5, days: 7, name: '30 minute' },
  '1h': { hours: 1, days: 30, name: '1 hour' },
  '2h': { hours: 2, days: 30, name: '2 hour' },
  '4h': { hours: 4, days: 30, name: '4 hour' },
  '12h': { hours: 12, days: 30, name: '12 hour' },
  '1D': { hours: 24, days: 30, name: '1 day' },
  '1W': { hours: Infinity, days: 30, name: '1 week' },
  '1M': { hours: Infinity, days: 30, name: '1 month' },
}

// needHours is how many hours of history the size needs before it is shown:
// 20 candles, but never more than its span or the history kept.
export function needHours(size, historyDays) {
  const s = SIZES[size] ?? { hours: Infinity, days: 30 }
  return Math.min(MIN_CANDLES * s.hours, s.days * 24, historyDays * 24)
}

// loadWait returns { pct } while the history is shorter than the size needs:
// pct is how much of that is loaded (0 to 99). It returns null once there is
// enough (or the whole history is in), so the chart shows.
export function loadWait(size, history) {
  if (!history || history.done || !history.days) return null
  const need = needHours(size, history.days)
  const have = history.loaded_hours || 0
  if (have >= need) return null
  return { pct: Math.min(99, Math.floor((have / need) * 100)) }
}
