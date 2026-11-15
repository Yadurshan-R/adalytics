package market

import "cardano-analytics/internal/minswap"

// pointOf makes the chart point for one pool state: the pool's price right
// after the transaction, the ADA it traded, and the lowest and highest price
// its orders were filled at (tradeRange). split says the batch was split into
// orders (fills); only then are the order prices used, see tradeRange.
func pointOf(p *minswap.Pool, trades []Trade, split bool) Point {
	pt := Point{Time: p.LastSwap.Unix(), Price: p.PriceADA(), VolumeADA: tradeVolume(trades)}
	if split {
		pt.Low, pt.High = tradeRange(trades)
	}
	return pt
}

// tradeRange returns the lowest and highest price the trades were filled at
// (ADA paid or received per token, PriceADA), or 0, 0 for no trades. These
// make a candle's wicks: a buyer pays a little above the pool price (the
// 0.3% pool fee plus slippage) and a seller gets a little below, and orders
// in one batch can move the price up and back down, which the pool price
// after the batch does not show. Charts such as Birdeye and DexScreener build
// highs and lows from trade prices the same way.
//
// Only orders read one by one are used (pointOf's split). A batch that could
// not be split is one trade of the pool's net change, and a net change can
// be mostly fees: ₳60 in for 1 token out would read as ₳60 a token. Trades
// under MinTradeADA are already left out (tradeFrom).
func tradeRange(trades []Trade) (low, high float64) {
	for _, tr := range trades {
		if tr.PriceADA <= 0 {
			continue
		}
		if low == 0 || tr.PriceADA < low {
			low = tr.PriceADA
		}
		high = max(high, tr.PriceADA)
	}
	return low, high
}

// tradeVolume is the ADA a pool state traded: the sum of its trades, so every
// order of a Minswap batch counts on its own (gross volume, as Minswap and
// DexScreener count it). The pool's own change nets buys against sells: one
// wallet buying for ₳1,000 and another selling for ₳800 in one batch move the
// pool by ₳200, but trade ₳1,800. A batch that could not be split into orders
// is one trade of the pool's change (tradesFrom).
func tradeVolume(trades []Trade) float64 {
	v := 0.0
	for _, tr := range trades {
		v += tr.AmountADA
	}
	return v
}
