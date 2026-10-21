// The three trade sizes on the Large Trades page. Every trade on the page is
// already big for its token (at least 1% of the pool, or ₳10,000 and more);
// the size tells how big it is in ADA. Big and Whale trades carry a small
// text tag in the table, like Birdeye's SWAP tag; small ones stay plain.
// The Size list shows each with its range (picking one shows only that size).
export const SIZES = [
  { key: 'small', label: 'Small', range: 'under ₳10K', min: 0, max: 10_000, tag: '', hint: '' },
  { key: 'big', label: 'Big', range: '₳10K–₳50K', min: 10_000, max: 50_000, tag: 'BIG', hint: 'Big trade: ₳10K to ₳50K' },
  { key: 'whale', label: 'Whale', range: '₳50K+', min: 50_000, max: 0, tag: 'WHALE', hint: 'Whale trade: ₳50K or more' },
]

// sizeOf is the size of one trade: the biggest whose minimum it reaches.
export const sizeOf = (t) => SIZES.findLast((s) => t.amount_ada >= s.min)
