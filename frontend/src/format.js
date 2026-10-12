// Number formatting for ADA prices, percentages and amounts.

const SUBSCRIPT = '₀₁₂₃₄₅₆₇₈₉'

// ₳0.1699, ₳0.002681, and for tiny prices CoinGecko-style ₳0.0₇3391
// (the small 7 means seven zeros after the decimal point).
export function formatPrice(v) {
  return priceWith(v, '₳')
}

// $0.0503, $1.01, $0.0₇8970: a price in US dollars, same style as formatPrice.
export function formatUSD(v) {
  return priceWith(v, '$')
}

// priceWith formats a price with ₳ or $, using 4 significant digits and the
// small-number style for tiny prices. formatPrice and formatUSD both use it.
function priceWith(v, symbol) {
  if (v == null || !isFinite(v) || v <= 0) return '—'
  if (v >= 1) return symbol + v.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 4 })
  const zeros = Math.floor(-Math.log10(v)) // zeros right after "0."
  const digits = Math.round(v * 10 ** (zeros + 4)).toString().slice(0, 4)
  if (zeros >= 5) {
    const sub = String(zeros).split('').map((d) => SUBSCRIPT[d]).join('')
    return `${symbol}0.0${sub}${digits}`
  }
  return symbol + v.toFixed(zeros + 4)
}

// Full price for tooltips, e.g. 0.00000003391
export function fullPrice(v) {
  if (v == null || !isFinite(v) || v <= 0) return ''
  return '₳' + v.toLocaleString('en-US', { maximumSignificantDigits: 6 })
}

// ₳1,234,567 (whole ADA), like CoinGecko's volume and market cap columns.
export function formatADA(v) {
  if (v == null || !isFinite(v)) return '—'
  return '₳' + Math.round(v).toLocaleString('en-US')
}

// formatPct shows a change as a percentage with one decimal, without its sign:
// the arrow and colour show the direction.
export function formatPct(v) {
  if (v == null || !isFinite(v)) return '—'
  return Math.abs(v).toFixed(1) + '%'
}

// formatTime shows the clock time, e.g. 14:05:09.
export function formatTime(iso) {
  if (!iso) return ''
  return new Date(iso).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false })
}

// "just now", "12 min ago", "3 hours ago", "4 days ago"
export function formatAgo(iso, now = Date.now()) {
  if (!iso) return ''
  const s = Math.max(0, (now - new Date(iso).getTime()) / 1000)
  if (s < 60) return 'just now'
  const units = [[86400, 'day'], [3600, 'hour'], [60, 'min']]
  for (const [size, name] of units) {
    if (s >= size) {
      const n = Math.floor(s / size)
      return `${n} ${name}${n === 1 || name === 'min' ? '' : 's'} ago`
    }
  }
}

// ₳3.71T, ₳3.35M, ₳309.9K, ₳812, like CoinGecko's stats column.
export function formatCompact(v) {
  if (v == null || !isFinite(v)) return '—'
  const abs = Math.abs(v)
  if (abs >= 1e12) return '₳' + (v / 1e12).toFixed(2) + 'T'
  if (abs >= 1e9) return '₳' + (v / 1e9).toFixed(2) + 'B'
  if (abs >= 1e6) return '₳' + (v / 1e6).toFixed(2) + 'M'
  if (abs >= 1e3) return '₳' + (v / 1e3).toFixed(1) + 'K'
  return '₳' + Math.round(v).toLocaleString('en-US')
}

// "Oct 5, 2026, 13:20:00 GMT+5:30" for chart tooltips.
export function formatDateTime(unixSeconds) {
  return new Date(unixSeconds * 1000).toLocaleString('en-US', {
    month: 'short', day: 'numeric', year: 'numeric',
    hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false, timeZoneName: 'shortOffset',
  })
}

// 76.72B, 21.0M, 940K: token amounts without the ₳ sign.
export function formatAmount(v) {
  if (v == null || !isFinite(v)) return '—'
  const abs = Math.abs(v)
  if (abs >= 1e12) return (v / 1e12).toFixed(2) + 'T'
  if (abs >= 1e9) return (v / 1e9).toFixed(2) + 'B'
  if (abs >= 1e6) return (v / 1e6).toFixed(2) + 'M'
  if (abs >= 1e3) return (v / 1e3).toFixed(1) + 'K'
  return Math.round(v).toLocaleString('en-US')
}

// short shortens a long id for the info box: the first 8 and last 6 characters.
export const short = (s) => (s && s.length > 16 ? `${s.slice(0, 8)}…${s.slice(-6)}` : s)

// host shows only the website's name, e.g. https://www.snek.com/x → snek.com.
export function host(url) {
  try {
    return new URL(url).hostname.replace(/^www\./, '')
  } catch {
    return url
  }
}

// createdDate shows a date like Apr 24, 2023.
export const createdDate = (iso) => new Date(iso).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
