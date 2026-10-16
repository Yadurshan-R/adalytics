// What the token list shows and sorts by, like Birdeye's table: each token's
// numbers for one period (5m, 1h, 4h or 24h), and the column sorts.

// PERIODS are the change columns in the table. 5m is left out: most tokens
// do not trade every 5 minutes, so it was mostly empty (the token page's
// activity box still has it).
export const PERIODS = ['1h', '4h', '24h']
// HOURS is each period's length, to know when its history is fully loaded.
export const HOURS = { '5m': 1 / 12, '1h': 1, '4h': 4, '24h': 24 }

// periodOf is a token's trading over one period: change, volume, buys and sells.
export function periodOf(t, key) {
  return t.periods?.find((p) => p.key === key) ?? { change_pct: 0, volume_ada: 0, buys: 0, sells: 0 }
}

// COLUMNS is every sortable column: its key, title and the value it sorts by.
// Volume and Txns follow the period picked above the table.
export const COLUMNS = [
  { key: 'price', title: 'Price', value: (t) => t.price_ada },
  { key: 'fdv', title: 'FDV', value: (t) => t.fdv_ada ?? 0 },
  { key: 'volume', title: 'Vol', value: (t, p) => periodOf(t, p).volume_ada, period: true },
  { key: 'txns', title: 'Txns', value: (t, p) => periodOf(t, p).buys + periodOf(t, p).sells, period: true },
  { key: 'liquidity', title: 'Liquidity', value: (t) => t.reserve_ada * 2 },
  ...PERIODS.map((k) => ({ key: `c${k}`, title: `${k} %`, value: (t) => periodOf(t, k).change_pct, change: k })),
]

// sortRows sorts a copy of the rows by a column ({ key, dir: 1 up or -1 down}).
// Equal values (often quiet tokens) keep the bigger pool first.
export function sortRows(rows, sort, period) {
  const col = COLUMNS.find((c) => c.key === sort.key)
  if (!col) return rows
  return [...rows].sort((a, b) => (col.value(a, period) - col.value(b, period)) * sort.dir || b.reserve_ada - a.reserve_ada)
}

// signedPct shows a change like +3.65% or −0.64%, with 0.00% for no change.
export function signedPct(v) {
  if (!isFinite(v)) return '—'
  const s = Math.abs(v) >= 1000 ? Math.round(Math.abs(v)).toLocaleString('en-US') : Math.abs(v).toFixed(2)
  return (v > 0.005 ? '+' : v < -0.005 ? '−' : '') + s + '%'
}
