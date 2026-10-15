// Tests for splitting Minswap batch transactions into one trade per wallet.
package market

import (
	"math/big"
	"testing"

	"cardano-analytics/internal/koios"
	"cardano-analytics/internal/minswap"
)

var (
	snekAsset = koios.Asset{PolicyID: "279c909f348e533da5808898f87f9a14bb2c3dfbbacccd631d927a3f", AssetName: "534e454b"}
	snekLP    = koios.Asset{PolicyID: minswap.LPPolicy, AssetName: "aa"} // the token's pool
	otherLP   = koios.Asset{PolicyID: minswap.LPPolicy, AssetName: "bb"} // another pool, a step in a routed swap
	stable    = koios.Asset{PolicyID: "fe", AssetName: "55534441"}       // a stablecoin, e.g. USDA
)

// utxo makes one input or output: who holds it, its lovelace and its SNEK.
// who is "pool", "batcher", or a wallet's stake address (a payout).
func utxo(who string, lovelace, snek string) koios.TxOutput {
	var o koios.TxOutput
	switch who {
	case "pool", "other pool":
		o.PaymentAddr.Cred = minswap.PoolCredential
		o.PaymentAddr.Bech32 = PoolAddress
		o.StakeAddr = "stake1pool"
		lp := snekLP
		if who == "other pool" {
			lp = otherLP
		}
		o.AssetList = append(o.AssetList, koios.AssetAmount{PolicyID: lp.PolicyID, AssetName: lp.AssetName, Quantity: "9223367692996407324"})
	case "batcher":
		o.PaymentAddr.Bech32 = "addr1qbatcher"
		o.StakeAddr = "stake1batcher"
	default:
		o.PaymentAddr.Bech32 = "addr1q" + who
		o.StakeAddr = who
	}
	o.Value = lovelace
	if snek != "" {
		o.AssetList = append(o.AssetList, koios.AssetAmount{PolicyID: snekAsset.PolicyID, AssetName: snekAsset.AssetName, Quantity: snek})
	}
	return o
}

// plus adds an amount of another token to an input or output.
func plus(o koios.TxOutput, a koios.Asset, qty string) koios.TxOutput {
	o.AssetList = append(o.AssetList, koios.AssetAmount{PolicyID: a.PolicyID, AssetName: a.AssetName, Quantity: qty})
	return o
}

// order makes an order waiting at Minswap's order address. stake is the
// stake key on the order's address: often none, sometimes the owner's or
// Minswap's own, so it is never used to find the trader.
func order(stake string, lovelace, snek string) koios.TxOutput {
	o := utxo("x", lovelace, snek)
	o.PaymentAddr.Bech32 = "addr1w8p79rpminswaporder"
	if stake != "" {
		o.PaymentAddr.Bech32 = "addr1z8p79rpminswaporder" + stake
	}
	o.StakeAddr = stake
	return o
}

// check compares fills with the wanted wallet, lovelace into the pool and SNEK.
func check(t *testing.T, name string, got []fill, want [][3]string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: %d fills, want %d", name, len(got), len(want))
	}
	for i, w := range want {
		if got[i].wallet != w[0] || got[i].ada.String() != w[1] || got[i].token.String() != w[2] {
			t.Errorf("%s fill %d = %s %s %s, want %s %s %s", name, i, got[i].wallet, got[i].ada, got[i].token, w[0], w[1], w[2])
		}
	}
}

// TestFillsRealSwaps uses three SNEK swaps from mainnet (amounts as Koios
// showed them, to the cent): one buy, then the same wallet selling twice.
func TestFillsRealSwaps(t *testing.T) {
	// d4ec9ace…: a ₳491 order buys 136,172 SNEK; ₳487 reaches the pool.
	buy := koios.Tx{
		Inputs:  []koios.TxOutput{utxo("batcher", "83330000", ""), utxo("pool", "2296488550000", "648754593"), order("stake1u92s0s6x", "491000000", "")},
		Outputs: []koios.TxOutput{utxo("batcher", "84660000", ""), utxo("stake1u92s0s6x", "2000000", "136172"), utxo("pool", "2296975550000", "648618421")},
	}
	check(t, "buy", fillsFrom(buy, snekAsset, snekLP), [][3]string{{"stake1u92s0s6x", "487000000", "136172"}})

	// 24ce8738…: 293,760 SNEK sold; the pool pays ₳1,029.43.
	sell1 := koios.Tx{
		Inputs:  []koios.TxOutput{utxo("batcher", "276970000", ""), utxo("pool", "2296975550000", "648618421"), order("stake1u82zmwat", "4000000", "293760")},
		Outputs: []koios.TxOutput{utxo("batcher", "278310000", ""), utxo("stake1u82zmwat", "1031440000", ""), utxo("pool", "2295946120000", "648912181")},
	}
	check(t, "sell 1", fillsFrom(sell1, snekAsset, snekLP), [][3]string{{"stake1u82zmwat", "-1029430000", "-293760"}})

	// 3b0d0858…: the same wallet sells 472,160 SNEK for ₳1,652.68.
	sell2 := koios.Tx{
		Inputs:  []koios.TxOutput{utxo("batcher", "278310000", ""), utxo("pool", "2295946120000", "648912181"), order("stake1u82zmwat", "4000000", "472160")},
		Outputs: []koios.TxOutput{utxo("batcher", "279640000", ""), utxo("stake1u82zmwat", "1654680000", ""), utxo("pool", "2294293440000", "649384341")},
	}
	check(t, "sell 2", fillsFrom(sell2, snekAsset, snekLP), [][3]string{{"stake1u82zmwat", "-1652680000", "-472160"}})
}

// TestFillsOrderWithoutOwnerKey uses three more SNEK swaps from mainnet whose
// orders do not carry the owner's stake key: two have none, one has the
// pool's. The trader is the wallet the payout goes to.
func TestFillsOrderWithoutOwnerKey(t *testing.T) {
	// 2b4897bc…: a ₳2,531.02 order with no stake key buys 725,832 SNEK.
	buy := koios.Tx{
		Inputs:  []koios.TxOutput{utxo("batcher", "257170000", ""), order("", "2531020000", ""), utxo("pool", "2265164020000", "657941463")},
		Outputs: []koios.TxOutput{utxo("stake1u83ekh6q", "2000000", "725832"), utxo("batcher", "258500000", ""), utxo("pool", "2267691040000", "657215631")},
	}
	check(t, "no key, buy", fillsFrom(buy, snekAsset, snekLP), [][3]string{{"stake1u83ekh6q", "2527020000", "725832"}})

	// 28b7b235…: the order carries the pool's stake key; 11,831 SNEK sold.
	sell := koios.Tx{
		Inputs:  []koios.TxOutput{utxo("batcher", "258500000", ""), utxo("pool", "2267677680000", "657189598"), order("stake1pool", "4000000", "11831")},
		Outputs: []koios.TxOutput{utxo("batcher", "259840000", ""), utxo("stake1u97878jt", "42410000", ""), utxo("pool", "2267637260000", "657201429")},
	}
	check(t, "pool key, sell", fillsFrom(sell, snekAsset, snekLP), [][3]string{{"stake1u97878jt", "-40420000", "-11831"}})

	// 48dcf4b7…: no stake key; 1,659,809 SNEK sold for ₳5,653.10.
	sell2 := koios.Tx{
		Inputs:  []koios.TxOutput{utxo("batcher", "279900000", ""), order("", "4000000", "1659809"), utxo("pool", "2267124990000", "657351429")},
		Outputs: []koios.TxOutput{utxo("batcher", "281240000", ""), utxo("stake1u8g9z7uq", "5655100000", ""), utxo("pool", "2261471890000", "659011238")},
	}
	check(t, "no key, sell", fillsFrom(sell2, snekAsset, snekLP), [][3]string{{"stake1u8g9z7uq", "-5653100000", "-1659809"}})
}

// TestFillsRouted uses three NIGHT swaps from mainnet routed through a second
// pool: a stablecoin is swapped to ADA, then ADA to NIGHT (or the other way).
// NIGHT's own pool change is the wallet's trade in NIGHT.
func TestFillsRouted(t *testing.T) {
	// b0b13bb9…: 500 USDA → ₳2,250.46 → 11,893.598436 NIGHT.
	buy := koios.Tx{
		Inputs: []koios.TxOutput{
			utxo("batcher", "170070000", ""), plus(utxo("other pool", "1645663990000", ""), stable, "364033069925"),
			utxo("pool", "3644152320000", "19329106017161"), plus(order("stake1u90uyty0", "4000000", ""), stable, "500000000"),
		},
		Outputs: []koios.TxOutput{
			utxo("batcher", "171300000", ""), utxo("stake1u90uyty0", "2000000", "11893598436"),
			plus(utxo("other pool", "1643413530000", ""), stable, "364533069925"), utxo("pool", "3646402780000", "19317212418725"),
		},
	}
	check(t, "routed buy", fillsFrom(buy, snekAsset, snekLP), [][3]string{{"stake1u90uyty0", "2250460000", "11893598436"}})

	// e1d5106c…: 252 USDCx → NIGHT (through a NIGHT/USDCx pool, which holds
	// NIGHT too) → ₳1,134.57 through NIGHT's own pool.
	sell := koios.Tx{
		Inputs: []koios.TxOutput{
			utxo("batcher", "298470000", ""), utxo("pool", "3646536950000", "19316503779124"),
			plus(utxo("other pool", "4500000", "2860272940307"), stable, "118923095327"), plus(order("stake1u9hnaaxk", "4000000", ""), stable, "252000000"),
		},
		Outputs: []koios.TxOutput{
			utxo("batcher", "299700000", ""), utxo("stake1u9hnaaxk", "1136570000", ""),
			plus(utxo("other pool", "4500000", "2854242926136"), stable, "119175095327"), utxo("pool", "3645402380000", "19322533793295"),
		},
	}
	check(t, "routed sell", fillsFrom(sell, snekAsset, snekLP), [][3]string{{"stake1u9hnaaxk", "-1134570000", "-6030014171"}})
}

// TestFillsBatch checks a batch of three orders from three wallets (two buys
// and a sell, ₳2 fee each) becomes three fills that add up to the pool's
// change, each paired with its payout.
func TestFillsBatch(t *testing.T) {
	// The pool holds ₳5,000,000 and 25,000,000 tokens: ₳0.2 per token.
	tx := koios.Tx{
		Inputs: []koios.TxOutput{
			utxo("batcher", "100000000", ""), utxo("pool", "5000000000000", "25000000"),
			order("", "102000000", ""), order("stake1pool", "4000000", "750"), order("stake1c", "54000000", ""),
		},
		Outputs: []koios.TxOutput{
			utxo("stake1a", "2000000", "490"), utxo("stake1b", "152000000", ""), utxo("stake1c", "2000000", "250"),
			utxo("batcher", "105500000", ""), utxo("pool", "4999998000000", "25000010"),
		},
	}
	check(t, "batch", fillsFrom(tx, snekAsset, snekLP), [][3]string{
		{"stake1a", "98000000", "490"}, {"stake1b", "-150000000", "-750"}, {"stake1c", "50000000", "250"},
	})
}

// TestFillsNotSplit checks transactions that cannot be split for certain give
// no fills, so the swap is shown as one trade without a wallet.
func TestFillsNotSplit(t *testing.T) {
	pool := utxo("pool", "1000000000000", "5000000000") // ₳0.2 per token
	cases := map[string]koios.Tx{
		"no inputs (old data)": {Outputs: []koios.TxOutput{pool}},
		"tokens missing": {
			Inputs:  []koios.TxOutput{order("", "102000000", ""), pool},
			Outputs: []koios.TxOutput{utxo("stake1a", "2000000", "100"), utxo("pool", "1000098000000", "4999999000")},
		},
		"fee too big": {
			Inputs:  []koios.TxOutput{order("", "500000000", ""), pool},
			Outputs: []koios.TxOutput{utxo("stake1a", "2000000", "490"), utxo("pool", "1000098000000", "4999999510")},
		},
		"an order without a payout": {
			Inputs:  []koios.TxOutput{order("", "102000000", ""), order("", "54000000", ""), pool},
			Outputs: []koios.TxOutput{utxo("stake1a", "2000000", "490"), utxo("pool", "1000098000000", "4999999510")},
		},
		"two orders through two pools": {
			Inputs:  []koios.TxOutput{order("", "102000000", ""), order("", "54000000", ""), utxo("other pool", "1000000", ""), pool},
			Outputs: []koios.TxOutput{utxo("stake1a", "2000000", "490"), utxo("stake1b", "2000000", "250"), utxo("other pool", "1000000", ""), utxo("pool", "1000148000000", "4999999260")},
		},
		// Two buys (₳98 and ₳998) with the payouts the wrong way round: each
		// pair would trade far from the pool's price.
		"payouts swapped": {
			Inputs: []koios.TxOutput{order("", "102000000", ""), order("", "1002000000", ""), pool},
			Outputs: []koios.TxOutput{utxo("stake1a", "2000000", "4990"), utxo("stake1b", "2000000", "490"),
				utxo("pool", "1001096000000", "4999994520")},
		},
	}
	for name, tx := range cases {
		if f := fillsFrom(tx, snekAsset, snekLP); f != nil {
			t.Errorf("%s: got %d fills, want none", name, len(f))
		}
	}
}

// TestTradesFromFills checks fills become trades with wallets, and a swap
// without fills still becomes one trade from the pool's change.
func TestTradesFromFills(t *testing.T) {
	p := minswap.Pool{UTxO: "abc#0", DecimalsB: 0, LastSwap: testNow}
	fills := []fill{{"stake1a", big.NewInt(98_000_000), big.NewInt(25_000)}, {"stake1b", big.NewInt(-150_000_000), big.NewInt(-40_000)}}
	got := tradesFrom(p, fills, big.NewInt(-52_000_000), big.NewInt(15_000))
	if len(got) != 2 || got[0].Wallet != "stake1a" || got[0].Side != "buy" || got[0].AmountADA != 98 ||
		got[1].Wallet != "stake1b" || got[1].Side != "sell" || got[1].AmountToken != 40_000 || got[1].TxHash != "abc" {
		t.Errorf("trades from fills: %+v", got)
	}
	// No fills: the pool got ₳98 and gave 25,000 tokens, a buy with no wallet.
	got = tradesFrom(p, nil, big.NewInt(98_000_000), big.NewInt(-25_000))
	if len(got) != 1 || got[0].Side != "buy" || got[0].Wallet != "" || got[0].AmountADA != 98 {
		t.Errorf("trade from the pool's change: %+v", got)
	}
}
