// Tests for the activity box periods (5m, 1h, 4h, 24h).
package market

import (
	"context"
	"testing"
	"time"

	"cardano-analytics/internal/koios"
)

// TestPeriods checks buys, sells, ADA and the price change for each period:
// a buy 3h20m ago and a sell 1h40m ago fall in 4h and 24h, not in 5m or 1h.
func TestPeriods(t *testing.T) {
	f := &fake{t: t, tip: 14030000}
	at := func(m int) int64 { return testNow.Add(-time.Duration(m) * time.Minute).Unix() }
	f.txs = []fakeTx{
		{hash: "s0", height: 14029000, time: at(300), outs: []koios.TxOutput{swapOutput(t, "s0", "1000", "2000000000000", "750000000")}},
		{hash: "s1", height: 14029100, time: at(200), outs: []koios.TxOutput{swapOutput(t, "s1", "1000", "2001000000000", "749625000")}},
		{hash: "s2", height: 14029200, time: at(100), outs: []koios.TxOutput{swapOutput(t, "s2", "1000", "2000500000000", "749812500")}},
	}
	tr := newTracker(t, f)
	tr.refresh(context.Background())

	_, got, err := tr.Periods("snek")
	if err != nil || len(got) != 4 {
		t.Fatalf("%v, %d periods", err, len(got))
	}
	for i, key := range []string{"5m", "1h", "4h", "24h"} {
		p := got[i]
		recent := key == "5m" || key == "1h"
		switch {
		case p.Key != key:
			t.Errorf("period %d is %s, want %s", i, p.Key, key)
		case recent && (p.Buys != 0 || p.Sells != 0 || p.ChangePct != 0):
			t.Errorf("%s should be quiet: %+v", key, p)
		case !recent && (p.Buys != 1 || p.Sells != 1 || p.BuyADA != 1000 || p.SellADA != 500 || p.ChangePct == 0):
			t.Errorf("%s: %+v", key, p)
		}
	}
	if _, _, err := tr.Periods("nope"); err == nil {
		t.Error("an unknown token should be an error")
	}
}
