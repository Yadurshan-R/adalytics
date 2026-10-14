// Tests for the candle sizes on the candlestick chart.
package market

import (
	"context"
	"testing"
	"time"

	"cardano-analytics/internal/koios"
)

// TestCandleSizes checks that each candle size covers all the history loaded
// so far: 24 hours at first, then 7 days once the background load is done.
func TestCandleSizes(t *testing.T) {
	f := &fake{t: t, tip: 14030000}
	day := 24 * 60 * 60 / secondsPerBlock // blocks in a day
	f.txs = []fakeTx{
		{hash: "o1", height: 14030000 - 6*day, time: testNow.Add(-6 * 24 * time.Hour).Unix(), outs: []koios.TxOutput{poolOutput(t, "o1", "1900000000000")}},
		{hash: "a1", height: 14029600, time: testNow.Add(-2 * time.Hour).Unix(), outs: []koios.TxOutput{poolOutput(t, "a1", "2000000000000")}},
	}
	tr := newTracker(t, f)
	tr.SetHistoryDays(7)

	if _, _, _, err := tr.Candles("snek", "1h"); err != ErrRangeLoading {
		t.Errorf("before the first load: %v", err)
	}

	ctx := context.Background()
	tr.refresh(ctx) // the first 24 hours
	// The first load reads a little more than 24 hours, never the full 7 days.
	if _, size, got, err := tr.Candles("snek", "1H"); err != nil || size.Key != "1h" || len(got) < 25 || len(got) > 30 {
		t.Errorf("1h after 24 hours: %v, %s, %d candles, want 25 to 30", err, size.Key, len(got))
	}

	tr.backfill(ctx) // 7 days
	want := map[string]int{
		"5m": 2*24*12 + 1, "15m": 7*24*4 + 1, "30m": 7*24*2 + 1, "1h": 7*24 + 1,
		"2h": 7*12 + 1, "4h": 7*6 + 1, "12h": 7*2 + 1, "1D": 7 + 1,
	}
	for key, n := range want {
		_, _, got, err := tr.Candles("snek", key)
		if err != nil || len(got) != n {
			t.Errorf("%s: %v, %d candles, want %d", key, err, len(got), n)
			continue
		}
		last := got[len(got)-1]
		if last.Close != 2000000.0/744260765 {
			t.Errorf("%s: last close %v", key, last.Close)
		}
	}

	if _, _, _, err := tr.Candles("snek", "3m"); err == nil {
		t.Error("3m should be an unknown size")
	}
	if _, _, _, err := tr.Candles("nope", "1h"); err == nil {
		t.Error("an unknown token should be an error")
	}
}
