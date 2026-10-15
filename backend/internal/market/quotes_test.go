// Tests for what the token list shows: trades, USD prices, ranking and links.
package market

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"testing"
	"time"

	"cardano-analytics/internal/koios"
)

// TestTradesFromSwaps checks that buys and sells are worked out from pool
// changes, and that dust and liquidity deposits are not counted as trades.
func TestTradesFromSwaps(t *testing.T) {
	f := &fake{t: t, tip: 14030000}
	at := func(m int) int64 { return testNow.Add(-time.Duration(m) * time.Minute).Unix() }
	f.txs = []fakeTx{
		// Starting state, then a buy (ADA in, SNEK out), a sell, and a deposit.
		{hash: "s0", height: 14029000, time: at(300), outs: []koios.TxOutput{swapOutput(t, "s0", "1000", "2000000000000", "750000000")}},
		{hash: "s1", height: 14029100, time: at(200), outs: []koios.TxOutput{swapOutput(t, "s1", "1000", "2001000000000", "749625000")}},
		{hash: "s2", height: 14029200, time: at(100), outs: []koios.TxOutput{swapOutput(t, "s2", "1000", "2000500000000", "749812500")}},
		// Dust: 0.5 ADA out for 187 SNEK in, under MinTradeADA.
		{hash: "d1", height: 14029250, time: at(75), outs: []koios.TxOutput{swapOutput(t, "d1", "1000", "2000499500000", "749812687")}},
		{hash: "s3", height: 14029300, time: at(50), outs: []koios.TxOutput{swapOutput(t, "s3", "2000", "4001000000000", "1499625000")}},
	}
	tr := newTracker(t, f)
	tr.refresh(context.Background())

	_, trades, err := tr.Trades("snek", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(trades) != 2 {
		t.Fatalf("got %d trades, want 2 (the dust and the deposit are not trades): %+v", len(trades), trades)
	}
	sell, buy := trades[0], trades[1] // newest first
	if sell.Side != "sell" || sell.AmountADA != 500 || sell.AmountToken != 187500 || sell.TxHash != "s2" {
		t.Errorf("sell: %+v", sell)
	}
	if buy.Side != "buy" || buy.AmountADA != 1000 || buy.AmountToken != 375000 || buy.TxHash != "s1" {
		t.Errorf("buy: %+v", buy)
	}
	if math.Abs(buy.PriceADA-1000.0/375000) > 1e-15 || !buy.Time.Equal(time.Unix(at(200), 0)) {
		t.Errorf("buy price %v, time %v", buy.PriceADA, buy.Time)
	}
	if _, one, _ := tr.Trades("snek", 1); len(one) != 1 || one[0].TxHash != "s2" {
		t.Errorf("limit 1 should give the newest trade: %+v", one)
	}
	if _, _, err := tr.Trades("nope", 0); err == nil {
		t.Error("unknown token should be an error")
	}

	q := tr.Quotes()[0]
	if q.Trades24h != 2 || q.Buys24h != 1 || q.Sells24h != 1 || q.BuyVol24hADA != 1000 || q.SellVol24hADA != 500 {
		t.Errorf("24h trades %d, buys %d, sells %d, buy vol %v, sell vol %v", q.Trades24h, q.Buys24h, q.Sells24h, q.BuyVol24hADA, q.SellVol24hADA)
	}
	if q.Fingerprint != "asset108xu02ckwrfc8qs9d97mgyh4kn8gdu9w8f5sxk" || q.Description != "The snek token." || q.Website != "https://snek.com" ||
		q.Created == nil || !q.Created.Equal(time.Unix(1682340000, 0)) {
		t.Errorf("info: %q %q %q %v", q.Fingerprint, q.Description, q.Website, q.Created)
	}
}

// TestUSDPriceFromUSDMPool checks USD prices come from the USDM pool, and are
// left empty when that pool is missing.
func TestUSDPriceFromUSDMPool(t *testing.T) {
	mk := func(asset koios.Asset, ticker string, price float64) *tokenState {
		s := &tokenState{token: Token{Ticker: ticker, Asset: asset}, poolRef: "x#0"}
		s.quote = Quote{Ticker: ticker, PriceADA: price, ReserveADA: 50_000}
		return s
	}
	snek := mk(koios.Asset{PolicyID: "279c909f"}, "SNEK", 0.37)
	tr := &Tracker{states: []*tokenState{snek, mk(USDM, "USDM", 3.7)}, now: func() time.Time { return testNow }}
	for _, q := range tr.Quotes() {
		if q.Ticker == "SNEK" && math.Abs(q.PriceUSD-0.1) > 1e-12 {
			t.Errorf("SNEK in USD = %v, want 0.1", q.PriceUSD)
		}
	}
	// Without the USDM pool there is no USD price, rather than a wrong one.
	tr.states = tr.states[:1]
	if q := tr.Quotes()[0]; q.PriceUSD != 0 {
		t.Errorf("USD price without USDM = %v", q.PriceUSD)
	}
}

// TestSafeURL checks that only plain http(s) website links are kept.
func TestSafeURL(t *testing.T) {
	cases := map[string]string{
		"https://snek.com":    "https://snek.com",
		" http://a.io/x ":     "http://a.io/x",
		"javascript:alert(1)": "",
		"snek.com":            "",
		"":                    "",
	}
	for in, want := range cases {
		if got := safeURL(in); got != want {
			t.Errorf("safeURL(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestRankingAndIDs checks the ranking by volume then pool size, that small
// pools are left out, and that tokens sharing a ticker get unique ids.
func TestRankingAndIDs(t *testing.T) {
	mk := func(ticker, policy string, volume, reserve float64) *tokenState {
		s := &tokenState{token: Token{Ticker: ticker, Asset: koios.Asset{PolicyID: policy}}, poolRef: "x#0", supply: big.NewInt(1_000)}
		s.quote = Quote{Ticker: ticker, PriceADA: 1, ReserveADA: reserve}
		s.points = []Point{{Time: testNow.Add(-time.Hour).Unix(), Price: 1, VolumeADA: volume}}
		return s
	}
	states := []*tokenState{
		mk("AAA", "aaaaaaaa11", 1_000, 50_000), // volume 1,000
		mk("BBB", "bbbbbbbb22", 9_000, 50_000), // volume 9,000
		mk("AAA", "cccccccc33", 99_000, 5_000), // pool too small: left out
		mk("DDD", "dddddddd44", 0, 80_000),     // no trades, bigger pool
		mk("EEE", "eeeeeeee55", 0, 60_000),     // no trades, smaller pool
	}
	assignSlugs(states)
	if states[0].slug != "aaa-aaaaaa" || states[2].slug != "aaa-cccccc" || states[1].slug != "bbb" {
		t.Errorf("slugs: %s %s %s", states[0].slug, states[1].slug, states[2].slug)
	}
	tr := &Tracker{states: states, now: func() time.Time { return testNow }}
	got := ""
	for _, q := range tr.Quotes() {
		got += fmt.Sprintf("%d:%s ", q.Rank, q.ID)
	}
	if got != "1:bbb 2:aaa-aaaaaa 3:ddd 4:eee " {
		t.Errorf("ranking: %s", got)
	}
}

// TestTradersCount checks that 24h traders counts each wallet once and skips
// trades whose wallet is not known.
func TestTradersCount(t *testing.T) {
	f := &fake{t: t, tip: 14030000}
	f.txs = []fakeTx{{hash: "s0", height: 14029000, time: testNow.Add(-time.Hour).Unix(), outs: []koios.TxOutput{swapOutput(t, "s0", "1000", "2000000000000", "750000000")}}}
	tr := newTracker(t, f)
	tr.refresh(context.Background())
	tr.mu.Lock()
	for i, w := range []string{"stake1a", "stake1b", "stake1a", ""} {
		tr.states[0].trades = append(tr.states[0].trades, Trade{Time: testNow.Add(-time.Duration(i+1) * time.Minute), Side: "buy", AmountADA: 10, Wallet: w})
	}
	tr.states[0].trades = append([]Trade{{Time: testNow.Add(-25 * time.Hour), Side: "buy", Wallet: "stake1old"}}, tr.states[0].trades...)
	tr.mu.Unlock()
	if q := tr.Quotes()[0]; q.Traders24h != 2 {
		t.Errorf("24h traders %d, want 2", q.Traders24h)
	}
}
