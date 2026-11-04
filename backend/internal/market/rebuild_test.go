// Tests for turning a token's pool history into trades.
package market

import (
	"encoding/json"
	"io"
	"log"
	"math"
	"os"
	"testing"
	"time"

	"cardano-analytics/internal/koios"
	"cardano-analytics/internal/minswap"
)

// minToken is MIN from tokens.json.
func minToken(t *testing.T) Token {
	for _, tok := range Tokens {
		if tok.Ticker == "MIN" && tok.Asset.AssetName == "4d494e" {
			return tok
		}
	}
	t.Fatal("MIN not in tokens.json")
	return Token{}
}

// TestFirstStateTrades uses a real routed trade (4 Nov 2026, 14:47:
// NIGHT → MIN → ADA): ₳677.89 came out of the MIN/ADA pool. When it is the
// only pool change loaded, as just after a start when only the newest hour is
// read, it still counts as a trade; and once an older state is loaded too, it
// is still one trade, not two.
func TestFirstStateTrades(t *testing.T) {
	data, err := os.ReadFile("testdata/min_routed.json")
	if err != nil {
		t.Fatal(err)
	}
	var txs []koios.Tx
	if err := json.Unmarshal(data, &txs); err != nil {
		t.Fatal(err)
	}
	tx := txs[0]
	tok := minToken(t)
	tr := NewTracker(koios.New("http://unused", "test-token"), []Token{tok}, time.Minute, log.New(io.Discard, "", 0))
	tr.now = func() time.Time { return time.Unix(tx.TxTimestamp, 0).Add(10 * time.Minute) }
	s := tr.states[0]

	var state poolState
	for _, o := range tx.Outputs {
		if p, err := minswap.Parse(o.UTxO(tx)); err == nil && isPool(o) && p.LPAsset.AssetName == tok.PoolLP {
			state = poolState{pool: p, height: tx.BlockHeight, order: chainOrder(tx.BlockHeight, tx.TxBlockIndex),
				address: o.PaymentAddr.Bech32, fills: fillsFrom(tx, tok.Asset, p.LPAsset)}
		}
	}
	if state.fills == nil {
		t.Fatal("the routed trade gave no fills")
	}

	check := func(when string) {
		t.Helper()
		if len(s.trades) != 1 {
			t.Fatalf("%s: %d trades, want 1", when, len(s.trades))
		}
		tr0 := s.trades[0]
		if tr0.Side != "sell" || math.Abs(tr0.AmountADA-677.893639) > 1e-6 || tr0.Wallet == "" {
			t.Errorf("%s: trade %+v, want a sell of ₳677.893639 with a wallet", when, tr0)
		}
		hour := periodsOf(s, tr.now())[1]
		if hour.Key != "1h" || hour.Sells != 1 || math.Abs(hour.VolumeADA-677.893639) > 1e-6 {
			t.Errorf("%s: 1h period %+v, want 1 sell and ₳677.89 volume", when, hour)
		}
	}

	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.merge(s, []poolState{state})
	check("only this pool change loaded")

	// The pool as it was before the trade (the transaction's input), 5 minutes
	// earlier: the older history arriving later.
	for _, o := range tx.Inputs {
		if p, err := minswap.Parse(o.UTxO(tx)); err == nil && isPool(o) && p.LPAsset.AssetName == tok.PoolLP {
			p.LastSwap = p.LastSwap.Add(-5 * time.Minute)
			tr.merge(s, []poolState{{pool: p, height: tx.BlockHeight - 15, order: chainOrder(tx.BlockHeight-15, 0), address: o.PaymentAddr.Bech32}})
		}
	}
	if len(s.history) != 2 {
		t.Fatalf("history has %d states, want 2", len(s.history))
	}
	check("with the state before it loaded too")
}
