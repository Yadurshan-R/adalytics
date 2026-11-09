// The chart library works in UTC: it starts a new day on its time axis at
// midnight UTC, which is 05:30 in Sri Lanka, so "9 Nov" would sit at 05:30.
// As the library's docs advise (Time zones), intraday candles (5m to 12h) are
// moved by the visitor's offset from UTC before they are drawn, and the axis
// then writes them as UTC: the new day starts at 00:00 local time, like
// TradingView. 1D, 1W and 1M candles are UTC days, like Birdeye's and
// Binance's, so they are not moved and show only their date: moved, the
// 9 Nov candle would read 8 Nov for a visitor behind UTC (America).

// DAY_SIZES are the candle sizes that are whole UTC days.
const DAY_SIZES = ['1D', '1W', '1M']
export const isDaySize = (size) => DAY_SIZES.includes(size)

// localShift is the seconds to add to a UTC time to get the visitor's clock.
// One offset for the whole chart (today's), so times stay strictly in order
// even across a daylight saving change, which the library requires.
export const localShift = (now = new Date()) => -now.getTimezoneOffset() * 60

// chartShift is the shift for a candle size: none for 1D, 1W and 1M.
export const chartShift = (size, now) => (isDaySize(size) ? 0 : localShift(now))

// toLocal moves candles by a shift (from chartShift).
export const toLocal = (candles, shift) => candles.map((c) => ({ ...c, time: c.time + shift }))

// utc writes time t (seconds) with Intl options, read as UTC (see above).
const utc = (t, opts, locale = 'en-GB') => new Date(t * 1000).toLocaleString(locale, { ...opts, timeZone: 'UTC' })

// tickMarkFor returns the time axis writer for a size: hours like "16:00";
// a new day as "5 Oct", like CoinGecko; a new month or year as its name.
// 1M writes each candle as its month ("Nov"), and no size writes a clock
// time on a day candle.
export function tickMarkFor(size) {
  return (t, type) => {
    if (type === 0) return utc(t, { year: 'numeric' })
    if (type === 1 || size === '1M') return utc(t, { month: 'short' }, 'en-US')
    if (type === 2 || isDaySize(size)) return utc(t, { day: 'numeric', month: 'short' })
    return utc(t, { hour: '2-digit', minute: '2-digit', hour12: false })
  }
}

// crosshairFor returns how the date under the mouse is written for a size:
// "Tue 13 Oct, 10:00"; "Mon 9 Nov" (1D); "Week of 2 Nov" (1W); "November 2026" (1M).
export function crosshairFor(size) {
  if (size === '1M') return (t) => utc(t, { month: 'long', year: 'numeric' })
  if (size === '1W') return (t) => `Week of ${utc(t, { day: 'numeric', month: 'short' })}`
  if (size === '1D') return (t) => utc(t, { weekday: 'short', day: 'numeric', month: 'short' }).replace(',', '')
  return (t) => utc(t, { weekday: 'short', day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit', hour12: false })
}
