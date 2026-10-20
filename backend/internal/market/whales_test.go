// Tests for the Whale Trades page.
package market

import (
	"context"
	"testing"
	"time"

	"cardano-analytics/internal/koios"
)

// TestWhaleTrades checks the Whale Trades feed: trades of at least 1% of the
// pool's ADA or ₳10,000, newest first, with their share, and each filter.
func TestWhaleTrades(t *testing.T) {
	f := &fake{t: t, tip: 14030000}
	f.txs = []fakeTx{{hash: "s0", height: 14029000, time: testNow.Add(-time.Hour).Unix(), outs: []koios.TxOutput{swapOutput(t, "s0", "1000", "2000000000000", "750000000")}}}
	tr := newTracker(t, f)
	tr.refresh(context.Background())
	tr.mu.Lock()
	s := tr.states[0] // the pool holds ₳2,000,000, so 1% is ₳20,000
	s.trades = append(s.trades,
		Trade{Time: testNow.Add(-50 * time.Hour), Side: "buy", AmountADA: 60000, Wallet: "stake1e"},
		Trade{Time: testNow.Add(-30 * time.Minute), Side: "buy", AmountADA: 30000, Wallet: "stake1a"},
		Trade{Time: testNow.Add(-20 * time.Minute), Side: "sell", AmountADA: 5000, Wallet: "stake1b"},
		Trade{Time: testNow.Add(-15 * time.Minute), Side: "buy", AmountADA: 12000, Wallet: "stake1d"},
		Trade{Time: testNow.Add(-10 * time.Minute), Side: "sell", AmountADA: 20000, Wallet: "stake1c"})
	tr.mu.Unlock()

	wallets := func(f WhaleFilter) (out []string) {
		got, _ := tr.WhaleTrades(f)
		for _, w := range got {
			out = append(out, w.Wallet)
		}
		return out
	}
	same := func(name string, got []string, want ...string) {
		if len(got) != len(want) {
			t.Errorf("%s: %v, want %v", name, got, want)
			return
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("%s: %v, want %v", name, got, want)
				return
			}
		}
	}
	same("24h", wallets(WhaleFilter{}), "stake1c", "stake1d", "stake1a")
	same("7 days", wallets(WhaleFilter{Hours: 168}), "stake1c", "stake1d", "stake1a", "stake1e")
	same("₳25K and up", wallets(WhaleFilter{Hours: 168, MinADA: 25000}), "stake1a", "stake1e")
	same("fish", wallets(WhaleFilter{MaxADA: 15000}), "stake1d")
	same("₳20K to ₳50K", wallets(WhaleFilter{Hours: 168, MinADA: 20000, MaxADA: 50000}), "stake1c", "stake1a")
	same("wallet", wallets(WhaleFilter{Wallet: "STAKE1D"}), "stake1d")
	same("other token", wallets(WhaleFilter{Token: "nope"}))
	same("offset", wallets(WhaleFilter{Offset: 2}), "stake1a")

	got, more := tr.WhaleTrades(WhaleFilter{})
	if more || got[0].ID != "snek" || got[2].PoolShare != 1.5 || got[1].PoolShare != 0.6 {
		t.Errorf("first page: more=%v %+v", more, got)
	}
	if toks := tr.WhaleTokens(); len(toks) == 0 || toks[0].ID != "snek" {
		t.Errorf("tokens: %+v", toks)
	}
}
