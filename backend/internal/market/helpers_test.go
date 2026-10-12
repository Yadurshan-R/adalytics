// Test helpers: a fake Koios server and real Minswap pool data to feed it.
package market

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"cardano-analytics/internal/koios"
	"cardano-analytics/internal/minswap"
)

var testNow = time.Unix(1791180000, 0)

// loadPools reads real Minswap pools (SNEK/MIN and SNEK/ADA) saved from Koios.
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

// fakeTx is one transaction the fake Koios server knows about.
type fakeTx struct {
	hash   string
	height int
	time   int64
	outs   []koios.TxOutput
}

// fake is a pretend Koios server. Tests set its transactions and pools,
// then check which requests the Tracker sent.
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

// serve answers each Koios endpoint the Tracker uses from the fake's data,
// and records every request.
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
		json.NewEncoder(w).Encode([]koios.Tip{{BlockHeight: f.tip, BlockTime: testNow.Unix()}})
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
				rows = append(rows, koios.AddressTx{TxHash: f.txs[i].hash, BlockHeight: f.txs[i].height, BlockTime: f.txs[i].time})
			}
		}
		// Pages, like Koios: offset and limit.
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if offset > len(rows) {
			offset = len(rows)
		}
		rows = rows[offset:]
		if limit > 0 && limit < len(rows) {
			rows = rows[:limit]
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

// newTracker starts a fake Koios server and a Tracker for SNEK that uses it,
// with the clock fixed at testNow.
func newTracker(t *testing.T, f *fake) *Tracker {
	srv := httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(srv.Close)
	tr := NewTracker(koios.New(srv.URL, "test-token"), []Token{snekToken(t)}, time.Minute, log.New(io.Discard, "", 0))
	tr.now = func() time.Time { return testNow }
	return tr
}
