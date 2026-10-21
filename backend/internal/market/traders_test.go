// Tests for the Top Traders tab and one wallet's trades.
package market

import (
	"context"
	"fmt"
	"testing"
	"time"

	"cardano-analytics/internal/koios"
)

// tradersTracker is a tracker with one token and the given trades.
func tradersTracker(t *testing.T, trades []Trade) (*Tracker, string) {
	f := &fake{t: t, tip: 14030000}
	f.txs = []fakeTx{{hash: "s0", height: 14029000, time: testNow.Add(-time.Hour).Unix(), outs: []koios.TxOutput{swapOutput(t, "s0", "1000", "2000000000000", "750000000")}}}
	tr := newTracker(t, f)
	tr.refresh(context.Background())
	tr.mu.Lock()
	tr.states[0].trades = trades
	tr.mu.Unlock()
	return tr, tr.states[0].slug
}

// TestTopTraders checks that trades are added up per wallet, ranked by volume
// or net buying, cut to the time range, and that unknown wallets, unknown
// tokens, two orders in one batch and the 50-wallet limit are handled.
func TestTopTraders(t *testing.T) {
	tr, id := tradersTracker(t, []Trade{
		{Time: testNow.Add(-100 * time.Hour), Side: "buy", AmountADA: 90000, AmountToken: 9, Wallet: "stake1old"},
		{Time: testNow.Add(-3 * time.Hour), Side: "buy", AmountADA: 1000, AmountToken: 10, Wallet: "stake1a"},
		{Time: testNow.Add(-2 * time.Hour), Side: "sell", AmountADA: 4000, AmountToken: 30, Wallet: "stake1a", TxHash: "t1", Fill: 0},
		{Time: testNow.Add(-2 * time.Hour), Side: "sell", AmountADA: 500, AmountToken: 4, Wallet: "stake1a", TxHash: "t1", Fill: 1},
		{Time: testNow.Add(-90 * time.Minute), Side: "buy", AmountADA: 3000, AmountToken: 25, Wallet: "stake1b"},
		{Time: testNow.Add(-1 * time.Hour), Side: "buy", AmountADA: 99999, Wallet: ""},
		{Time: testNow.Add(-30 * time.Minute), Side: "buy", AmountADA: 2500, AmountToken: 20, Wallet: "stake1b"},
	})

	_, got, err := tr.TopTraders(id, 24, "volume")
	if err != nil || len(got) != 2 {
		t.Fatalf("24h by volume: %v, %v; want 2 wallets", got, err)
	}
	b, a := got[0], got[1] // stake1b ₳5,500, stake1a ₳5,500 too: ties go by wallet name
	if a.Wallet != "stake1a" || b.Wallet != "stake1b" {
		a, b = b, a
	}
	if b.Buys != 2 || b.Sells != 0 || b.BuyADA != 5500 || b.BuyToken != 45 || !b.Last.Equal(testNow.Add(-30*time.Minute)) {
		t.Errorf("stake1b = %+v", b)
	}
	if a.Buys != 1 || a.Sells != 2 || a.BuyADA != 1000 || a.SellADA != 4500 || a.SellToken != 34 {
		t.Errorf("stake1a (two orders in one batch) = %+v", a)
	}
	if _, got, _ := tr.TopTraders(id, 24, "net"); got[0].Wallet != "stake1b" || got[1].Wallet != "stake1a" {
		t.Errorf("by net buying: %s then %s, want stake1b (+₳5,500) then stake1a (−₳3,500)", got[0].Wallet, got[1].Wallet)
	}
	if _, got, _ := tr.TopTraders(id, 168, "volume"); len(got) != 3 || got[0].Wallet != "stake1old" {
		t.Errorf("7 days: %v, want stake1old first of 3", got)
	}
	if _, got, _ := tr.TopTraders(id, 5, "volume"); len(got) != 2 {
		t.Errorf("hours 5 should fall back to 24: %d wallets", len(got))
	}
	if _, _, err := tr.TopTraders("nope", 24, "volume"); err == nil {
		t.Error("an unknown token should give an error")
	}

	tr.mu.Lock()
	for i := range 60 {
		tr.states[0].trades = append(tr.states[0].trades, Trade{Time: testNow.Add(-time.Minute), Side: "buy", AmountADA: float64(i + 1), Wallet: fmt.Sprintf("stake1w%02d", i)})
	}
	tr.mu.Unlock()
	if _, got, _ := tr.TopTraders(id, 24, "volume"); len(got) != MaxTopTraders {
		t.Errorf("got %d wallets, want at most %d", len(got), MaxTopTraders)
	}
}

// TestTraderTrades checks a busy token: one wallet's trades are older than
// the 50 latest trades, yet its list and totals match its Top Traders row,
// also when it has more trades than the list holds.
func TestTraderTrades(t *testing.T) {
	var trades []Trade
	for i := range 2000 { // oldest first, about one trade a minute over 33 hours
		w := fmt.Sprintf("stake1x%d", i%40)
		if i%7 == 0 && i < 1500 {
			w = "stake1top"
		}
		side := "buy"
		if i%3 == 0 {
			side = "sell"
		}
		trades = append(trades, Trade{Time: testNow.Add(-time.Duration(2000-i) * time.Minute), Side: side, AmountADA: float64(100 + i%500), AmountToken: float64(i % 90), Wallet: w})
	}
	tr, id := tradersTracker(t, trades)

	_, recent, _ := tr.Trades(id, 50)
	for _, x := range recent {
		if x.Wallet == "stake1top" {
			t.Fatal("test setup: stake1top should not be in the 50 latest trades")
		}
	}
	for _, hours := range []int{24, 168} {
		_, top, _ := tr.TopTraders(id, hours, "volume")
		var row TopTrader
		for _, w := range top {
			if w.Wallet == "stake1top" {
				row = w
			}
		}
		_, sum, list, err := tr.TraderTrades(id, "stake1top", hours)
		if err != nil || row.Wallet == "" || sum != row {
			t.Fatalf("%dh: totals %+v, Top Traders row %+v, err %v", hours, sum, row, err)
		}
		var listed TopTrader
		for _, x := range list {
			if x.Wallet != "stake1top" {
				t.Fatalf("%dh: another wallet's trade in the list: %s", hours, x.Wallet)
			}
			listed.add(x)
		}
		if n := row.Buys + row.Sells; n <= MaxWalletTrades && (listed.Buys != row.Buys || listed.Sells != row.Sells || listed.BuyADA != row.BuyADA || listed.SellADA != row.SellADA) {
			t.Errorf("%dh: the list adds up to %+v, the row says %+v", hours, listed, row)
		}
		if len(list) != min(row.Buys+row.Sells, MaxWalletTrades) || (len(list) > 1 && list[0].Time.Before(list[1].Time)) {
			t.Errorf("%dh: %d trades listed for %d, newest first", hours, len(list), row.Buys+row.Sells)
		}
	}
	if _, sum, list, _ := tr.TraderTrades(id, "stake1nobody", 24); sum.Buys+sum.Sells != 0 || len(list) != 0 {
		t.Error("a wallet with no trades should get none")
	}
	if _, _, _, err := tr.TraderTrades("nope", "stake1top", 24); err == nil {
		t.Error("an unknown token should give an error")
	}
}
