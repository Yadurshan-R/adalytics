package market

import (
	"math/big"
	"testing"
	"time"

	"cardano-analytics/internal/minswap"
)

// TestTradeRange checks the lowest and highest order price of a batch, and
// that a trade without a price is left out.
func TestTradeRange(t *testing.T) {
	low, high := tradeRange([]Trade{{PriceADA: 0.2418}, {PriceADA: 0}, {PriceADA: 0.2405}, {PriceADA: 0.2432}})
	if low != 0.2405 || high != 0.2432 {
		t.Errorf("range %v to %v, want 0.2405 to 0.2432", low, high)
	}
	if low, high := tradeRange(nil); low != 0 || high != 0 {
		t.Errorf("no trades: %v to %v, want 0, 0", low, high)
	}
}

// TestPointOfWicksOnlyFromSplitBatches checks that order prices make the
// wicks only when the batch was read order by order: an unsplit batch is one
// trade of the pool's net change, whose price can be mostly fees.
func TestPointOfWicksOnlyFromSplitBatches(t *testing.T) {
	p := &minswap.Pool{ReserveA: big.NewInt(2_000_000), ReserveB: big.NewInt(1_000_000), DecimalsB: 6, LastSwap: testNow}
	trades := []Trade{{PriceADA: 60, AmountADA: 60}} // ₳60 of fees for 1 token
	if pt := pointOf(p, trades, false); pt.High != 0 || pt.Low != 0 || pt.VolumeADA != 60 || pt.Price != 2 {
		t.Errorf("unsplit: %+v, want price 2, volume 60 and no wicks", pt)
	}
	if pt := pointOf(p, []Trade{{PriceADA: 2.01}, {PriceADA: 1.99}}, true); pt.High != 2.01 || pt.Low != 1.99 {
		t.Errorf("split: %+v, want wicks 1.99 to 2.01", pt)
	}
}

// TestCandleWicksFromOrderPrices checks that a candle's high and low take the
// order prices while open and close stay pool prices, and that a point
// without order prices (Low 0) does not pull the low to zero.
func TestCandleWicksFromOrderPrices(t *testing.T) {
	start := testNow.Truncate(5 * time.Minute).Unix()
	points := []Point{
		{Time: start - 60, Price: 0.24104},
		{Time: start + 28, Price: 0.24116, Low: 0.24182, High: 0.24182}, // a buy paid above the pool price
		{Time: start + 100, Price: 0.24110},                             // a deposit: no orders
	}
	c := candlesAt(points, []int64{start}, time.Unix(start+299, 0))[0]
	if c.Open != 0.24104 || c.Close != 0.24110 {
		t.Errorf("open %v close %v, want the pool prices 0.24104 and 0.24110", c.Open, c.Close)
	}
	if c.High != 0.24182 || c.Low != 0.24104 {
		t.Errorf("high %v low %v, want 0.24182 (the buy) and 0.24104", c.High, c.Low)
	}
}
