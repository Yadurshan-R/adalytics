package market

import (
	"context"
	"sort"

	"cardano-analytics/internal/koios"
)

// update checks the chain tip and, with utxo_info (50 pools per request), which
// pools were spent (a swap happened) since the previous check. Only then it
// reads Minswap's transactions since that check and takes the new pool states.
func (t *Tracker) update(ctx context.Context) {
	t.mu.RLock()
	known := t.tracked()
	refs := make([]string, len(known))
	for i, s := range known {
		refs[i] = s.poolRef
	}
	t.mu.RUnlock()
	if len(known) == 0 {
		return
	}

	tip, err := t.client.Tip(ctx)
	if err != nil {
		t.failAll(err)
		return
	}
	var utxos []koios.UTxO
	for start := 0; start < len(refs); start += 50 {
		batch, err := t.client.UTxOInfo(ctx, refs[start:min(start+50, len(refs))])
		if err != nil {
			t.failAll(err)
			return
		}
		utxos = append(utxos, batch...)
	}
	unspent := map[string]bool{}
	for _, u := range utxos {
		if !u.IsSpent {
			unspent[u.Ref()] = true
		}
	}

	// A pool spent now was spent after the previous check, so its swap is in a
	// block from lastCheck on (minus a few blocks of safety margin).
	since := t.lastCheck - 3
	t.lastCheck = tip.BlockHeight
	addresses := map[string]bool{PoolAddress: true}
	fromHeight, spent := 0, 0
	t.mu.Lock()
	for i, s := range known {
		if unspent[refs[i]] {
			continue
		}
		spent++
		if s.address != "" {
			addresses[s.address] = true
		}
		if s.readFrom == 0 {
			s.readFrom = max(since, s.lastHeight+1)
		}
		if fromHeight == 0 || s.readFrom < fromHeight {
			fromHeight = s.readFrom
		}
	}
	t.mu.Unlock()

	found := map[*tokenState][]poolState{}
	if spent > 0 {
		list := make([]string, 0, len(addresses))
		for a := range addresses {
			list = append(list, a)
		}
		sort.Strings(list)
		txs, err := t.client.AddressTxs(ctx, list, fromHeight, nil)
		if err != nil {
			t.failAll(err)
			return
		}
		if found, err = t.readPools(ctx, txs, nil); err != nil {
			t.failAll(err)
			return
		}
	}

	changed := 0
	for _, s := range known {
		t.mu.Lock()
		var fresh []poolState
		for _, st := range found[s] {
			if st.height > s.lastHeight {
				fresh = append(fresh, st)
			}
		}
		if len(fresh) == 0 {
			// No swap, or Koios has not indexed the swap's block yet (then the
			// pool stays "spent" and is read again next minute).
			now := t.now()
			s.quote.UpdatedAt, s.quote.Error = &now, ""
			t.mu.Unlock()
			continue
		}
		changed++
		prev := s.quote.PriceADA
		t.merge(s, fresh)
		s.readFrom = 0
		q := s.quote
		t.mu.Unlock()
		t.log.Printf("  %-8s %-16s %-9s %d swap%s",
			q.Ticker, FormatADA(q.PriceADA), change(prev, q.PriceADA), len(fresh), plural(len(fresh)))
	}
	t.log.Printf("Checked %d pools: %d had new swaps", len(known), changed)
}
