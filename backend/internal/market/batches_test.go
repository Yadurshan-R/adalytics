// Tests for batches where one wallet placed several orders, from 5 real
// mainnet transactions. There, only the batcher's own list (the batching
// redeemer) says which payout belongs to which order.
package market

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"testing"
	"time"

	"cardano-analytics/internal/koios"
	"cardano-analytics/internal/minswap"
)

// multiOrderBatches reads testdata/multi_order_batches.json: 5 real batches
// (4 Nov 2026) in which one wallet placed several orders, each with the
// batcher's list (as TxContracts returns it; only that script is kept).
func multiOrderBatches(t *testing.T) map[string]koios.Tx {
	t.Helper()
	data, err := os.ReadFile("testdata/multi_order_batches.json")
	if err != nil {
		t.Fatal(err)
	}
	var txs []koios.Tx
	if err := json.Unmarshal(data, &txs); err != nil {
		t.Fatal(err)
	}
	out := map[string]koios.Tx{}
	for _, tx := range txs {
		out[tx.TxHash[:8]] = tx
	}
	return out
}

// poolAsset finds the token and LP asset of the pool a transaction swapped in.
func poolAsset(t *testing.T, tx koios.Tx) (koios.Asset, koios.Asset) {
	t.Helper()
	for _, o := range tx.Inputs {
		if isPool(o) {
			p, err := minswap.Parse(o.UTxO(tx))
			if err != nil {
				t.Fatal(err)
			}
			return p.AssetB, p.LPAsset
		}
	}
	t.Fatal("no pool in the transaction")
	return koios.Asset{}, koios.Asset{}
}

// TestFillsSeveralOrdersOneWallet checks every trade of the 5 batches against
// the batcher's list: the wallet, the ADA into the pool and the tokens, in the
// order they were swapped. Each was checked by hand: every payout goes to the
// wallet its order names and trades within 3% of the pool's price.
func TestFillsSeveralOrdersOneWallet(t *testing.T) {
	want := map[string][][3]string{
		"5963986f": { // USDM: two wallets buying, two selling twice each
			{"stake1u9zjjxgaajfnpfp7zw2s7z5e9ycyl64ucd8l7h6f2qfj6ksntdr4d", "9533700000", "1781754618"},
			{"stake1u83ekh6q42zlhsfp5cjawaa8wmk2rj6vjg6zd9yun97cs2q26390l", "-874872090", "-165252263"},
			{"stake1u9tx9n7w796zyqjtdj8fsg3cj7r59khdvfr72g69xjf0mnqtmss5r", "-2006803133", "-379620000"},
			{"stake1u83ekh6q42zlhsfp5cjawaa8wmk2rj6vjg6zd9yun97cs2q26390l", "-4767072273", "-904911364"},
			{"stake1u9zjjxgaajfnpfp7zw2s7z5e9ycyl64ucd8l7h6f2qfj6ksntdr4d", "9900000000", "1846425558"},
			{"stake1u9tx9n7w796zyqjtdj8fsg3cj7r59khdvfr72g69xjf0mnqtmss5r", "-1471855898", "-277479953"},
		},
		"012bde68": { // NIGHT: one wallet selling 3 times
			{"stake1u8tpsrkceradm3mdrgynnfjpezvy9np2e2d7jynfvpjd63cnsdjge", "-9194418335", "-40000000000"},
			{"stake1u8tpsrkceradm3mdrgynnfjpezvy9np2e2d7jynfvpjd63cnsdjge", "-5241917493", "-22892515680"},
			{"stake1u8tpsrkceradm3mdrgynnfjpezvy9np2e2d7jynfvpjd63cnsdjge", "-5227318029", "-22892515680"},
		},
		"32433561": { // NIGHT: output #0 is an order the batcher killed and refunded, so no trade
			{"stake1u9szf52e8f0a97am7x5ua6xh27anm6a99m2f6xgjnx5yavcah2v20", "4015630535", "17445399978"},
			{"stake1u97878jtd82lxn74kha7s9s84zea7lre52zqrd8kqjyrnsqqfnd39", "-48833209", "-213203406"},
			{"stake1u8x35ez0geeqjftf58jx29srgvj76qzw2xmnlry5uwfzrcckwtt9p", "-8990288831", "-39345709541"},
			{"stake1u8x35ez0geeqjftf58jx29srgvj76qzw2xmnlry5uwfzrcckwtt9p", "-22061403569", "-97355260195"},
		},
		"debe3afe": { // NIGHT: a buyer without a stake key, named by its address
			{"stake1u97878jtd82lxn74kha7s9s84zea7lre52zqrd8kqjyrnsqqfnd39", "-24829507", "-110861644"},
			{"addr1vyvqyqgpahejmllmm373ngrwywz935dtk7yvt8sjfea5k3sdfg07d", "11125983180", "49232474418"},
			{"stake1u97878jtd82lxn74kha7s9s84zea7lre52zqrd8kqjyrnsqqfnd39", "-63642738", "-282476810"},
		},
		"baa70c23": { // NIGHT: one wallet selling 3 times
			{"stake1u8x35ez0geeqjftf58jx29srgvj76qzw2xmnlry5uwfzrcckwtt9p", "-13477974467", "-59291499822"},
			{"stake1u8x35ez0geeqjftf58jx29srgvj76qzw2xmnlry5uwfzrcckwtt9p", "-5538975940", "-24490474718"},
			{"stake1u8x35ez0geeqjftf58jx29srgvj76qzw2xmnlry5uwfzrcckwtt9p", "-3876910397", "-17184817624"},
		},
	}
	txs := multiOrderBatches(t)
	for id, w := range want {
		tx, ok := txs[id]
		if !ok {
			t.Fatalf("%s not in the test data", id)
		}
		if !needsBatchOrder(tx) {
			t.Errorf("%s: needsBatchOrder = false, want true", id)
		}
		asset, lp := poolAsset(t, tx)
		check(t, id, fillsFrom(tx, asset, lp), w)
	}
}

// TestFillsSeveralOrdersWithoutList checks that without the batcher's list
// (if fetching it failed) such a batch is not split: one trade without a
// wallet, never a guess.
func TestFillsSeveralOrdersWithoutList(t *testing.T) {
	for id, tx := range multiOrderBatches(t) {
		tx.Contracts = nil
		asset, lp := poolAsset(t, tx)
		if got := fillsFrom(tx, asset, lp); got != nil {
			t.Errorf("%s: split into %d trades without the batcher's list, want not split", id, len(got))
		}
	}
}

// TestBatchOrder reads the batcher's list from a real batch: output k pays
// the order at position order[k] among the orders sorted by UTxO.
func TestBatchOrder(t *testing.T) {
	got, ok := minswap.BatchOrder(multiOrderBatches(t)["012bde68"].Contracts)
	if !ok || len(got) != 3 || got[0] != 2 || got[1] != 0 || got[2] != 1 {
		t.Errorf("BatchOrder = %v, %v; want [2 0 1]", got, ok)
	}
}

// TestAddBatchOrders checks that the batcher's lists are fetched only for the
// transactions that need them, in one request, and attached to the right ones.
func TestAddBatchOrders(t *testing.T) {
	stored := multiOrderBatches(t)
	var asked [][]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Hashes []string `json:"_tx_hashes"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		asked = append(asked, req.Hashes)
		var out []koios.Tx
		for _, h := range req.Hashes {
			out = append(out, koios.Tx{TxHash: h, Contracts: stored[h[:8]].Contracts})
		}
		json.NewEncoder(w).Encode(out)
	}))
	t.Cleanup(srv.Close)
	tr := NewTracker(koios.New(srv.URL, "test-token"), nil, time.Minute, log.New(io.Discard, "", 0))

	txs := []koios.Tx{{TxHash: "no-orders"}} // a transaction that does not need it
	var want []string
	for _, tx := range stored {
		tx.Contracts = nil
		txs = append(txs, tx)
		want = append(want, tx.TxHash)
	}
	tr.addBatchOrders(context.Background(), txs)

	if len(asked) != 1 || len(asked[0]) != len(want) {
		t.Fatalf("asked for %v, want one request for the %d batches", asked, len(want))
	}
	sort.Strings(asked[0])
	sort.Strings(want)
	for i := range want {
		if asked[0][i] != want[i] {
			t.Errorf("asked for %s, want %s", asked[0][i], want[i])
		}
	}
	for _, tx := range txs[1:] {
		if _, ok := minswap.BatchOrder(tx.Contracts); !ok {
			t.Errorf("%s: no batcher's list attached", tx.TxHash[:8])
		}
	}
}
