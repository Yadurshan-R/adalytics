// Tests for pairing Minswap orders with their payouts, using a real batch.
package market

import (
	"encoding/json"
	"math/big"
	"os"
	"testing"

	"cardano-analytics/internal/koios"
	"cardano-analytics/internal/minswap"
)

var (
	nightAsset = koios.Asset{PolicyID: "0691b2fecca1ac4f53cb6dfb00b7013e561d1f34403b957cbb5af1fa", AssetName: "4e49474854"}
	nightLP    = koios.Asset{PolicyID: minswap.LPPolicy, AssetName: "e74c52975908a612d5ce68327040d449aae99f8b463bb6de046a1b23c5713169"}
)

// nightBatch is a real mainnet transaction (fa72d267…, 4 Nov 2026): a batcher
// swapped 7 NIGHT sell orders from 6 wallets in one go. Koios lists its
// outputs out of order (#3, #8, #1, #2, #4, #0, #5, #6, #7), so orders and
// payouts cannot be paired by position.
func nightBatch(t *testing.T) koios.Tx {
	t.Helper()
	data, err := os.ReadFile("testdata/night_batch.json")
	if err != nil {
		t.Fatal(err)
	}
	var txs []koios.Tx
	if err := json.Unmarshal(data, &txs); err != nil {
		t.Fatal(err)
	}
	return txs[0]
}

// TestFillsRealBatch checks that every order of the real batch goes to the
// wallet its datum names, with the exact amounts Cardanoscan shows (order 1:
// 8,908.925276 NIGHT sold for ₳1,990.346534), and that together they add up
// to the pool's change.
func TestFillsRealBatch(t *testing.T) {
	tx := nightBatch(t)
	got := fillsFrom(tx, nightAsset, nightLP)
	// wallet, lovelace into the pool (negative: a sell got ADA), NIGHT units into the pool's opposite.
	check(t, "night batch", got, [][3]string{
		{"stake1ux5fvfh89gmzfdv2fcj3nz8zmfv0wst87eysnwjvjc4aqks34hj8f", "-221765218", "-972537262"},
		{"stake1u8tpsrkceradm3mdrgynnfjpezvy9np2e2d7jynfvpjd63cnsdjge", "-5132080359", "-22538280850"},
		{"stake1u8tpsrkceradm3mdrgynnfjpezvy9np2e2d7jynfvpjd63cnsdjge", "-5118170238", "-22538280850"},
		{"stake1u8fqs5wqpyuw0lulwp4tfuc45tkrzws0qllw4lsx7me79ncy6wpzd", "-10107967408", "-44691365647"},
		{"stake1uxwexh0uy4wdgmzjzv0pj5kq39563wmasvkpjkwexwy3yxgqdajat", "-9056386652", "-40246423613"},
		{"stake1u84a7hzcjgg3fyz87fzkcaaq0qqjh8s8r66km6zq395t7kg09xn2g", "-7964108709", "-35553249464"},
		{"stake1u9tx9n7w796zyqjtdj8fsg3cj7r59khdvfr72g69xjf0mnqtmss5r", "-1990346534", "-8908925276"},
	})

	// Together the 7 trades are exactly what left and entered the pool.
	ada, tok := new(big.Int), new(big.Int)
	for _, f := range got {
		ada.Add(ada, f.ada)
		tok.Add(tok, f.token)
	}
	in, out := tx.Inputs[findPool(tx.Inputs, nightLP)], tx.Outputs[findPool(tx.Outputs, nightLP)]
	if want := sub(lovelace(out), lovelace(in)); ada.Cmp(want) != 0 {
		t.Errorf("ADA adds up to %s, pool changed by %s", ada, want)
	}
	if want := sub(units(in, nightAsset), units(out, nightAsset)); tok.Cmp(want) != 0 {
		t.Errorf("NIGHT adds up to %s, pool changed by %s", tok, want)
	}
}

// TestFillsRealBatchUnsure changes the real batch in ways that make the
// pairing uncertain, and checks that it is then not split (shown as one
// trade without a wallet) instead of naming a wrong wallet.
func TestFillsRealBatchUnsure(t *testing.T) {
	orderAt := func(tx koios.Tx) int { // the first order input
		for i, o := range tx.Inputs {
			if o.InlineDatum != nil && !isPool(o) {
				return i
			}
		}
		t.Fatal("no order in the batch")
		return -1
	}
	cases := map[string]func(tx *koios.Tx){
		"order for another pool": func(tx *koios.Tx) {
			other := "aa"
			tx.Inputs[orderAt(*tx)].InlineDatum.Value.Fields[5].Fields[1].Bytes = &other
		},
		"payout to a wallet no order names": func(tx *koios.Tx) {
			for i, o := range tx.Outputs {
				if o.PaymentAddr.Cred == "2a649f44bd03b1d05ac934e48511aeee3880ab362f736ace62f26a03" {
					tx.Outputs[i].PaymentAddr.Cred = "ff"
				}
			}
		},
		"one order without a datum": func(tx *koios.Tx) { tx.Inputs[orderAt(*tx)].InlineDatum = nil },
	}
	for name, change := range cases {
		tx := nightBatch(t)
		change(&tx)
		if got := fillsFrom(tx, nightAsset, nightLP); got != nil {
			t.Errorf("%s: split into %d trades, want not split", name, len(got))
		}
	}
}

// TestReadOrder reads the receiver and pool from a real order's datum.
func TestReadOrder(t *testing.T) {
	tx := nightBatch(t)
	for _, o := range tx.Inputs {
		if o.TxHash[:8] != "2f752703" {
			continue
		}
		got, ok := minswap.ReadOrder(o.InlineDatum)
		if !ok || got.Receiver != "2a649f44bd03b1d05ac934e48511aeee3880ab362f736ace62f26a03" || got.LPAsset != nightLP {
			t.Errorf("ReadOrder = %+v, %v", got, ok)
		}
		return
	}
	t.Fatal("order 2f752703 not in the batch")
}
