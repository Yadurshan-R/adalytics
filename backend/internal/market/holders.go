package market

import (
	"context"
	"errors"
	"sync"
	"time"

	"cardano-analytics/internal/koios"
)

// HoldersWorkers is how many tokens are counted at the same time. Koios takes
// from a few seconds to over a minute per token, so three at a time fill the
// Holders column about three times as fast. Each count is one request, and
// the rate limiter (90 requests per 10 seconds) still holds.
const HoldersWorkers = 3

// HoldersRetry is how soon tokens whose count failed are tried again (a busy
// Koios often times out on big tokens), instead of waiting HoldersEvery.
const HoldersRetry = 5 * time.Minute

// holdersLoop waits for the first load, counts every listed token's holders,
// then counts them all again every HoldersEvery, until ctx ends. Tokens whose
// count failed are tried again every HoldersRetry in between.
//
// Waiting for the first load matters: if Koios is down at startup the list is
// empty, and counting then would leave every token without holders.
func (t *Tracker) holdersLoop(ctx context.Context) {
	for {
		t.mu.RLock()
		ready := t.loaded
		t.mu.RUnlock()
		if ready {
			break
		}
		if err := sleepCtx(ctx, 10*time.Second); err != nil {
			return
		}
	}
	for {
		next := time.Now().Add(HoldersEvery)
		failed := t.loadHolders(ctx)
		for len(failed) > 0 && time.Until(next) > HoldersRetry {
			if err := sleepCtx(ctx, HoldersRetry); err != nil {
				return
			}
			t.log.Printf("Counting holders again for %d tokens that failed…", len(failed))
			failed = t.countHolders(ctx, failed)
		}
		if err := sleepCtx(ctx, time.Until(next)); err != nil {
			return
		}
	}
}

// loadHolders counts the holders of every token in the current list, busiest
// first, and returns the page ids of the tokens whose count failed.
func (t *Tracker) loadHolders(ctx context.Context) []string {
	top := t.Quotes()
	ids := make([]string, len(top))
	for i, q := range top {
		ids[i] = q.ID
	}
	t.log.Printf("Counting holders for %d tokens in the background, %d at a time (slow, can take a while)…", len(ids), HoldersWorkers)
	return t.countHolders(ctx, ids)
}

// countHolders counts the holders of the tokens with these page ids,
// HoldersWorkers at a time: wallets with a stake key plus addresses without
// one, from one slow asset_summary request per token. It returns the ids whose
// count failed, in the order given, so they can be tried again later.
func (t *Tracker) countHolders(ctx context.Context, ids []string) []string {
	failed := make([]bool, len(ids))
	var mu sync.Mutex // keeps the log lines whole
	koios.Parallel(ctx, len(ids), HoldersWorkers, func(i int) error {
		s, ticker := t.stateByID(ids[i])
		if s == nil {
			return nil // left the list since; nothing to count
		}
		start := time.Now()
		sum, err := t.client.AssetSummary(ctx, s.token.Asset)
		if err != nil {
			failed[i] = true
			if !errors.Is(err, context.Canceled) {
				mu.Lock()
				t.log.Printf("  %-8s could not count holders: %v (trying again in %s)", ticker, err, HoldersRetry)
				mu.Unlock()
			}
			return nil // keep counting the other tokens
		}
		n := sum.StakedWallets + sum.UnstakedAddresses
		t.mu.Lock()
		s.holders = n
		t.mu.Unlock()
		mu.Lock()
		t.log.Printf("  %-8s %s holders (counted in %s)", ticker, group(float64(n)), time.Since(start).Round(time.Second))
		mu.Unlock()
		return nil
	})
	var out []string
	for i, f := range failed {
		if f && ctx.Err() == nil {
			out = append(out, ids[i])
		}
	}
	return out
}

// stateByID finds a token by its page id (its slug) and returns it with its
// ticker, read under the lock; nil if there is no such token.
func (t *Tracker) stateByID(id string) (*tokenState, string) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, s := range t.states {
		if s.slug == id {
			return s, s.quote.Ticker
		}
	}
	return nil, ""
}
