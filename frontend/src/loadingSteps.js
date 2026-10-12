// Text for the loading page: what each step of the backend's first load is
// called and what it does, in plain words, plus small formatting helpers.

// The four steps, in the order the backend runs them. The keys match the
// "steps" the backend sends in /api/tokens → history.loading.
export const STEPS = {
  trades_found: {
    title: 'Finding trades',
    text: 'Listing every Minswap transaction from the last 24 hours on the Cardano blockchain.',
  },
  reading: {
    title: 'Reading trades',
    text: 'Opening each transaction to read pool sizes, which gives prices, volume, buys and sells.',
  },
  quiet: {
    title: 'Checking quiet tokens',
    text: 'Finding the current pool of tokens that did not trade today, so they are listed too.',
  },
  details: {
    title: 'Loading token details',
    text: 'Getting total supply, logos and website links from the Cardano token registry.',
  },
}

// n formats a whole number with thousands separators: 12345 → "12,345".
export const n = (v) => Number(v ?? 0).toLocaleString('en-US')

// stepCount is the number shown next to a step, e.g. "1,200 / 3,400".
// Steps not started yet, and the details step, show nothing.
export function stepCount(step) {
  if (step.state === 'todo' || step.key === 'details') return ''
  if (step.key === 'trades_found') return n(step.done)
  return step.total ? `${n(step.done)} / ${n(step.total)}` : n(step.done)
}

// formatDuration turns seconds into short text: 45 → "45s", 135 → "2m 15s".
export function formatDuration(seconds) {
  const s = Math.max(0, Math.round(seconds))
  if (s < 60) return `${s}s`
  const m = Math.floor(s / 60)
  return s % 60 ? `${m}m ${s % 60}s` : `${m}m`
}

// clockTime shows an ISO time as the local clock time, e.g. "10:42:07".
export function clockTime(iso) {
  return new Date(iso).toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}
