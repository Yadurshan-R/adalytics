package market

import (
	"sort"

	"cardano-analytics/internal/koios"
	"cardano-analytics/internal/minswap"
)

// pairByReceiver puts each order next to its own payout, using the order's
// datum: its success receiver is the wallet the payout goes to. Koios lists a
// transaction's inputs and outputs in no fixed order, so their positions alone
// cannot pair them.
//
// It returns the orders and their payouts in the same order, sorted by the
// payout's place in the transaction (the order the batcher swapped them in).
// Several orders from one wallet are paired in order; the wallet is right
// either way. When no order has a Minswap V2 datum (older orders), the lists
// come back as they were. It returns nil when some orders have a datum and
// others do not, an order is for another pool, or an order has no payout to
// its receiver: then the transaction is not split.
func pairByReceiver(orders, payouts []koios.TxOutput, lp koios.Asset) ([]koios.TxOutput, []koios.TxOutput) {
	read := make([]minswap.Order, len(orders))
	known := 0
	for i, o := range orders {
		if d, ok := minswap.ReadOrder(o.InlineDatum); ok {
			read[i] = d
			known++
		}
	}
	if known == 0 {
		return orders, payouts
	}
	if known != len(orders) {
		return nil, nil
	}

	// Orders in chain order (by UTxO), payouts in transaction order, so a
	// wallet's several orders pair the same way every time.
	idx := make([]int, len(orders))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(a, b int) bool {
		x, y := orders[idx[a]], orders[idx[b]]
		return x.TxHash < y.TxHash || (x.TxHash == y.TxHash && x.TxIndex < y.TxIndex)
	})
	free := append([]koios.TxOutput(nil), payouts...)
	sort.SliceStable(free, func(a, b int) bool { return free[a].TxIndex < free[b].TxIndex })
	used := make([]bool, len(free))

	type pair struct{ order, payout koios.TxOutput }
	pairs := make([]pair, 0, len(orders))
	for _, i := range idx {
		if read[i].LPAsset != lp {
			return nil, nil
		}
		found := -1
		for j, p := range free {
			if !used[j] && p.PaymentAddr.Cred == read[i].Receiver {
				found = j
				break
			}
		}
		if found < 0 {
			return nil, nil
		}
		used[found] = true
		pairs = append(pairs, pair{orders[i], free[found]})
	}
	sort.SliceStable(pairs, func(a, b int) bool { return pairs[a].payout.TxIndex < pairs[b].payout.TxIndex })
	outOrders, outPayouts := make([]koios.TxOutput, len(pairs)), make([]koios.TxOutput, len(pairs))
	for i, p := range pairs {
		outOrders[i], outPayouts[i] = p.order, p.payout
	}
	return outOrders, outPayouts
}
