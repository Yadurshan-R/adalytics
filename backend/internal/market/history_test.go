// Tests for the background history load and the loading progress.
package market

import (
	"context"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"cardano-analytics/internal/koios"
)

// TestBackfillLoadsOlderHistory checks that 7 days of history are loaded in the
// background, and that the 7D chart and change appear only once they are.
func TestBackfillLoadsOlderHistory(t *testing.T) {
	f := &fake{t: t, tip: 14030000}
	day := 24 * 60 * 60 / secondsPerBlock // blocks in a day
	f.txs = []fakeTx{
		{hash: "o1", height: 14030000 - 6*day, time: testNow.Add(-6 * 24 * time.Hour).Unix(), outs: []koios.TxOutput{poolOutput(t, "o1", "1900000000000")}},
		{hash: "o2", height: 14030000 - 3*day, time: testNow.Add(-3 * 24 * time.Hour).Unix(), outs: []koios.TxOutput{poolOutput(t, "o2", "1950000000000")}},
		{hash: "a1", height: 14029600, time: testNow.Add(-2 * time.Hour).Unix(), outs: []koios.TxOutput{poolOutput(t, "a1", "2000000000000")}},
	}
	tr := newTracker(t, f)
	tr.SetHistoryDays(7)
	ctx := context.Background()

	tr.refresh(ctx) // the first 24 hours
	h := tr.History()
	if len(h.Ranges) != 2 || h.Ranges[0].Key != "24H" || !h.Ranges[0].Ready || h.Ranges[1].Ready || h.Ranges[1].Progress != 14 {
		t.Fatalf("after 24h: %+v", h)
	}
	if _, _, _, _, err := tr.Chart("snek", "7D", false); err != ErrRangeLoading {
		t.Errorf("7D before it is loaded: %v", err)
	}
	if q := tr.Quotes()[0]; q.Change7d != nil || q.Sparkline7d != nil {
		t.Errorf("7d change or line before 7 days are loaded: %v %v", q.Change7d, q.Sparkline7d)
	}

	before := len(f.calls)
	tr.backfill(ctx)
	// One page of the 7 day list, then only the two older transactions are read.
	if got := strings.Join(f.calls[before:], " "); got != "/tip /address_txs /tx_info" {
		t.Errorf("backfill requests: %s", got)
	}
	h = tr.History()
	if !h.Done || !h.Ranges[1].Ready || h.Ranges[1].Progress != 100 {
		t.Errorf("after backfill: %+v", h)
	}
	q := tr.Quotes()[0]
	if want := (2000000.0/1900000 - 1) * 100; q.Change7d == nil || math.Abs(*q.Change7d-want) > 1e-9 {
		t.Errorf("7d change %v, want %v", q.Change7d, want)
	}
	if len(q.Sparkline) != 25 || len(q.Sparkline7d) != 85 {
		t.Errorf("sparklines have %d and %d points, want 25 (hourly) and 85 (every 2 hours)", len(q.Sparkline), len(q.Sparkline7d))
	}
	_, r, points, _, err := tr.Chart("snek", "7d", false)
	if err != nil || r.Key != "7D" || len(points) != 169 || points[0].Price != 1900000.0/744260765 {
		t.Errorf("7D chart: %v, %d points", err, len(points))
	}
	_, _, _, candles, err := tr.Chart("snek", "7D", true)
	if err != nil || len(candles) != 169 {
		t.Errorf("7D candles: %v, %d", err, len(candles))
	}
	if _, _, _, _, err := tr.Chart("snek", "30D", false); err != ErrRangeLoading {
		t.Errorf("30D with HISTORY_DAYS=7 should never be ready: %v", err)
	}
}

// The background history load runs while the minute updates and the API read
// the same tokens; run them together so the race detector can check it.
func TestBackfillWhileUpdating(t *testing.T) {
	f := &fake{t: t, tip: 14030000, spent: true}
	day := 24 * 60 * 60 / secondsPerBlock
	for i := 0; i < 40; i++ {
		h := 14030000 - (i+1)*day/8
		hash := fmt.Sprintf("t%02d", i)
		f.txs = append([]fakeTx{{hash: hash, height: h, time: testNow.Add(-time.Duration(i+1) * 3 * time.Hour).Unix(),
			outs: []koios.TxOutput{poolOutput(t, hash, fmt.Sprint(2000000000000+i*1000000))}}}, f.txs...)
	}
	tr := newTracker(t, f)
	tr.SetHistoryDays(7)
	ctx := context.Background()
	tr.refresh(ctx)

	done := make(chan struct{})
	go func() { tr.backfill(ctx); close(done) }()
	for i := 0; i < 5; i++ {
		tr.update(ctx)
		tr.Quotes()
		tr.History()
		tr.Chart("snek", "24H", true)
		tr.Trades("snek", 5)
	}
	<-done
	if h := tr.History(); !h.Done {
		t.Errorf("history not done: %+v", h)
	}
}

// TestLoadingProgress checks the loading screen's steps and percentage.
func TestLoadingProgress(t *testing.T) {
	f := &fake{t: t, tip: 14030000}
	tr := newTracker(t, f)

	h := tr.History()
	if h.Loading == nil || h.Loading.Percent != 0 || h.Loading.Steps[0].State != "active" || h.Loading.Steps[3].State != "todo" {
		t.Fatalf("before loading: %+v", h.Loading)
	}
	// Half way through reading 2,000 transactions: 10% + 70% / 2.
	tr.setBoot(func(b *bootProgress) { b.step, b.found, b.read = 1, 2000, 1000 })
	l := tr.History().Loading
	if l.Percent != 45 || l.Steps[0].State != "done" || l.Steps[1].State != "active" || l.Steps[1].Done != 1000 || l.Steps[1].Total != 2000 {
		t.Errorf("while reading: %+v", l)
	}
	tr.Note("Listed %d transactions", 2000)
	l = tr.History().Loading
	if len(l.Activity) != 1 || l.Activity[0].Text != "Listed 2000 transactions" {
		t.Errorf("activity: %+v", l.Activity)
	}
	if got := timeLeft(60*time.Second, 25); got != 180 {
		t.Errorf("time left after 60s at 25%% = %ds, want 180s", got)
	}
	if got := timeLeft(10*time.Second, 2); got != 0 {
		t.Errorf("time left should be unknown below 5%%, got %ds", got)
	}

	tr.refresh(context.Background())
	if l := tr.History().Loading; l != nil {
		t.Errorf("loading still shown after the first load: %+v", l)
	}
}
