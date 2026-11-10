// Tests for volume: every order of a Minswap batch counts on its own, and the
// token list, the activity box and the candles all show the same number.
package market

import (
	"encoding/json"
	"io"
	"log"
	"math"
	"math/big"
	"testing"
	"time"

	"cardano-analytics/internal/koios"
	"cardano-analytics/internal/minswap"
)

// TestBatchVolumeCountsEveryOrder: in one batch Alice buys for ₳1,000 and Bob
// sells for ₳800. The pool's ADA moves by only ₳200, but ₳1,800 was traded,
// which is what Minswap shows for the pool. Before the fix Adalytics counted
// ₳200 here.
func TestBatchVolumeCountsEveryOrder(t *testing.T) {
	tr := NewTracker(koios.New("http://unused", "test-token"), []Token{snekToken(t)}, time.Minute, log.New(io.Discard, "", 0))
	tr.now = func() time.Time { return testNow }
	s := tr.states[0]

	pool := func(hash, reserveADA, reserveSNEK string, at time.Time) minswap.Pool {
		t.Helper()
		out := swapAt(t, hash, reserveADA)
		if reserveSNEK != "" {
			out = swapOutput(t, hash, out.InlineDatum.Value.Fields[3].Int.String(), reserveADA, reserveSNEK)
		}
		p, err := minswap.Parse(out.UTxO(koios.Tx{BlockHeight: 1, TxTimestamp: at.Unix()}))
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	ada := func(n int64) *big.Int { return new(big.Int).Mul(big.NewInt(n), big.NewInt(1_000_000)) }
	before := pool("b0", "2000000000000", "", testNow.Add(-2*time.Hour))
	// Alice gets 372,000 SNEK for ₳1,000; Bob gives 297,600 SNEK for ₳800:
	// the pool ends ₳200 up and 74,400 SNEK down.
	snek := new(big.Int).Sub(snekAt("2000000000000"), big.NewInt(74_400))
	after := pool("b1", "2000200000000", snek.String(), testNow.Add(-time.Hour))
	s.history = []poolState{
		{pool: before, height: 1},
		{pool: after, height: 2, fills: []fill{
			{wallet: "stake1alice", ada: ada(1000), token: big.NewInt(372_000)},
			{wallet: "stake1bob", ada: ada(-800), token: big.NewInt(-297_600)},
		}},
	}
	tr.mu.Lock()
	tr.rebuild(s)
	periods := periodsOf(s, testNow)
	tr.mu.Unlock()

	day := periods[len(periods)-1]
	if day.VolumeADA != 1800 {
		t.Errorf("24h volume %v, want ₳1,800 (not the pool's net ₳200)", day.VolumeADA)
	}
	if day.BuyADA != 1000 || day.SellADA != 800 || day.Buys != 1 || day.Sells != 1 {
		t.Errorf("buys %d for %v, sells %d for %v", day.Buys, day.BuyADA, day.Sells, day.SellADA)
	}
	if day.VolumeADA != day.BuyADA+day.SellADA {
		t.Errorf("volume %v is not buys + sells %v: the list and the activity box would differ", day.VolumeADA, day.BuyADA+day.SellADA)
	}
	if v := s.points[len(s.points)-1].VolumeADA; math.Abs(v-1800) > 1e-9 {
		t.Errorf("the candle's volume %v, want ₳1,800", v)
	}
}

// TestUnsplitBatchVolume: a batch that cannot be split into orders (no fills)
// still counts the pool's own change, as one trade.
func TestUnsplitBatchVolume(t *testing.T) {
	trades := tradesFrom(mustPool(t, "u1", "2000500000000"), nil, big.NewInt(500_000_000), big.NewInt(-185_000))
	if len(trades) != 1 || tradeVolume(trades) != 500 {
		t.Errorf("%d trades, volume %v, want one trade of ₳500", len(trades), tradeVolume(trades))
	}
}

// mustPool is the SNEK/ADA pool after a swap to reserveADA (lovelace).
func mustPool(t *testing.T, hash, reserveADA string) minswap.Pool {
	t.Helper()
	p, err := minswap.Parse(swapAt(t, hash, reserveADA).UTxO(koios.Tx{BlockHeight: 1, TxTimestamp: testNow.Unix()}))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// snekAt is the SNEK reserve that goes with reserveADA (lovelace) after a
// swap: the reserves' product stays as at ₳2,000,000 and 744,260,765 SNEK.
func snekAt(reserveADA string) *big.Int {
	a, _ := new(big.Int).SetString(reserveADA, 10)
	k := new(big.Int).Mul(big.NewInt(2_000_000_000_000), big.NewInt(744_260_765))
	return k.Quo(k, a)
}

// swapAt is the SNEK/ADA pool after a real swap: both reserves move (ADA in,
// SNEK out, or the other way), so the change is a trade. At ₳2,000,000 it is
// the same pool as poolOutput.
func swapAt(t *testing.T, txHash, reserveADA string) koios.TxOutput {
	out := poolOutput(t, txHash, reserveADA)
	b := json.Number(snekAt(reserveADA).String())
	out.InlineDatum.Value.Fields[5] = koios.PlutusData{Int: &b}
	return out
}
