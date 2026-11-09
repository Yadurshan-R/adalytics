// The chart library works in UTC: it starts a new day on its time axis at
// midnight UTC, which is 05:30 in Sri Lanka, so "9 Nov" would sit at 05:30.
// As the library's docs advise (Time zones), the chart's times are moved by
// the visitor's offset from UTC before they are drawn, and the axis then
// writes them as UTC. So a candle at 00:00 UTC is drawn at 05:30, and the
// new day starts at 00:00 local time, like TradingView. Only the drawing
// moves: the candles themselves (and 1D, 1W, 1M on UTC days) stay the same.

// localShift is the seconds to add to a UTC time to get the visitor's clock.
// One offset for the whole chart (today's), so times stay strictly in order
// even across a daylight saving change, which the library requires.
export const localShift = (now = new Date()) => -now.getTimezoneOffset() * 60

// toLocal moves candles to the visitor's clock (shift from localShift).
export const toLocal = (candles, shift) => candles.map((c) => ({ ...c, time: c.time + shift }))

// tickMark writes the time axis: hours like "16:00"; a new day as "5 Oct",
// like CoinGecko; a new month or year as its name. The chart's times are
// already on the visitor's clock (toLocal), so they are written as UTC.
export function tickMark(t, type) {
  const d = new Date(t * 1000)
  if (type === 0) return String(d.getUTCFullYear())
  const month = d.toLocaleString('en-US', { month: 'short', timeZone: 'UTC' })
  if (type === 1) return month
  if (type === 2) return `${d.getUTCDate()} ${month}`
  return d.toLocaleString('en-GB', { hour: '2-digit', minute: '2-digit', hour12: false, timeZone: 'UTC' })
}

// crosshairTime writes the date under the mouse: "Tue 13 Oct, 10:00" (as UTC, see tickMark).
export const crosshairTime = (t) =>
  new Date(t * 1000).toLocaleString('en-GB', { weekday: 'short', day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit', hour12: false, timeZone: 'UTC' })
