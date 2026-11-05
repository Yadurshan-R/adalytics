// priceStep is the smallest price change the chart works with, for prices
// around `price`: one unit of the fifth significant digit (a price of
// 0.004650 gives 0.0000001; an FDV of 320,000,000 gives 10,000).
//
// The drawings use it twice: to count "ticks" (the measure tool and the long
// and short positions show a move as so many steps) and to choose how many
// decimals their labels show. One tiny step for every token (1e-12, small
// enough for HOSKY) made those counts absurd: 329,707,313 ticks for a move of
// 0.00033 ADA. Five significant digits match what the price axis shows.
export function priceStep(price) {
  if (!(price > 0) || !Number.isFinite(price)) return FALLBACK_STEP
  return 10 ** (Math.floor(Math.log10(price)) - 4)
}

// FALLBACK_STEP is the step while there is no price yet (an empty chart).
export const FALLBACK_STEP = 1e-12
