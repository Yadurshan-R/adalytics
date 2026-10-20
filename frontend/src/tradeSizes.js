// The three trade sizes on the Large Trades page, like Birdeye's shrimp,
// dolphin and whale icons. Every trade on the page is already big for its
// token (at least 1% of the pool, or ₳10,000 and more); the icon tells how
// big it is in ADA. The Size list shows each with its range (picking one
// shows only that size); hovering an icon in the table shows its hint.
export const SIZES = [
  { key: 'fish', icon: '🐟', label: 'Fish', range: 'under ₳10K', min: 0, max: 10_000, hint: 'Fish: under ₳10K, but at least 1% of the pool, so it still moves the price about 2%' },
  { key: 'dolphin', icon: '🐬', label: 'Dolphin', range: '₳10K–₳50K', min: 10_000, max: 50_000, hint: 'Dolphin: ₳10K to ₳50K' },
  { key: 'whale', icon: '🐋', label: 'Whale', range: '₳50K+', min: 50_000, max: 0, hint: 'Whale: ₳50K or more' },
]

// sizeOf is the size of one trade: the biggest whose minimum it reaches.
export const sizeOf = (t) => SIZES.findLast((s) => t.amount_ada >= s.min)
