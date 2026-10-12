// Tests for the first load: finding pools and loading 24 hours of swaps.
package market

import (
	"context"
	"io"
	"log"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cardano-analytics/internal/koios"
)

// TestLoadsHistoryThenFollowsSwaps checks the first 24 hour load (price,
// volume, change, chart), then that a new swap is picked up a minute later.
func TestLoadsHistoryThenFollowsSwaps(t *testing.T) {
	f := &fake{t: t, tip: 14030000}
	f.txs = []fakeTx{
		{hash: "a1", height: 14029600, time: testNow.Add(-2 * time.Hour).Unix(), outs: []koios.TxOutput{poolOutput(t, "a1", "1977518866875")}},
		{hash: "x1", height: 14029700, time: testNow.Add(-90 * time.Minute).Unix(), outs: []koios.TxOutput{otherOutput(t, "x1")}},
		{hash: "a2", height: 14029900, time: testNow.Add(-30 * time.Minute).Unix(), outs: []koios.TxOutput{poolOutput(t, "a2", "2000000000000")}},
	}
	f.pools = []koios.UTxO{poolOutput(t, "a2", "2000000000000").UTxO(koios.Tx{BlockHeight: 14029900, TxTimestamp: f.txs[2].time})}
	tr := newTracker(t, f)
	ctx := context.Background()

	tr.refresh(ctx) // finds the pool, then loads 24 hours
	q := tr.Quotes()[0]
	if q.Status != "live" {
		t.Fatalf("status %s (%s)", q.Status, q.Error)
	}
	if want := 2000000.0 / 744260765; math.Abs(q.PriceADA-want) > 1e-12 {
		t.Errorf("price %v, want %v", q.PriceADA, want)
	}
	if math.Abs(q.Volume24hADA-22481.133125) > 1e-6 {
		t.Errorf("24h volume %v", q.Volume24hADA)
	}
	if want := (2000000.0/1977518.866875 - 1) * 100; math.Abs(q.Change24h-want) > 1e-9 {
		t.Errorf("24h change %v, want %v", q.Change24h, want)
	}
	if len(q.Sparkline) != 25 {
		t.Errorf("sparkline has %d points, want 25 hourly points", len(q.Sparkline))
	}

	tr.refresh(ctx) // quiet minute: tip + utxo_info only

	// A swap happens; next check finds the pool spent and reads only recent blocks.
	f.mu.Lock()
	f.tip = 14030006
	f.txs = append(f.txs, fakeTx{hash: "b1", height: 14030004, time: testNow.Add(-time.Minute).Unix(),
		outs: []koios.TxOutput{poolOutput(t, "b1", "1990000000000")}})
	f.spent = true
	f.mu.Unlock()
	tr.refresh(ctx)

	q = tr.Quotes()[0]
	if q.PoolUTxO != "b1#0" {
		t.Errorf("pool after swap = %s", q.PoolUTxO)
	}
	if math.Abs(q.Volume24hADA-32481.133125) > 1e-6 {
		t.Errorf("24h volume after swap %v", q.Volume24hADA)
	}
	_, _, points, _, err := tr.Chart("snek", "24h", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 289 {
		t.Errorf("24H chart has %d points, want 289", len(points))
	}
	if last := points[len(points)-1]; math.Abs(last.Price-1990000.0/744260765) > 1e-12 {
		t.Errorf("last chart price %v", last.Price)
	}

	if q.TotalSupply != 76715880000 || math.Abs(q.FDVADA-q.PriceADA*76715880000) > 1e-3 || !q.HasLogo {
		t.Errorf("supply %v, fdv %v, logo %v", q.TotalSupply, q.FDVADA, q.HasLogo)
	}
	if logo, ok := tr.Logo("snek"); !ok || string(logo[1:4]) != "PNG" {
		t.Errorf("logo not served: %q", logo)
	}

	tr.loadHolders(ctx)
	if h := tr.Quotes()[0].Holders; h != 39514 {
		t.Errorf("holders %d, want 39514", h)
	}

	if q.Rank != 1 || q.ID != "snek" {
		t.Errorf("rank %d, id %q", q.Rank, q.ID)
	}

	// SNEK traded in the last 24 hours, so no pool search (asset_utxos) is needed.
	want := "/tip /address_txs /tx_info /asset_info /tip /utxo_info /tip /utxo_info /address_txs /tx_info /asset_summary"
	if got := strings.Join(f.calls, " "); got != want {
		t.Errorf("requests:\n got  %s\n want %s", got, want)
	}
	// History from 25 hours back; the live read only from the previous check.
	if len(f.queries) != 2 || f.queries[0] != "14025500" || f.queries[1] != "14029997" {
		t.Errorf("address_txs from heights %v, want [14025500 14029997]", f.queries)
	}
}

// TestQuietTokenIsListed checks that a token with a pool but no trades today
// is still listed, with a flat price line.
func TestQuietTokenIsListed(t *testing.T) {
	f := &fake{t: t, tip: 14030000}
	f.txs = []fakeTx{{hash: "x1", height: 14029700, time: testNow.Add(-time.Hour).Unix(), outs: []koios.TxOutput{otherOutput(t, "x1")}}}
	// The SNEK pool last traded three days ago.
	f.pools = []koios.UTxO{poolOutput(t, "old", "2000000000000").UTxO(koios.Tx{BlockHeight: 14017000, TxTimestamp: testNow.Add(-72 * time.Hour).Unix()})}
	tr := newTracker(t, f)

	tr.refresh(context.Background())
	quotes := tr.Quotes()
	if len(quotes) != 1 {
		t.Fatalf("a token with a pool but no trades today should be listed: %+v", quotes)
	}
	q := quotes[0]
	if want := 2000000.0 / 744260765; math.Abs(q.PriceADA-want) > 1e-12 {
		t.Errorf("price %v, want %v", q.PriceADA, want)
	}
	if q.Change24h != 0 || q.Change1h != 0 || q.Volume24hADA != 0 || q.PoolUTxO != "old#0" {
		t.Errorf("change %v/%v, volume %v, pool %s", q.Change1h, q.Change24h, q.Volume24hADA, q.PoolUTxO)
	}
	if len(q.Sparkline) != 25 || q.Sparkline[0] != q.PriceADA || q.Sparkline[24] != q.PriceADA {
		t.Errorf("sparkline should be flat at the price: %v", q.Sparkline)
	}
	_, _, points, _, _ := tr.Chart("snek", "24H", false)
	if len(points) != 289 || points[0].Price != q.PriceADA {
		t.Errorf("chart should be flat: %d points", len(points))
	}
}

// TestTokenWithoutPoolIsLeftOut checks that a token with no Minswap pool is not listed.
func TestTokenWithoutPoolIsLeftOut(t *testing.T) {
	f := &fake{t: t, tip: 14030000}
	tr := newTracker(t, f)
	tr.refresh(context.Background())
	if q := tr.Quotes(); len(q) != 0 {
		t.Fatalf("a token without a Minswap pool should not be listed: %+v", q)
	}
	if got := strings.Join(f.calls, " "); got != "/tip /address_txs /asset_utxos" {
		t.Errorf("requests: %s", got)
	}
}

// TestFindsAllPoolsInSmallRequests checks that the pool search is split into
// requests small enough for the Koios free tier.
func TestFindsAllPoolsInSmallRequests(t *testing.T) {
	f := &fake{t: t, tip: 14030000}
	srv := httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(srv.Close)
	tr := NewTracker(koios.New(srv.URL, "test-token"), Tokens, time.Minute, log.New(io.Discard, "", 0))
	if _, err := tr.findPools(context.Background(), tr.states); err != nil {
		t.Fatal(err)
	}
	if want := (len(Tokens) + 29) / 30; len(f.bodies) != want {
		t.Errorf("%d asset_utxos requests for %d tokens, want %d", len(f.bodies), len(Tokens), want)
	}
	for i, n := range f.bodies {
		if n >= koios.MaxBodyBytes {
			t.Errorf("request %d body is %d bytes, Koios accepts less than %d", i, n, koios.MaxBodyBytes)
		}
	}
}
