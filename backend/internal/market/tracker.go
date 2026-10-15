// Package market keeps token prices and their recent history in memory.
//
// The Tracker is split over several files:
//
//	tracker.go   the Tracker type, settings and the main loop
//	quotes.go    what the API reads: token list, charts, trades, logos
//	startup.go   the first load (last 24 hours)
//	backfill.go  older history, loaded in the background
//	update.go    the check for new swaps every minute
//	pools.go     turning transactions into pool states, prices and trades
//	details.go   supply, logos, links and holder counts
//	progress.go  loading progress for the loading screen
//	format.go    small number and text helpers
package market

import (
	"context"
	"errors"
	"log"
	"math/big"
	"strings"
	"sync"
	"time"

	"cardano-analytics/internal/koios"
	"cardano-analytics/internal/minswap"
)

// PoolAddress is where Minswap V2 pools live. One address_txs request for it
// returns the recent transactions of every Minswap V2 pool at once.
const PoolAddress = "addr1z84q0denmyep98ph3tmzwsmw0j7zau9ljmsqx6a4rvaau66j2c79gy9l76sdg0xwhd7r0c0kna0tycz4y5s6mlenh8pq777e2a"

// Cardano makes a block about every 20 seconds.
const secondsPerBlock = 20

// TopN is how many tokens the list shows, ranked by 24h trading volume.
const TopN = 93

// MinPoolADA leaves out pools with less ADA than this: their prices move on
// tiny trades and are not reliable.
const MinPoolADA = 10_000

// MinTradeADA hides trades smaller than this: tiny pool changes left over from
// rounding inside Minswap's batches, not trades anyone would look for.
const MinTradeADA = 1

// USDM is a US dollar stablecoin with a Minswap ADA pool. Its price in ADA gives
// the ADA/USD rate used for USD prices, so they come from Koios like the rest.
var USDM = koios.Asset{PolicyID: "c48cbb3d5e57ed56e276bc45f99ab39abe94e6cd7ac39fb402da47ad", AssetName: "0014df105553444d"}

// HoldersEvery is how often holder counts are refreshed. Koios takes up to a
// minute or more per token to count them, so this runs in the background.
const HoldersEvery = 6 * time.Hour

// tokenState is everything the Tracker keeps in memory for one token.
type tokenState struct {
	token      Token
	slug       string // unique id for page addresses
	lp         koios.Asset
	poolRef    string          // current pool UTxO "tx_hash#index"; empty until found
	address    string          // address the pool lives at
	lastHeight int             // block height of the newest pool state we read
	history    []poolState     // every pool state in the kept history, oldest first
	refs       map[string]bool // pool UTxOs already in history, so none is added twice
	points     []Point         // one per pool state, oldest first (built from history)
	trades     []Trade         // swaps with amounts, oldest first (built from history)
	readFrom   int             // when the pool is spent but its swap not read yet: block to read from
	supply     *big.Int        // total supply in the smallest unit, from asset_info
	logo       []byte          // PNG from the token registry
	holders    int             // from asset_summary, loaded in the background
	mint       *mintPolicy     // the minting policy from script_info; nil until read
	info       Quote           // fingerprint, created, description and website from asset_info
	quote      Quote
}

// poolState is a pool right after one transaction.
type poolState struct {
	pool    minswap.Pool
	height  int
	order   int64  // sorts states in chain order: block height, then position in the block
	address string // address the pool output was at
	fills   []fill // each wallet's part of the swap; nil if it could not be split
	current bool   // found as the pool is now (asset_utxos), not read from its transaction
}

// chainOrder gives one number for sorting transactions in the order they
// happened: block height first, then position inside the block.
func chainOrder(height, blockIndex int) int64 { return int64(height)*100_000 + int64(blockIndex) }

// Tracker finds every token's Minswap pool and loads the last 24 hours of swaps
// at startup, then follows new swaps on a fixed interval. Everything is kept in
// memory only.
type Tracker struct {
	client   *koios.Client
	interval time.Duration
	log      *log.Logger
	now      func() time.Time

	span time.Duration // how much history is loaded and kept (HISTORY_DAYS)

	mu          sync.RWMutex
	states      []*tokenState
	loaded      bool
	dayDone     bool         // the whole first 24 hours are read (see day.go)
	async       func(func()) // runs the rest of the first day; tests run it in place
	infoReady   bool         // supply and logos loaded
	lastCheck   int          // chain block height at the previous check
	firstHeight int          // block the first (24 hour) load started from
	coveredFrom time.Time    // history is complete from this time until now
	historyDone bool         // the whole span has been loaded
	boot        bootProgress // where the first load is, for the loading screen
	bootStart   time.Time    // when the first load started
	activity    []Activity   // recent messages for the loading page (see activity.go)
}

// NewTracker prepares a Tracker for the given tokens. Nothing is loaded until Run is called.
func NewTracker(client *koios.Client, tokens []Token, interval time.Duration, logger *log.Logger) *Tracker {
	t := &Tracker{client: client, interval: interval, log: logger, now: time.Now, span: MaxHistoryDays * 24 * time.Hour,
		async: func(f func()) { go f() }}
	for _, tok := range tokens {
		t.states = append(t.states, &tokenState{
			token: tok,
			lp:    koios.Asset{PolicyID: minswap.LPPolicy, AssetName: tok.PoolLP},
			quote: Quote{
				Ticker: tok.Ticker, Name: tok.Name,
				PolicyID: tok.Asset.PolicyID, AssetName: tok.Asset.AssetName,
				Decimals: tok.Decimals,
				Status:   "loading",
			},
		})
	}
	assignSlugs(t.states)
	return t
}

// SetHistoryDays sets how many days of history are loaded and kept (1 to 30).
// Call it before Run.
func (t *Tracker) SetHistoryDays(days int) {
	days = min(max(days, 1), MaxHistoryDays)
	t.span = time.Duration(days) * 24 * time.Hour
}

// assignSlugs gives each token a short unique id, normally its ticker in lower
// case; tokens sharing a ticker get part of their policy ID added.
func assignSlugs(states []*tokenState) {
	base := func(s *tokenState) string {
		src := s.token.Ticker
		if src == "" {
			src = s.token.Name
		}
		var b strings.Builder
		for _, c := range strings.ToLower(src) {
			if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
				b.WriteRune(c)
			}
		}
		if b.Len() == 0 {
			return s.token.Asset.PolicyID[:8]
		}
		return b.String()
	}
	count := map[string]int{}
	for _, s := range states {
		count[base(s)]++
	}
	for _, s := range states {
		s.slug = base(s)
		if count[s.slug] > 1 {
			s.slug += "-" + s.token.Asset.PolicyID[:6]
		}
		s.quote.ID = s.slug
	}
}

// Run loads history now, then checks for new swaps every interval, until ctx ends.
// Holder counts are loaded separately in the background.
func (t *Tracker) Run(ctx context.Context) {
	t.refresh(ctx)
	go t.holdersLoop(ctx)
	go t.backfillLoop(ctx)
	tick := time.NewTicker(t.interval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			t.refresh(ctx)
		}
	}
}

// refresh runs once a minute: the first time it does the first load, and after
// that it loads token details if they are still missing and checks for new swaps.
func (t *Tracker) refresh(ctx context.Context) {
	if !t.loaded {
		t.loadHistory(ctx)
		return
	}
	if !t.infoReady {
		t.loadInfo(ctx)
	}
	t.update(ctx)
}

// tracked returns the tokens whose pool was found and holds at least
// MinPoolADA: the ones listed and checked every minute. Call with t.mu held.
func (t *Tracker) tracked() []*tokenState {
	var out []*tokenState
	for _, s := range t.states {
		if s.poolRef != "" && s.quote.ReserveADA >= MinPoolADA {
			out = append(out, s)
		}
	}
	return out
}

// sleepCtx waits for d, or returns early with an error when ctx is cancelled.
func sleepCtx(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// blocksIn is about how many blocks Cardano makes in d.
func blocksIn(d time.Duration) int { return int(d/time.Second) / secondsPerBlock }

// failAll records an error that affects every token and logs it once.
func (t *Tracker) failAll(err error) {
	if errors.Is(err, context.Canceled) {
		return
	}
	t.mu.Lock()
	for _, s := range t.states {
		s.quote.Error = err.Error()
		if s.quote.PriceADA == 0 {
			s.quote.Status = "error"
		}
	}
	t.mu.Unlock()
	t.log.Printf("Could not update from Koios: %v (trying again in %s)", err, t.interval)
}
