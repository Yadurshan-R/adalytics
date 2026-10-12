package market

import (
	"context"
	"time"

	"cardano-analytics/internal/koios"
	"cardano-analytics/internal/minswap"
)

// loadHistory is the first load, kept short so the site is usable within a
// few minutes:
//  1. every Minswap V2 transaction of the last 24 hours (address_txs, then
//     tx_info 50 at a time): prices, volume and trades, and the pool of every
//     token that traded;
//  2. the current pool of each token that did not trade in 24 hours
//     (asset_utxos, 30 tokens per request).
//
// Older history is loaded afterwards in the background (backfill).
func (t *Tracker) loadHistory(ctx context.Context) {
	start, before := time.Now(), t.client.Requests()
	t.mu.Lock()
	if t.bootStart.IsZero() {
		t.bootStart = t.now() // kept if the first try fails and is run again
	}
	t.mu.Unlock()

	t.Note("Asking Koios for the newest Cardano block…")
	tip, err := t.client.Tip(ctx)
	if err != nil {
		t.Note("Koios did not answer: %v. Trying again in %s.", err, t.interval)
		t.failAll(err)
		return
	}
	from := tip.BlockHeight - blocksIn(25*time.Hour)
	t.setBoot(func(b *bootProgress) { *b = bootProgress{} })

	t.Note("Newest block is %s. Listing every Minswap transaction since block %s (the last 24 hours)…",
		group(float64(tip.BlockHeight)), group(float64(from)))
	txs, err := t.client.AddressTxs(ctx, []string{PoolAddress}, from, func(rows int) {
		t.setBoot(func(b *bootProgress) { b.found = rows })
		t.Note("Listed %s Minswap transactions so far", group(float64(rows)))
	})
	if err != nil {
		t.failAll(err)
		return
	}
	t.setBoot(func(b *bootProgress) { b.step, b.found = 1, len(txs) })
	found, err := t.readPools(ctx, txs, func(done, total int) {
		t.setBoot(func(b *bootProgress) { b.read = done })
		if done == total || done%500 == 0 {
			t.Note("Read %s of %s transactions (prices, volume and trades)", group(float64(done)), group(float64(total)))
		}
	})
	if err != nil {
		t.failAll(err)
		return
	}

	t.mu.Lock()
	traded := 0
	var quiet []*tokenState
	for _, s := range t.states {
		s.history, s.refs, s.points, s.trades, s.poolRef, s.lastHeight = nil, nil, nil, nil, "", 0
		if t.merge(s, found[s]) > 0 {
			traded++
		} else {
			quiet = append(quiet, s)
		}
	}
	t.mu.Unlock()

	t.setBoot(func(b *bootProgress) { b.step, b.quietTotal = 2, len(quiet) })
	t.Note("%d tokens traded in the last 24 hours. Finding the pools of the other %d…", traded, len(quiet))
	current, err := t.findPools(ctx, quiet)
	if err != nil {
		t.failAll(err)
		return
	}

	t.mu.Lock()
	listed := 0
	for s, cur := range current {
		t.merge(s, []poolState{cur})
	}
	for _, s := range t.states {
		if s.poolRef != "" && s.quote.ReserveADA >= MinPoolADA {
			listed++
		}
	}
	t.boot.step = 3
	t.mu.Unlock()

	// Supply, logos and details before the list is shown. If this fails the
	// list is still shown, and it is tried again every minute.
	t.Note("%d tokens have a pool with at least ₳%s. Loading their supply, logos and links…", listed, group(MinPoolADA))
	t.loadInfo(ctx)
	t.Note("Done. Opening the token list.")

	t.mu.Lock()
	t.loaded = true
	t.lastCheck = tip.BlockHeight
	t.firstHeight = from
	t.coveredFrom = time.Unix(tip.BlockTime, 0).Add(-25 * time.Hour)
	t.mu.Unlock()

	t.log.Printf("Ready: %d tokens traded on Minswap in 24h, %d more found quiet, %d pools with at least ₳%s are listed (%s transactions, %d Koios requests, %s)",
		traded, len(current), listed, group(MinPoolADA), group(float64(len(txs))),
		t.client.Requests()-before, time.Since(start).Round(time.Second))
}

// findPools asks Koios which UTxOs at the Minswap pool address hold each
// token's LP token: that UTxO is the token's pool, as it is now.
func (t *Tracker) findPools(ctx context.Context, tokens []*tokenState) (map[*tokenState]poolState, error) {
	out := map[*tokenState]poolState{}
	if len(tokens) == 0 {
		return out, nil
	}
	byLP := map[string]*tokenState{}
	lps := make([]koios.Asset, 0, len(tokens))
	for _, s := range tokens {
		byLP[s.lp.AssetName] = s
		lps = append(lps, s.lp)
	}
	utxos, err := t.client.PoolUTxOs(ctx, PoolAddress, lps, func(done, total int) {
		t.setBoot(func(b *bootProgress) { b.quietDone = done })
		if done == total || done%120 == 0 {
			t.Note("Checked %d of %d quiet tokens", done, total)
		}
	})
	if err != nil {
		return nil, err
	}
	for _, u := range utxos {
		p, err := minswap.Parse(u)
		if err != nil {
			continue
		}
		s, ok := byLP[p.LPAsset.AssetName]
		if !ok || !p.IsADAPair(s.token.Asset) {
			continue
		}
		if prev, ok := out[s]; ok && prev.height >= u.BlockHeight {
			continue
		}
		out[s] = poolState{pool: p, height: u.BlockHeight, order: chainOrder(u.BlockHeight, 0), address: u.Address}
	}
	return out, nil
}
