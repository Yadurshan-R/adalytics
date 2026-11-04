// Tests for loading history that spans several pages of the transaction list.
package market

import (
	"context"
	"fmt"
	"testing"
	"time"

	"cardano-analytics/internal/koios"
)

// TestBackfillSeveralPages loads 2,500 older transactions: three pages of the
// list (1,000 each). Every page is read and every transaction ends up in the
// history, and the next page's list is asked for while the current page's
// transactions are still being read, so no time is lost between pages.
func TestBackfillSeveralPages(t *testing.T) {
	const n = 2500
	f := &fake{t: t, tip: 14030000}
	day := 24 * 60 * 60 / secondsPerBlock
	for i := 0; i < n; i++ { // spread over days 2 to 6 back, newest first
		h := 14030000 - 2*day - i*(4*day)/n
		hash := fmt.Sprintf("old%04d", i)
		f.txs = append([]fakeTx{{hash: hash, height: h, time: testNow.Add(-time.Duration(14030000-h) * secondsPerBlock * time.Second).Unix(),
			outs: []koios.TxOutput{poolOutput(t, hash, fmt.Sprint(1900000000000+i))}}}, f.txs...)
	}
	tr := newTracker(t, f)
	tr.SetHistoryDays(7)
	ctx := context.Background()
	tr.refresh(ctx) // the first 24 hours: none of these transactions

	before := len(f.calls)
	tr.backfill(ctx)
	calls := f.calls[before:]

	var lists []int // where each page of the list was asked for
	read := 0
	for i, c := range calls {
		switch c {
		case "/address_txs":
			lists = append(lists, i)
		case "/tx_info":
			read++
		}
	}
	if len(lists) != 3 {
		t.Fatalf("asked for %d pages of the list, want 3", len(lists))
	}
	if want := n / 50; read != want {
		t.Errorf("%d tx_info requests, want %d (50 transactions each)", read, want)
	}
	// Page 1 needs 20 tx_info requests; page 2's list must be asked for before
	// all of them are done.
	firstPageReads := 0
	for _, c := range calls[lists[0]:lists[1]] {
		if c == "/tx_info" {
			firstPageReads++
		}
	}
	if firstPageReads >= 20 {
		t.Errorf("page 2 was asked for only after all %d reads of page 1: pages are not overlapped", firstPageReads)
	}

	tr.mu.RLock()
	got := len(tr.states[0].history)
	tr.mu.RUnlock()
	if got < n {
		t.Errorf("history has %d pool states, want at least %d", got, n)
	}
	if h := tr.History(); !h.Done {
		t.Errorf("history not done: %+v", h)
	}
}
