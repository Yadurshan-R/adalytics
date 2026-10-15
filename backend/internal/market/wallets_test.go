// Tests for one wallet's trades across tokens.
package market

import (
	"testing"
	"time"
)

// TestWallet checks a wallet's trades are added up per token for the last 24
// hours, latest token first, and other wallets' trades are left out.
func TestWallet(t *testing.T) {
	at := func(h float64) time.Time { return testNow.Add(-time.Duration(h * float64(time.Hour))) }
	snek := &tokenState{slug: "snek", quote: Quote{Ticker: "SNEK", ReserveADA: 2_300_000}, trades: []Trade{
		{Time: at(30), Side: "buy", AmountADA: 900, Wallet: "stake1me"}, // older than 24 hours
		{Time: at(5), Side: "buy", AmountADA: 100, Wallet: "stake1me"},
		{Time: at(4), Side: "sell", AmountADA: 70, Wallet: "stake1other"},
		{Time: at(3), Side: "sell", AmountADA: 40, Wallet: "stake1me"},
	}}
	min := &tokenState{slug: "min", quote: Quote{Ticker: "MIN"}, trades: []Trade{
		{Time: at(1), Side: "buy", AmountADA: 25, Wallet: "stake1me"},
	}}
	quiet := &tokenState{slug: "iag", quote: Quote{Ticker: "IAG"}}
	tr := &Tracker{states: []*tokenState{snek, min, quiet}, now: func() time.Time { return testNow }}

	got := tr.Wallet("stake1me")
	if len(got) != 2 || got[0].ID != "min" || got[1].ID != "snek" {
		t.Fatalf("tokens: %+v", got)
	}
	if s := got[1]; s.Buys != 1 || s.Sells != 1 || s.BuyADA != 100 || s.SellADA != 40 || !s.Last.Equal(at(3)) {
		t.Errorf("SNEK: %+v", s)
	}
	if len(tr.Wallet("stake1nobody")) != 0 {
		t.Error("a wallet with no trades should have no tokens")
	}
	if WhaleADA(snek.quote) != 23_000 {
		t.Errorf("whale level = %v, want 23000 (1%% of the pool)", WhaleADA(snek.quote))
	}
}
