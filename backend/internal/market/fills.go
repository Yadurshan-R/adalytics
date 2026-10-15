package market

import (
	"math/big"

	"cardano-analytics/internal/koios"
	"cardano-analytics/internal/minswap"
)

// Minswap does not swap each order on its own: a batcher takes several
// people's orders and swaps them against the pool in one transaction. A fill
// is one wallet's part of such a transaction.
//
// A transaction holds the pool (in and out), the batcher's own coins (in and
// out, at the same address), the orders being swapped and one payout per
// order. The wallet that traded is the one the payout goes to: an order's own
// address often has no stake key, or Minswap's. Orders and payouts come in
// the same order, so the first order belongs to the first payout, and so on.
//
// What a wallet put in or got back also holds the batcher's fee (taken off
// every order) and the order deposit (paid back). The deposit cancels out; the
// fee is worked out from the transaction itself: it is whatever ADA the
// orders gave up that did not end up in the pool.
type fill struct {
	wallet string   // stake address, or the payment address of a wallet without one
	ada    *big.Int // lovelace into the pool (a buy) or out of it (a sell, negative)
	token  *big.Int // token units the wallet got (a buy) or gave (a sell, negative)
}

// MaxBatchFeeADA is the most batcher fee one order can pay. A larger "fee"
// means the transaction is not a plain batch of swaps, so it is not split.
const MaxBatchFeeADA = 10

// fillsFrom splits a swap transaction into one fill per order. It returns nil
// when the transaction cannot be split with certainty (no inputs, more than
// one pool, orders and payouts that do not pair up, or amounts that do not
// add up); the caller then shows the pool's change as one trade without a
// wallet.
func fillsFrom(tx koios.Tx, asset koios.Asset) []fill {
	poolIn, poolOut := onePool(tx.Inputs), onePool(tx.Outputs)
	if poolIn < 0 || poolOut < 0 {
		return nil
	}
	in, out := tx.Inputs[poolIn], tx.Outputs[poolOut]
	poolADA := sub(lovelace(out), lovelace(in))
	poolTok := sub(units(out, asset), units(in, asset))

	// The batcher's coins go in and come back out at the same address; every
	// other input is an order, every other output a payout.
	inAddr, outAddr := map[string]bool{}, map[string]bool{}
	for _, o := range tx.Inputs {
		inAddr[o.PaymentAddr.Bech32] = true
	}
	for _, o := range tx.Outputs {
		outAddr[o.PaymentAddr.Bech32] = true
	}
	var orders, payouts []koios.TxOutput
	for i, o := range tx.Inputs {
		if i != poolIn && !outAddr[o.PaymentAddr.Bech32] {
			orders = append(orders, o)
		}
	}
	for i, o := range tx.Outputs {
		if i != poolOut && !inAddr[o.PaymentAddr.Bech32] {
			payouts = append(payouts, o)
		}
	}
	n := len(orders)
	if n == 0 || n != len(payouts) {
		return nil
	}

	// Each order's net change (payout minus order), and the fee: the ADA the
	// orders gave up that the pool did not get.
	net := make([]*big.Int, n)
	sumADA, sumTok := new(big.Int), new(big.Int)
	for i := range orders {
		net[i] = sub(lovelace(payouts[i]), lovelace(orders[i]))
		sumADA.Add(sumADA, net[i])
		sumTok.Add(sumTok, sub(units(payouts[i], asset), units(orders[i], asset)))
	}
	// Every token the pool gave or took must come from or go to an order.
	if new(big.Int).Add(sumTok, poolTok).Sign() != 0 {
		return nil
	}
	fee := new(big.Int).Neg(new(big.Int).Add(sumADA, poolADA))
	if fee.Sign() < 0 || fee.Cmp(big.NewInt(int64(n)*MaxBatchFeeADA*1_000_000)) > 0 {
		return nil
	}
	each, rest := new(big.Int).QuoRem(fee, big.NewInt(int64(n)), new(big.Int))

	fills := make([]fill, 0, n)
	for i := range orders {
		f := new(big.Int).Set(each)
		if i == 0 {
			f.Add(f, rest)
		}
		// Into the pool: what the order gave up, less its share of the fee.
		ada := new(big.Int).Neg(new(big.Int).Add(net[i], f))
		tok := sub(units(payouts[i], asset), units(orders[i], asset))
		// A buy pays ADA for tokens; a sell gets ADA for tokens. With several
		// orders, each must also trade near the pool's price, which shows the
		// orders and payouts were paired right.
		if ada.Sign() == 0 || ada.Sign() != tok.Sign() || (n > 1 && !nearPrice(ada, tok, in, asset)) {
			return nil
		}
		w := payouts[i].StakeAddr
		if w == "" {
			w = payouts[i].PaymentAddr.Bech32
		}
		fills = append(fills, fill{wallet: w, ada: ada, token: tok})
	}
	return fills
}

// onePool is the index of the only Minswap pool among outputs, or -1 when
// there is none or more than one.
func onePool(list []koios.TxOutput) int {
	at := -1
	for i, o := range list {
		if o.PaymentAddr.Cred == minswap.PoolCredential {
			if at >= 0 {
				return -1
			}
			at = i
		}
	}
	return at
}

// nearPrice reports whether ADA per token is within 20% of the pool's price
// before the batch.
func nearPrice(ada, tok *big.Int, pool koios.TxOutput, asset koios.Asset) bool {
	reserve := units(pool, asset)
	if reserve.Sign() == 0 {
		return false
	}
	// ada/tok against poolADA/reserve, compared without dividing.
	got := new(big.Int).Mul(new(big.Int).Abs(ada), reserve)
	want := new(big.Int).Mul(lovelace(pool), new(big.Int).Abs(tok))
	low := new(big.Int).Mul(want, big.NewInt(8))
	high := new(big.Int).Mul(want, big.NewInt(12))
	got10 := new(big.Int).Mul(got, big.NewInt(10))
	return got10.Cmp(low) >= 0 && got10.Cmp(high) <= 0
}

// lovelace is the ADA in an output, in lovelace.
func lovelace(o koios.TxOutput) *big.Int {
	n, _ := new(big.Int).SetString(o.Value, 10)
	if n == nil {
		return new(big.Int)
	}
	return n
}

// units is how much of one token an output holds, in the token's smallest unit.
func units(o koios.TxOutput, asset koios.Asset) *big.Int {
	total := new(big.Int)
	for _, a := range o.AssetList {
		if a.PolicyID == asset.PolicyID && a.AssetName == asset.AssetName {
			if n, ok := new(big.Int).SetString(a.Quantity, 10); ok {
				total.Add(total, n)
			}
		}
	}
	return total
}

// sub is a − b as a new number.
func sub(a, b *big.Int) *big.Int { return new(big.Int).Sub(a, b) }
