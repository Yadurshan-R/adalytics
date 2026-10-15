package market

import (
	"context"
	"time"

	"cardano-analytics/internal/koios"
)

// DayChunk is how many transactions readDay reads before it updates the
// prices, volume and trades on the site: six tx_info requests, three at a time.
const DayChunk = 300

// readDay reads the rest of the first 24 hours after the list has opened,
// newest first, in chunks of DayChunk transactions. After each chunk the
// site has more of the day: the 4h numbers are complete after a few chunks,
// the 24h numbers at the end. A chunk that fails is tried again a minute
// later. end is when the day's history is complete from (25 hours ago);
// when it is done, the older history (backfill) starts.
func (t *Tracker) readDay(ctx context.Context, txs []koios.AddressTx, end time.Time) {
	start, before := time.Now(), t.client.Requests()
	for i := 0; i < len(txs); {
		j := min(i+DayChunk, len(txs))
		found, err := t.readPools(ctx, txs[i:j], nil)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			t.log.Printf("First day: %v (trying again in 1m)", err)
			if sleepCtx(ctx, time.Minute) != nil {
				return
			}
			continue
		}
		t.mu.Lock()
		for s, states := range found {
			t.merge(s, states)
		}
		// Everything from the oldest transaction of this chunk on is read; one
		// second later, as the next chunk may hold more of that same block.
		if oldest := time.Unix(txs[j-1].BlockTime, 0).Add(time.Second); oldest.Before(t.coveredFrom) {
			t.coveredFrom = oldest
		}
		hours := t.now().Sub(t.coveredFrom).Hours()
		t.mu.Unlock()
		t.Note("Read %s of %s transactions: the last %.0f hours are in", group(float64(j)), group(float64(len(txs))), hours)
		i = j
	}
	t.mu.Lock()
	if end.Before(t.coveredFrom) {
		t.coveredFrom = end
	}
	t.dayDone = true
	t.mu.Unlock()
	t.log.Printf("First day: all 24 hours loaded (%s more transactions, %d Koios requests, %s)",
		group(float64(len(txs))), t.client.Requests()-before, time.Since(start).Round(time.Second))
}
