package market

import (
	"context"
	"time"

	"cardano-analytics/internal/koios"
	"cardano-analytics/internal/minswap"
)

// loadHistory is the first load, kept short so the site opens within about
// half a minute:
//  1. list every Minswap V2 transaction of the last 24 hours (address_txs,
//     newest first);
//  2. read the newest hour of them (tx_info): prices, 5m and 1h numbers and
//     the pool of every token that traded;
//  3. the current pool of every other token (asset_utxos, 30 per request);
//  4. supply, logos and links. The list opens here.
//
// The rest of the 24 hours is then read in the background, newest first
// (readDay in day.go), and after it the older history (backfill).
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

	// The newest hour first: the list can open with it.
	hourAgo := time.Unix(tip.BlockTime, 0).Add(-time.Hour)
	first := 0
	for first < len(txs) && !time.Unix(txs[first].BlockTime, 0).Before(hourAgo) {
		first++
	}
	t.setBoot(func(b *bootProgress) { b.step, b.found = 1, first })
	t.Note("Reading the newest hour first: %s of %s transactions", group(float64(first)), group(float64(len(txs))))
	found, err := t.readPools(ctx, txs[:first], func(done, total int) {
		t.setBoot(func(b *bootProgress) { b.read = done })
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
	t.Note("%d tokens traded in the last hour. Finding the pools of the other %d…", traded, len(quiet))
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
	t.coveredFrom = hourAgo
	t.mu.Unlock()

	t.log.Printf("Ready: %d tokens traded in the last hour, %d more found by their pool, %d pools with at least ₳%s are listed (%d Koios requests, %s). Reading the rest of the 24 hours…",
		traded, len(current), listed, group(MinPoolADA), t.client.Requests()-before, time.Since(start).Round(time.Second))
	rest := txs[first:]
	end := time.Unix(tip.BlockTime, 0).Add(-25 * time.Hour)
	t.async(func() { t.readDay(ctx, rest, end) })
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
		out[s] = poolState{pool: p, height: u.BlockHeight, order: chainOrder(u.BlockHeight, 0), address: u.Address, current: true}
	}
	return out, nil
}
