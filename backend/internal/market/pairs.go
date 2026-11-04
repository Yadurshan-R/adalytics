package market

import (
	"sort"

	"cardano-analytics/internal/koios"
	"cardano-analytics/internal/minswap"
)

// ordersAndPayouts sorts a transaction's inputs and outputs, like Minswap's
// pool validator does: the batcher's coins go in and come back out at the
// same address, pools are counted, every other input is an order and every
// other output a payout. Orders come sorted by UTxO (tx hash, then index), the
// order the ledger keeps inputs in; payouts by their place in the transaction.
func ordersAndPayouts(tx koios.Tx) (orders, payouts []koios.TxOutput, pools int) {
	inAddr, outAddr := map[string]bool{}, map[string]bool{}
	for _, o := range tx.Inputs {
		inAddr[o.PaymentAddr.Bech32] = true
	}
	for _, o := range tx.Outputs {
		outAddr[o.PaymentAddr.Bech32] = true
	}
	for _, o := range tx.Inputs {
		if isPool(o) {
			pools++
		} else if !outAddr[o.PaymentAddr.Bech32] {
			orders = append(orders, o)
		}
	}
	for _, o := range tx.Outputs {
		if !isPool(o) && !inAddr[o.PaymentAddr.Bech32] {
			payouts = append(payouts, o)
		}
	}
	sort.SliceStable(orders, func(a, b int) bool {
		x, y := orders[a], orders[b]
		return x.TxHash < y.TxHash || (x.TxHash == y.TxHash && x.TxIndex < y.TxIndex)
	})
	sort.SliceStable(payouts, func(a, b int) bool { return payouts[a].TxIndex < payouts[b].TxIndex })
	return orders, payouts, pools
}

// pairOrders puts each order next to its own payout and returns both lists in
// the order the batcher swapped them. Koios lists a transaction's outputs in no
// fixed order, so positions alone cannot pair them.
//
//   - Each order's datum names the wallet its payout goes to (or its refund
//     wallet, when the batcher killed the order). When every order names a
//     different wallet, that pairs them.
//   - When one wallet has several orders, the batcher's own list says which
//     output pays which order (minswap.BatchOrder; tx.Contracts holds it, see
//     addBatchOrders). Each payout must still go to its order's wallet.
//
// sure is true when the pairing came from the datums. With no datum at all
// (older orders) the lists come back as they were, and sure is false. It
// returns nil lists when the pairing is not certain: only some orders have a
// datum, an order is for another pool, a payout goes to another wallet, or one
// wallet has several orders and the batcher's list is missing.
func pairOrders(tx koios.Tx, orders, payouts []koios.TxOutput, lp koios.Asset) (o, p []koios.TxOutput, sure bool) {
	read := make([]minswap.Order, len(orders))
	known, wallets := 0, map[string]int{}
	for i, x := range orders {
		if d, ok := minswap.ReadOrder(x.InlineDatum); ok {
			read[i] = d
			known++
			wallets[d.Receiver]++
		}
	}
	if known == 0 {
		return orders, payouts, false
	}
	if known != len(orders) {
		return nil, nil, false
	}
	for _, d := range read {
		if d.LPAsset != lp {
			return nil, nil, false
		}
	}
	// paidTo reports whether a payout went to order i's wallet (refund: or to
	// its refund wallet).
	paidTo := func(i int, out koios.TxOutput, refund bool) bool {
		c := out.PaymentAddr.Cred
		return c == read[i].Receiver || (refund && c == read[i].Refund)
	}

	o, p = make([]koios.TxOutput, 0, len(orders)), make([]koios.TxOutput, 0, len(orders))
	if len(wallets) < len(orders) { // one wallet has several orders
		order, ok := minswap.BatchOrder(tx.Contracts)
		if !ok || len(order) != len(orders) {
			return nil, nil, false
		}
		// Like the pool validator: output k pays the k-th order processed.
		outs := append([]koios.TxOutput(nil), tx.Outputs...)
		sort.SliceStable(outs, func(a, b int) bool { return outs[a].TxIndex < outs[b].TxIndex })
		taken := map[int]bool{}
		for k, i := range order {
			if i < 0 || i >= len(orders) || taken[i] || k >= len(outs) || isPool(outs[k]) || !paidTo(i, outs[k], true) {
				return nil, nil, false
			}
			taken[i] = true
			o, p = append(o, orders[i]), append(p, outs[k])
		}
		return o, p, true
	}
	used := make([]bool, len(payouts))
	for i := range orders {
		found := -1
		for _, refund := range []bool{false, true} { // its own wallet first, then its refund wallet
			for j, out := range payouts {
				if found < 0 && !used[j] && paidTo(i, out, refund) {
					found = j
				}
			}
		}
		if found < 0 {
			return nil, nil, false
		}
		used[found] = true
		o, p = append(o, orders[i]), append(p, payouts[found])
	}
	o, p = sortPairs(o, p)
	return o, p, true
}

// sortPairs returns the pairs ordered by the payout's place in the
// transaction: the order the batcher swapped them in.
func sortPairs(o, p []koios.TxOutput) ([]koios.TxOutput, []koios.TxOutput) {
	idx := make([]int, len(p))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(a, b int) bool { return p[idx[a]].TxIndex < p[idx[b]].TxIndex })
	so, sp := make([]koios.TxOutput, len(p)), make([]koios.TxOutput, len(p))
	for k, i := range idx {
		so[k], sp[k] = o[i], p[i]
	}
	return so, sp
}

// needsBatchOrder reports whether pairing a transaction's orders needs the
// batcher's own list: one wallet placed several orders in it, so the orders'
// datums alone do not say which payout is whose.
func needsBatchOrder(tx koios.Tx) bool {
	orders, _, pools := ordersAndPayouts(tx)
	if pools != 1 || len(orders) < 2 {
		return false
	}
	seen := map[string]bool{}
	for _, o := range orders {
		d, ok := minswap.ReadOrder(o.InlineDatum)
		if !ok {
			return false
		}
		if seen[d.Receiver] {
			return true
		}
		seen[d.Receiver] = true
	}
	return false
}
