package market

import (
	"context"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cardano-analytics/internal/koios"
	"cardano-analytics/internal/minswap"
)

const (
	djedPolicy = "8db269c3ec630e06ae29f74bc39edd1f87c819f1056206e879a1cd61"
	djedName   = "446a65644d6963726f555344" // DjedMicroUSD
	shenName   = "5368656e4d6963726f555344" // ShenMicroUSD
)

// djedTracker serves the Djed reserve from a fake Koios (the numbers read on
// 10 Nov 2026) and returns a Tracker with DJED and SHEN listed, each with a
// supply of one trillion (6 decimals). assets is address_assets' answer.
func djedTracker(t *testing.T, assets string, calls *int) (*Tracker, *tokenState, *tokenState) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*calls++
		switch r.URL.Path {
		case "/asset_nft_address":
			if r.URL.Query().Get("_asset_name") != "446a6564537461626c65436f696e4e4654" {
				t.Errorf("asked for the holder of %s, want the DjedStableCoinNFT", r.URL.Query().Get("_asset_name"))
			}
			w.Write([]byte(`[{"payment_address":"addr1z8mcpc26j64fmhhd6sv5qj5mk9xqnfxgm6k"}]`))
		case "/address_assets":
			w.Write([]byte(assets))
		default:
			t.Errorf("unexpected request %s", r.URL.Path)
		}
	}))
	t.Cleanup(srv.Close)
	toks := []Token{
		{Ticker: "DJED", Decimals: 6, Asset: koios.Asset{PolicyID: djedPolicy, AssetName: djedName}},
		{Ticker: "SHEN", Decimals: 6, Asset: koios.Asset{PolicyID: djedPolicy, AssetName: shenName}},
	}
	tr := NewTracker(koios.New(srv.URL, "test-token"), toks, time.Minute, log.New(io.Discard, "", 0))
	tr.now = func() time.Time { return testNow }
	for _, s := range tr.states {
		s.poolRef, s.quote.ReserveADA = "ab#0", MinPoolADA
		s.supply, _ = new(big.Int).SetString("1000000000000000000", 10)
	}
	return tr, tr.states[0], tr.states[1]
}

// TestDjedSupplyLeavesOutTheReserve checks that DJED's and SHEN's supply is
// what was given out, not the trillion minted, with both answers Koios has
// given for address_assets (flat rows, and rows nested in asset_list).
func TestDjedSupplyLeavesOutTheReserve(t *testing.T) {
	flat := `[{"address":"a","policy_id":"` + djedPolicy + `","asset_name":"` + djedName + `","quantity":"999998589731000000"},
		{"address":"a","policy_id":"` + djedPolicy + `","asset_name":"` + shenName + `","quantity":"999968313614000000"},
		{"address":"a","policy_id":"` + djedPolicy + `","asset_name":"446a6564537461626c65436f696e4e4654","quantity":"1"}]`
	nested := `[{"address":"a","asset_list":[{"policy_id":"` + djedPolicy + `","asset_name":"` + djedName + `","quantity":"999998589731000000"},
		{"policy_id":"` + djedPolicy + `","asset_name":"` + shenName + `","quantity":"999968313614000000"}]}]`
	for name, answer := range map[string]string{"flat": flat, "nested": nested} {
		calls := 0
		tr, djed, shen := djedTracker(t, answer, &calls)
		tr.loadReserves(context.Background())
		if got := scaled(djed.issued(), 6); got != 1_410_269 {
			t.Errorf("%s: DJED given out %v, want 1,410,269", name, got)
		}
		if got := scaled(shen.issued(), 6); got != 31_686_386 {
			t.Errorf("%s: SHEN given out %v, want 31,686,386", name, got)
		}
		if calls != 2 {
			t.Errorf("%s: %d Koios requests, want 2", name, calls)
		}
	}
}

// TestReserveReadEveryTenMinutes checks that the reserve is not read again
// within ReservesEvery, and is after it.
func TestReserveReadEveryTenMinutes(t *testing.T) {
	calls := 0
	tr, _, _ := djedTracker(t, `[]`, &calls)
	tr.loadReserves(context.Background())
	tr.now = func() time.Time { return testNow.Add(ReservesEvery - time.Second) }
	tr.loadReserves(context.Background())
	if calls != 2 {
		t.Errorf("%d requests within %s, want 2", calls, ReservesEvery)
	}
	tr.now = func() time.Time { return testNow.Add(ReservesEvery) }
	tr.loadReserves(context.Background())
	if calls != 4 {
		t.Errorf("%d requests after %s, want 4", calls, ReservesEvery)
	}
}

// TestNoSupplyUntilTheReserveIsKnown checks that DJED shows no supply (so no
// FDV) before the reserve is read or when nothing of it is found there,
// instead of the trillion pre-minted.
func TestNoSupplyUntilTheReserveIsKnown(t *testing.T) {
	calls := 0
	tr, djed, _ := djedTracker(t, `[]`, &calls)
	if djed.issued() != nil {
		t.Error("supply shown before the reserve was read")
	}
	tr.loadReserves(context.Background())
	if djed.issued() != nil {
		t.Error("supply shown though no DJED was found in the reserve")
	}
}

// TestOtherTokensKeepTheirSupply checks that a token without a protocol
// reserve shows its on-chain total supply.
func TestOtherTokensKeepTheirSupply(t *testing.T) {
	s := &tokenState{token: Token{Asset: koios.Asset{PolicyID: "279c909f348e533da5808898f87f9a14bb2c3dfbbacccd631d927a3f", AssetName: "534e454b"}},
		supply: big.NewInt(76_715_880_000)}
	if got := s.issued(); got == nil || got.Int64() != 76_715_880_000 {
		t.Errorf("SNEK supply %v, want 76,715,880,000", got)
	}
}

// TestDecimalsFromTheTokenList checks that a pool Koios gave no decimals for
// takes the token list's (SUPERIOR: 8), and that Koios's decimals are kept
// when it has them.
func TestDecimalsFromTheTokenList(t *testing.T) {
	superior := Token{Ticker: "SUPERIOR", Decimals: 8}
	if p := superior.withDecimals(minswap.Pool{DecimalsB: 0}); p.DecimalsB != 8 {
		t.Errorf("SUPERIOR decimals %d, want 8 from the token list", p.DecimalsB)
	}
	if p := (Token{Decimals: 0}).withDecimals(minswap.Pool{DecimalsB: 6}); p.DecimalsB != 6 {
		t.Errorf("decimals %d, want Koios's 6", p.DecimalsB)
	}
}
