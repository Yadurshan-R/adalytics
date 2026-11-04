package market

import (
	"context"
	"errors"

	"cardano-analytics/internal/koios"
)

// addBatchOrders fetches the batcher's own list of which output pays which
// order (see minswap.BatchOrder) for the transactions that need it: those
// where one wallet placed several orders, so the orders' datums alone do not
// say which payout is whose. Those are a few in each request of 50, so this
// costs one small extra request now and then, instead of downloading every
// transaction's scripts. If it fails, those transactions are shown as before:
// one trade without a wallet.
func (t *Tracker) addBatchOrders(ctx context.Context, txs []koios.Tx) {
	at := map[string]int{}
	var need []string
	for i, tx := range txs {
		if needsBatchOrder(tx) {
			at[tx.TxHash] = i
			need = append(need, tx.TxHash)
		}
	}
	for start := 0; start < len(need); start += 50 {
		got, err := t.client.TxContracts(ctx, need[start:min(start+50, len(need))])
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				t.log.Printf("Could not read %d batch orders (their trades show without a wallet): %v", len(need)-start, err)
			}
			return
		}
		for _, g := range got {
			if i, ok := at[g.TxHash]; ok {
				txs[i].Contracts = g.Contracts
			}
		}
	}
}
