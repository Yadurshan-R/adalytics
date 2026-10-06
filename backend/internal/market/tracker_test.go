package market

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"cardano-analytics/internal/koios"
	"cardano-analytics/internal/minswap"
)

var testNow = time.Unix(1791180000, 0)

func loadPools(t *testing.T) []koios.UTxO {
	t.Helper()
	data, err := os.ReadFile("../minswap/testdata/snek_pools.json")
	if err != nil {
		t.Fatal(err)
	}
	var pools []koios.UTxO
	if err := json.Unmarshal(data, &pools); err != nil {
		t.Fatal(err)
	}
	return pools
}

// poolOutput is the real SNEK/ADA pool from testdata with a chosen ADA reserve.
func poolOutput(t *testing.T, txHash, reserveADA string) koios.TxOutput {
	p := loadPools(t)[1]
	out := koios.TxOutput{TxHash: txHash, TxIndex: 0, AssetList: p.AssetList}
	out.PaymentAddr.Cred = minswap.PoolCredential
	out.PaymentAddr.Bech32 = PoolAddress
	datum := *p.InlineDatum
	datum.Value.Fields = append([]koios.PlutusData(nil), datum.Value.Fields...)
	n := json.Number(reserveADA)
	datum.Value.Fields[4] = koios.PlutusData{Int: &n}
	out.InlineDatum = &datum
	return out
}

// swapOutput is the SNEK/ADA pool with chosen LP supply and reserves (SNEK has 0 decimals).
func swapOutput(t *testing.T, txHash, liquidity, reserveADA, reserveSNEK string) koios.TxOutput {
	out := poolOutput(t, txHash, reserveADA)
	fields := out.InlineDatum.Value.Fields
	l, b := json.Number(liquidity), json.Number(reserveSNEK)
	fields[3] = koios.PlutusData{Int: &l}
	fields[5] = koios.PlutusData{Int: &b}
	return out
}

// otherOutput is an output of some other Minswap pool (not one of ours).
func otherOutput(t *testing.T, txHash string) koios.TxOutput {
	p := loadPools(t)[0] // SNEK/MIN pool
	out := koios.TxOutput{TxHash: txHash, AssetList: p.AssetList, InlineDatum: p.InlineDatum}
	out.PaymentAddr.Cred = minswap.PoolCredential
	return out
}

type fakeTx struct {
	hash   string
	height int
	time   int64
	outs   []koios.TxOutput
}

type fake struct {
	t       *testing.T
	mu      sync.Mutex
	txs     []fakeTx
	tip     int
	spent   bool // what utxo_info answers next
	calls   []string
	queries []string     // _after_block_height of each address_txs call
	pools   []koios.UTxO // what asset_utxos answers: the pools as they are now
	bodies  []int        // size of each asset_utxos request body
}

func (f *fake) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if r.Header.Get("Authorization") != "Bearer test-token" {
		f.t.Errorf("missing bearer token on %s", r.URL.Path)
	}
	body, _ := io.ReadAll(r.Body)
	f.calls = append(f.calls, r.URL.Path)

	switch r.URL.Path {
	case "/asset_info":
		// 76,715,880,000 SNEK (0 decimals) and a tiny PNG logo.
		w.Write([]byte(`[{"policy_id":"279c909f348e533da5808898f87f9a14bb2c3dfbbacccd631d927a3f","asset_name":"534e454b","fingerprint":"asset108xu02ckwrfc8qs9d97mgyh4kn8gdu9w8f5sxk","creation_time":1682340000,"total_supply":"76715880000","token_registry_metadata":{"decimals":0,"logo":"iVBORw0KGgo=","description":"  The snek token.  ","url":"https://snek.com"}}]`))
	case "/asset_summary":
		w.Write([]byte(`[{"staked_wallets":38151,"unstaked_addresses":1363}]`))
	case "/tip":
		json.NewEncoder(w).Encode([]koios.Tip{{BlockHeight: f.tip}})
	case "/address_txs":
		var req struct {
			Addresses []string `json:"_addresses"`
			After     int      `json:"_after_block_height"`
		}
		json.Unmarshal(body, &req)
		if len(req.Addresses) == 0 || req.Addresses[0] != PoolAddress {
			f.t.Errorf("address_txs not asked for the Minswap pool address: %v", req.Addresses)
		}
		f.queries = append(f.queries, strings.TrimSpace(strings.Split(string(body), "_after_block_height\":")[1][:8]))
		var rows []koios.AddressTx
		for i := len(f.txs) - 1; i >= 0; i-- {
			if f.txs[i].height >= req.After {
				rows = append(rows, koios.AddressTx{TxHash: f.txs[i].hash, BlockHeight: f.txs[i].height})
			}
		}
		json.NewEncoder(w).Encode(rows)
	case "/tx_info":
		var req struct {
			Hashes []string `json:"_tx_hashes"`
		}
		json.Unmarshal(body, &req)
		var out []koios.Tx
		for _, h := range req.Hashes {
			for _, tx := range f.txs {
				if tx.hash == h {
					out = append(out, koios.Tx{TxHash: h, BlockHeight: tx.height, TxTimestamp: tx.time, Outputs: tx.outs})
				}
			}
		}
		json.NewEncoder(w).Encode(out)
	case "/utxo_info":
		var req struct {
			Refs []string `json:"_utxo_refs"`
		}
		json.Unmarshal(body, &req)
		hash, _, _ := strings.Cut(req.Refs[0], "#")
		json.NewEncoder(w).Encode([]map[string]any{{"tx_hash": hash, "tx_index": 0, "is_spent": f.spent}})
	case "/asset_utxos":
		if got := r.URL.Query().Get("address"); got != "eq."+PoolAddress {
			f.t.Errorf("asset_utxos not filtered to the pool address: %q", got)
		}
		f.bodies = append(f.bodies, len(body))
		pools := f.pools
		if pools == nil {
			pools = []koios.UTxO{}
		}
		json.NewEncoder(w).Encode(pools)
	default:
		f.t.Errorf("unexpected request %s", r.URL.Path)
	}
}

// snekToken is SNEK from tokens.json.
func snekToken(t *testing.T) Token {
	for _, tok := range Tokens {
		if tok.Asset.AssetName == "534e454b" && strings.HasPrefix(tok.Asset.PolicyID, "279c909f") {
			return tok
		}
	}
	t.Fatal("SNEK not in tokens.json")
	return Token{}
}

func newTracker(t *testing.T, f *fake) *Tracker {
	srv := httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(srv.Close)
	tr := NewTracker(koios.New(srv.URL, "test-token"), []Token{snekToken(t)}, time.Minute, log.New(io.Discard, "", 0))
	tr.now = func() time.Time { return testNow }
	return tr
}

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
	_, points, err := tr.Chart("snek", "24h")
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

	want := "/tip /asset_utxos /address_txs /tx_info /asset_info /tip /utxo_info /tip /utxo_info /address_txs /tx_info /asset_summary"
	if got := strings.Join(f.calls, " "); got != want {
		t.Errorf("requests:\n got  %s\n want %s", got, want)
	}
	// History from 25 hours back; the live read only from the previous check.
	if len(f.queries) != 2 || f.queries[0] != "14025500" || f.queries[1] != "14029997" {
		t.Errorf("address_txs from heights %v, want [14025500 14029997]", f.queries)
	}
}

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
	_, points, _ := tr.Chart("snek", "24H")
	if len(points) != 289 || points[0].Price != q.PriceADA {
		t.Errorf("chart should be flat: %d points", len(points))
	}
}

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

func TestTokenWithoutPoolIsLeftOut(t *testing.T) {
	f := &fake{t: t, tip: 14030000}
	tr := newTracker(t, f)
	tr.refresh(context.Background())
	if q := tr.Quotes(); len(q) != 0 {
		t.Fatalf("a token without a Minswap pool should not be listed: %+v", q)
	}
	if got := strings.Join(f.calls, " "); got != "/tip /asset_utxos /address_txs" {
		t.Errorf("requests: %s", got)
	}
}

func TestFindsAllPoolsInSmallRequests(t *testing.T) {
	f := &fake{t: t, tip: 14030000}
	srv := httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(srv.Close)
	tr := NewTracker(koios.New(srv.URL, "test-token"), Tokens, time.Minute, log.New(io.Discard, "", 0))
	if _, err := tr.findPools(context.Background()); err != nil {
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

func TestBucketCarriesPriceForward(t *testing.T) {
	base := time.Unix(1_000_200, 0) // a multiple of 300 seconds
	pts := []Point{
		{Time: base.Unix() - 1000, Price: 1},
		{Time: base.Unix() + 310, Price: 2, VolumeADA: 5},
		{Time: base.Unix() + 320, Price: 3, VolumeADA: 7},
	}
	got := bucket(pts, base, base.Add(15*time.Minute), 5*time.Minute)
	want := []ChartPoint{
		{Time: base.Unix(), Price: 1},
		{Time: base.Unix() + 300, Price: 3, Volume: 12},
		{Time: base.Unix() + 600, Price: 3},
		{Time: base.Unix() + 900, Price: 3},
	}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("point %d: got %v, want %v", i, got[i], want[i])
		}
	}
}

func TestFormatADA(t *testing.T) {
	cases := map[float64]string{0.002657024: "₳0.002657", 0.1934: "₳0.1934", 1.5: "₳1.5000", 0.00000123456: "₳0.000001235"}
	for in, want := range cases {
		if got := FormatADA(in); got != want {
			t.Errorf("FormatADA(%v) = %s, want %s", in, got, want)
		}
	}
}
