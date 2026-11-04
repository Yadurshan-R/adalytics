package minswap

import (
	"encoding/hex"

	"cardano-analytics/internal/koios"
)

// BatchingScript is the script hash of Minswap V2's pool batching validator
// (its "Liquidity Pool Batching" stake address; every pool's datum names it in
// field 0). A batch transaction runs it once, as a withdrawal.
const BatchingScript = "1eae96baf29e27682ea3f815aba361a0c6059d45e4bfbe95bbd2f44a"

// BatchOrder reads the order the batcher swapped a transaction's orders in,
// from the batching redeemer (field 2, input_indexes: one byte per order).
//
// As Minswap's pool validator does: the orders are the inputs that are neither
// a pool nor the batcher, sorted by UTxO (tx hash, then index), and the k-th
// order processed is the one at position order[k] of that sorted list. Each
// order processed pays out with the next output, starting at output 0, so
// output k pays the order at position order[k]. ok is false when the
// transaction has no batching redeemer, or it is not shaped as expected.
func BatchOrder(contracts []koios.Contract) (order []int, ok bool) {
	for _, c := range contracts {
		if c.ScriptHash != BatchingScript || c.Input.Redeemer.Purpose != "reward" {
			continue
		}
		f := c.Input.Redeemer.Datum.Value.Fields
		if len(f) < 3 || f[2].Bytes == nil {
			return nil, false
		}
		b, err := hex.DecodeString(*f[2].Bytes)
		if err != nil || len(b) == 0 {
			return nil, false
		}
		order = make([]int, len(b))
		for i, x := range b {
			order[i] = int(x)
		}
		return order, true
	}
	return nil, false
}
