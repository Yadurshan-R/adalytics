package market

import (
	"context"
	"time"

	"cardano-analytics/internal/koios"
)

// backfillLoop waits for the first 24 hours, then loads the rest of the history.
func (t *Tracker) backfillLoop(ctx context.Context) {
	for {
		t.mu.RLock()
		ready, done := t.dayDone, t.historyDone
		t.mu.RUnlock()
		if done {
			return
		}
		if ready {
			t.backfill(ctx)
			return
		}
		if err := sleepCtx(ctx, 10*time.Second); err != nil {
			return
		}
	}
}

// backfill loads the history older than the first 24 hours, newest first, one
// page of 1,000 transactions at a time, so 7 days are ready long before 30.
// A page that fails is tried again a minute later.
func (t *Tracker) backfill(ctx context.Context) {
	start, before := time.Now(), t.client.Requests()
	tip, err := t.client.Tip(ctx)
	for err != nil {
		t.log.Printf("History: %v (trying again in 1m)", err)
		if sleepCtx(ctx, time.Minute) != nil {
			return
		}
		tip, err = t.client.Tip(ctx)
	}
	days := int(t.span / (24 * time.Hour))
	if t.span <= 24*time.Hour {
		t.finishHistory(time.Unix(tip.BlockTime, 0).Add(-t.span - time.Hour))
		return
	}
	from := tip.BlockHeight - blocksIn(t.span+time.Hour)
	target := time.Unix(tip.BlockTime, 0).Add(-t.span - time.Hour)
	t.mu.RLock()
	skipFrom := t.firstHeight // transactions from here on were read by the first load
	t.mu.RUnlock()

	t.log.Printf("Loading %d days of history in the background…", days)
	seen := map[string]bool{}
	lastDays := 1
	for offset := 0; ; offset += koios.AddressTxsPageSize {
		rows, err := t.client.AddressTxsPage(ctx, []string{PoolAddress}, from, offset)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			t.log.Printf("History: %v (trying again in 1m)", err)
			if sleepCtx(ctx, time.Minute) != nil {
				return
			}
			offset -= koios.AddressTxsPageSize
			continue
		}
		var todo []koios.AddressTx
		for _, r := range rows {
			if r.BlockHeight >= skipFrom || seen[r.TxHash] {
				continue
			}
			seen[r.TxHash] = true
			todo = append(todo, r)
		}
		found, err := t.readPools(ctx, todo, nil)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			t.log.Printf("History: %v (trying again in 1m)", err)
			if sleepCtx(ctx, time.Minute) != nil {
				return
			}
			offset -= koios.AddressTxsPageSize
			for _, r := range todo {
				delete(seen, r.TxHash)
			}
			continue
		}

		t.mu.Lock()
		for s, states := range found {
			t.merge(s, states)
		}
		if len(rows) > 0 {
			if oldest := time.Unix(rows[len(rows)-1].BlockTime, 0); oldest.Before(t.coveredFrom) {
				t.coveredFrom = oldest
			}
		}
		loaded := int(t.now().Sub(t.coveredFrom) / (24 * time.Hour))
		t.mu.Unlock()

		if len(rows) < koios.AddressTxsPageSize {
			break
		}
		if loaded > lastDays && loaded < days {
			lastDays = loaded
			t.log.Printf("History: %d of %d days loaded (%d Koios requests so far)", loaded, days, t.client.Requests()-before)
		}
	}
	t.finishHistory(target)
	t.log.Printf("History: all %d days loaded (%d Koios requests, %s)", days, t.client.Requests()-before, time.Since(start).Round(time.Second))
}

// finishHistory marks the whole history span as loaded, so every chart range is ready.
func (t *Tracker) finishHistory(target time.Time) {
	t.mu.Lock()
	if target.Before(t.coveredFrom) {
		t.coveredFrom = target
	}
	t.historyDone = true
	t.mu.Unlock()
}
