// Tests for the first load: the newest hour first, then the rest of the day.
package market

import (
	"context"
	"math"
	"testing"
	"time"

	"cardano-analytics/internal/koios"
)

// TestListOpensAfterTheNewestHour checks the list opens with the newest hour
// read and the pool of every other token found, before the rest of the day:
// the price is right at once, the 24H range waits, and once the rest is read
// the 24h volume is complete and the pool found at startup is not counted
// twice.
func TestListOpensAfterTheNewestHour(t *testing.T) {
	f := &fake{t: t, tip: 14030000}
	f.txs = []fakeTx{
		{hash: "a0", height: 14029000, time: testNow.Add(-5 * time.Hour).Unix(), outs: []koios.TxOutput{poolOutput(t, "a0", "1977518866875")}},
		{hash: "a1", height: 14029600, time: testNow.Add(-2 * time.Hour).Unix(), outs: []koios.TxOutput{poolOutput(t, "a1", "2000000000000")}},
	}
	// SNEK did not trade in the last hour, so its pool is found as it is now.
	f.pools = []koios.UTxO{poolOutput(t, "a1", "2000000000000").UTxO(koios.Tx{BlockHeight: 14029600, TxTimestamp: f.txs[1].time})}
	tr := newTracker(t, f)
	var rest func()
	tr.async = func(fn func()) { rest = fn } // keep the rest of the day for later

	tr.refresh(context.Background())
	if !tr.loaded || rest == nil {
		t.Fatalf("list not open after the newest hour (loaded %v)", tr.loaded)
	}
	q := tr.Quotes()[0]
	if want := 2000000.0 / 744260765; math.Abs(q.PriceADA-want) > 1e-12 {
		t.Errorf("price %v, want %v", q.PriceADA, want)
	}
	h := tr.History()
	if h.LoadedHours != 1 || h.Ranges[0].Ready {
		t.Errorf("after the newest hour: %v hours, 24H ready %v", h.LoadedHours, h.Ranges[0].Ready)
	}

	rest() // the rest of the 24 hours
	q = tr.Quotes()[0]
	if math.Abs(q.Volume24hADA-22481.133125) > 1e-6 {
		t.Errorf("24h volume %v, want the one swap of 22481.133125", q.Volume24hADA)
	}
	if !tr.dayDone || !tr.History().Ranges[0].Ready {
		t.Error("the day should be done and the 24H range ready")
	}
	if n := len(tr.states[0].history); n != 2 {
		t.Errorf("%d pool states, want 2 (the one found at startup replaced by its transaction)", n)
	}
}
