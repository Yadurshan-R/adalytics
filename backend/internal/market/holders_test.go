package market

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"cardano-analytics/internal/koios"
)

// holdersFake is a pretend Koios that only answers asset_summary. It is slow
// (so several requests overlap), notes the most requests running at once, and
// times out (504, like Koios on a heavy token) the first few requests for the
// chosen tokens.
type holdersFake struct {
	mu      sync.Mutex
	running int
	most    int
	calls   int
	fails   map[string]int // policy ID -> how many of its next counts fail
}

// serve answers one asset_summary request.
func (f *holdersFake) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	f.running++
	f.calls++
	f.most = max(f.most, f.running)
	policy := r.URL.Query().Get("_asset_policy")
	fail := f.fails[policy] > 0
	if fail {
		f.fails[policy]--
	}
	f.mu.Unlock()
	time.Sleep(40 * time.Millisecond)
	f.mu.Lock()
	f.running--
	f.mu.Unlock()
	if fail {
		http.Error(w, "504 Gateway Time-out", http.StatusGatewayTimeout)
		return
	}
	w.Write([]byte(`[{"staked_wallets":100,"unstaked_addresses":5}]`))
}

// TestCountHolders checks that a full count runs HoldersWorkers at a time,
// that failed tokens are handed back to be tried again, and that the retry
// counts them one at a time (HoldersRetryWorkers) and fills them in.
func TestCountHolders(t *testing.T) {
	bad := []int{2, 3, 5} // tokens whose first count times out
	f := &holdersFake{fails: map[string]int{}}
	for _, i := range bad {
		f.fails[Tokens[i].Asset.PolicyID] = 1
	}
	srv := httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(srv.Close)
	tr := NewTracker(koios.New(srv.URL, "test-token"), Tokens[:7], time.Minute, log.New(io.Discard, "", 0))
	for i, s := range tr.states { // every token listed, with a pool big enough to show
		s.poolRef = "pool#0"
		s.quote.ReserveADA = MinPoolADA + float64(i)
	}
	ctx := context.Background()

	failed := tr.loadHolders(ctx)
	if f.most != HoldersWorkers {
		t.Errorf("%d counted at once, want %d", f.most, HoldersWorkers)
	}
	isBad, want := map[int]bool{}, map[string]bool{}
	for _, i := range bad {
		isBad[i], want[tr.states[i].slug] = true, true
	}
	got := map[string]bool{}
	for _, id := range failed {
		got[id] = true
	}
	if len(failed) != len(want) || len(got) != len(want) {
		t.Fatalf("failed %v, want the %d tokens %v", failed, len(want), want)
	}
	for id := range want {
		if !got[id] {
			t.Fatalf("failed %v, want the %d tokens %v", failed, len(want), want)
		}
	}
	for i, s := range tr.states {
		if got, wantN := s.holders, map[bool]int{true: 0, false: 105}[isBad[i]]; got != wantN {
			t.Errorf("%s: %d holders, want %d", s.slug, got, wantN)
		}
	}

	f.most = 0
	if again := tr.countHolders(ctx, failed, HoldersRetryWorkers); len(again) != 0 {
		t.Errorf("still failed after retry: %v", again)
	}
	if f.most != HoldersRetryWorkers {
		t.Errorf("retry counted %d at once, want %d", f.most, HoldersRetryWorkers)
	}
	for _, i := range bad {
		if h := tr.states[i].holders; h != 105 {
			t.Errorf("%s after retry: %d holders, want 105", tr.states[i].slug, h)
		}
	}
}

// TestHoldersUnknownAfterFailures checks that a token Koios keeps timing out
// on is counted once per try (no retries at once), shows "…" until it failed
// HoldersGiveUp times in a row, then shows as unknown, and that one count
// that works brings its number back, with the time it was counted.
func TestHoldersUnknownAfterFailures(t *testing.T) {
	f := &holdersFake{fails: map[string]int{Tokens[0].Asset.PolicyID: HoldersGiveUp}}
	srv := httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(srv.Close)
	tr := NewTracker(koios.New(srv.URL, "test-token"), Tokens[:1], time.Minute, log.New(io.Discard, "", 0))
	tr.states[0].poolRef, tr.states[0].quote.ReserveADA = "pool#0", MinPoolADA
	id := tr.states[0].slug
	quote := func() Quote { return tr.Quotes()[0] }
	ctx := context.Background()

	for try := 1; try <= HoldersGiveUp; try++ {
		if failed := tr.countHolders(ctx, []string{id}, HoldersRetryWorkers); len(failed) != 1 {
			t.Fatalf("try %d: failed %v, want [%s]", try, failed, id)
		}
		if f.calls != try {
			t.Fatalf("try %d: %d requests to Koios, want %d (a timeout is not retried at once)", try, f.calls, try)
		}
		if got, want := quote().HoldersUnknown, try >= HoldersGiveUp; got != want {
			t.Errorf("after %d failed counts: unknown = %v, want %v", try, got, want)
		}
	}
	if failed := tr.countHolders(ctx, []string{id}, HoldersRetryWorkers); len(failed) != 0 {
		t.Fatalf("count after Koios recovered failed: %v", failed)
	}
	if q := quote(); q.Holders != 105 || q.HoldersUnknown || q.HoldersAt == nil {
		t.Errorf("after a count that works: %d holders, unknown %v, counted at %v; want 105, false, a time", q.Holders, q.HoldersUnknown, q.HoldersAt)
	}
}
