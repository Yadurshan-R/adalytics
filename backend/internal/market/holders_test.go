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
// fails the first request for one chosen token.
type holdersFake struct {
	mu      sync.Mutex
	running int
	most    int
	failFor string // policy ID whose first count fails
	failed  bool
}

// serve answers one asset_summary request.
func (f *holdersFake) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	f.running++
	f.most = max(f.most, f.running)
	fail := r.URL.Query().Get("_asset_policy") == f.failFor && !f.failed
	if fail {
		f.failed = true
	}
	f.mu.Unlock()
	time.Sleep(40 * time.Millisecond)
	f.mu.Lock()
	f.running--
	f.mu.Unlock()
	if fail {
		http.Error(w, "query failed", http.StatusBadRequest) // a 400 is not retried by the client
		return
	}
	w.Write([]byte(`[{"staked_wallets":100,"unstaked_addresses":5}]`))
}

// TestCountHolders checks that holders are counted HoldersWorkers at a time,
// that a failed token is handed back to be tried again, and that the second
// try fills it in.
func TestCountHolders(t *testing.T) {
	f := &holdersFake{failFor: Tokens[2].Asset.PolicyID}
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
	want := tr.states[2].slug
	if len(failed) != 1 || failed[0] != want {
		t.Fatalf("failed %v, want [%s]", failed, want)
	}
	for i, s := range tr.states {
		if got, wantN := s.holders, map[bool]int{true: 0, false: 105}[i == 2]; got != wantN {
			t.Errorf("%s: %d holders, want %d", s.slug, got, wantN)
		}
	}

	if again := tr.countHolders(ctx, failed); len(again) != 0 {
		t.Errorf("still failed after retry: %v", again)
	}
	if h := tr.states[2].holders; h != 105 {
		t.Errorf("after retry %d holders, want 105", h)
	}
}
