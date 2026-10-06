// Package market keeps token prices and their recent history in memory.
package market

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"net/url"
	"sort"
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

// Quote is what the API returns for one token.
type Quote struct {
	Rank         int        `json:"rank"`
	ID           string     `json:"id"` // unique, used in page addresses and API paths
	Ticker       string     `json:"ticker"`
	Name         string     `json:"name"`
	PolicyID     string     `json:"policy_id"`
	AssetName    string     `json:"asset_name"`
	Status       string     `json:"status"` // "loading", "live" or "error"
	Error        string     `json:"error,omitempty"`
	PriceADA     float64    `json:"price_ada"`
	PriceUSD     float64    `json:"price_usd,omitempty"` // via the USDM pool; 0 if unknown
	Change1h     float64    `json:"change_1h"`
	Change24h    float64    `json:"change_24h"`
	Volume24hADA float64    `json:"volume_24h_ada"`
	TotalSupply  float64    `json:"total_supply,omitempty"` // whole tokens
	FDVADA       float64    `json:"fdv_ada,omitempty"`      // price × total supply
	Holders      int        `json:"holders,omitempty"`      // wallets holding the token; 0 until loaded
	HasLogo      bool       `json:"has_logo"`
	Sparkline    []float64  `json:"sparkline"` // hourly prices over the loaded history, oldest first
	ReserveADA   float64    `json:"reserve_ada"`
	ReserveToken float64    `json:"reserve_token"`
	Decimals     int        `json:"decimals"`
	PoolUTxO     string     `json:"pool_utxo,omitempty"`
	LastSwap     *time.Time `json:"last_swap,omitempty"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`

	// The last 24 hours of trades (MinTradeADA and up).
	Trades24h     int     `json:"trades_24h"`
	Buys24h       int     `json:"buys_24h"`
	Sells24h      int     `json:"sells_24h"`
	BuyVol24hADA  float64 `json:"buy_volume_24h_ada"`
	SellVol24hADA float64 `json:"sell_volume_24h_ada"`

	// From the token registry and asset_info; empty if unknown.
	Fingerprint string     `json:"fingerprint,omitempty"`
	Created     *time.Time `json:"created,omitempty"`
	Description string     `json:"description,omitempty"`
	Website     string     `json:"website,omitempty"`
}

// Trade is one swap against a token's pool, worked out from how the pool's
// reserves changed. Minswap fills orders in batches, so a trade can be several
// orders filled in the same transaction.
type Trade struct {
	Time        time.Time `json:"time"`
	Side        string    `json:"side"`         // "buy" (ADA in, token out) or "sell"
	PriceADA    float64   `json:"price_ada"`    // ADA paid or received per whole token
	AmountADA   float64   `json:"amount_ada"`   // ADA in or out of the pool
	AmountToken float64   `json:"amount_token"` // whole tokens out of or into the pool
	TxHash      string    `json:"tx_hash"`
}

// MaxTrades is the most recent trades the API returns for one token.
const MaxTrades = 50

type tokenState struct {
	token      Token
	slug       string // unique id for page addresses
	lp         koios.Asset
	poolRef    string // current pool UTxO "tx_hash#index"; empty until found
	address    string // address the pool lives at
	lastHeight int    // block height of the newest pool state we read
	last       minswap.Pool
	points     []Point  // one per swap, oldest first
	trades     []Trade  // swaps with amounts, oldest first; last 24h only
	swaps      int      // swaps read since start
	readFrom   int      // when the pool is spent but its swap not read yet: block to read from
	supply     *big.Int // total supply in the smallest unit, from asset_info
	logo       []byte   // PNG from the token registry
	holders    int      // from asset_summary, loaded in the background
	info       Quote    // fingerprint, created, description and website from asset_info
	quote      Quote
}

// poolState is a pool right after one transaction.
type poolState struct {
	pool   minswap.Pool
	height int
}

// Tracker finds every token's Minswap pool and loads the last 24 hours of swaps
// at startup, then follows new swaps on a fixed interval. Everything is kept in
// memory only.
type Tracker struct {
	client   *koios.Client
	interval time.Duration
	log      *log.Logger
	now      func() time.Time

	mu        sync.RWMutex
	states    []*tokenState
	loaded    bool
	infoReady bool // supply and logos loaded
	lastCheck int  // chain block height at the previous check
}

func NewTracker(client *koios.Client, tokens []Token, interval time.Duration, logger *log.Logger) *Tracker {
	t := &Tracker{client: client, interval: interval, log: logger, now: time.Now}
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

// Quotes returns the top TopN tokens ranked by 24h trading volume. Only tokens
// with a Minswap pool holding at least MinPoolADA are ranked; tokens that did
// not trade in 24h follow, by pool size.
func (t *Tracker) Quotes() []Quote {
	t.mu.RLock()
	defer t.mu.RUnlock()
	now := t.now()
	cut := now.Add(-24 * time.Hour)
	usdPerADA := t.usdPerADA()
	var out []Quote
	for _, s := range t.tracked() {
		q := s.quote
		q.PriceUSD = q.PriceADA * usdPerADA
		q.Fingerprint, q.Created, q.Description, q.Website = s.info.Fingerprint, s.info.Created, s.info.Description, s.info.Website
		for _, tr := range s.trades {
			if tr.Time.Before(cut) {
				continue
			}
			q.Trades24h++
			if tr.Side == "buy" {
				q.Buys24h++
				q.BuyVol24hADA += tr.AmountADA
			} else {
				q.Sells24h++
				q.SellVol24hADA += tr.AmountADA
			}
		}
		if s.supply != nil {
			q.TotalSupply = scaled(s.supply, q.Decimals)
			q.FDVADA = q.PriceADA * q.TotalSupply
		}
		q.HasLogo = len(s.logo) > 0
		q.Holders = s.holders
		if len(s.points) > 0 {
			q.Change1h = changePct(s.points, now.Add(-time.Hour).Unix())
			q.Change24h = changePct(s.points, now.Add(-24*time.Hour).Unix())
			q.Volume24hADA = volumeSince(s.points, now.Add(-24*time.Hour).Unix())
			spark := bucket(s.points, now.Add(-HistorySpan), now, time.Hour)
			q.Sparkline = make([]float64, len(spark))
			for j, p := range spark {
				q.Sparkline[j] = p.Price
			}
		}
		out = append(out, q)
	}
	// Highest 24h volume first; equal volumes (usually no trades) by pool size.
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Volume24hADA != out[j].Volume24hADA {
			return out[i].Volume24hADA > out[j].Volume24hADA
		}
		return out[i].ReserveADA > out[j].ReserveADA
	})
	if len(out) > TopN {
		out = out[:TopN]
	}
	for i := range out {
		out[i].Rank = i + 1
	}
	return out
}

// usdPerADA is the ADA/USD rate from the USDM pool (1 USDM ≈ 1 USD), or 0 if
// that pool is not tracked. Call with t.mu held.
func (t *Tracker) usdPerADA() float64 {
	for _, s := range t.states {
		if s.token.Asset == USDM && s.quote.PriceADA > 0 {
			return 1 / s.quote.PriceADA
		}
	}
	return 0
}

// Chart returns evenly spaced points for one token and range.
func (t *Tracker) Chart(id, rangeKey string) (Quote, []ChartPoint, error) {
	r, ok := Ranges[strings.ToUpper(rangeKey)]
	if !ok {
		return Quote{}, nil, fmt.Errorf("unknown range %q: use 24H", rangeKey)
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, s := range t.states {
		if !strings.EqualFold(s.slug, id) {
			continue
		}
		if len(s.points) == 0 {
			return s.quote, []ChartPoint{}, nil // still loading
		}
		now := t.now()
		return s.quote, bucket(s.points, now.Add(-r.Span), now, r.Interval), nil
	}
	return Quote{}, nil, fmt.Errorf("unknown token %q", id)
}

// Trades returns a token's most recent trades of the last 24 hours, newest
// first, at most limit (and never more than MaxTrades).
func (t *Tracker) Trades(id string, limit int) (Quote, []Trade, error) {
	if limit <= 0 || limit > MaxTrades {
		limit = MaxTrades
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	cut := t.now().Add(-HistorySpan)
	for _, s := range t.states {
		if !strings.EqualFold(s.slug, id) {
			continue
		}
		out := []Trade{}
		for i := len(s.trades) - 1; i >= 0 && len(out) < limit; i-- {
			if s.trades[i].Time.Before(cut) {
				break
			}
			out = append(out, s.trades[i])
		}
		return s.quote, out, nil
	}
	return Quote{}, nil, fmt.Errorf("unknown token %q", id)
}

// Logo returns a token's PNG logo from the token registry, if it has one.
func (t *Tracker) Logo(id string) ([]byte, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, s := range t.states {
		if strings.EqualFold(s.slug, id) && len(s.logo) > 0 {
			return s.logo, true
		}
	}
	return nil, false
}

func (t *Tracker) refresh(ctx context.Context) {
	if !t.loaded {
		t.loadHistory(ctx)
		if t.loaded {
			t.loadInfo(ctx)
		}
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

// loadInfo reads total supply, ticker and logo for every tracked token with
// asset_info, 25 tokens per request. If it fails, prices still work and it is
// tried again next minute.
func (t *Tracker) loadInfo(ctx context.Context) {
	t.mu.RLock()
	act := t.tracked()
	t.mu.RUnlock()
	if len(act) == 0 {
		return
	}
	var infos []koios.AssetInfo
	for start := 0; start < len(act); start += 25 {
		end := min(start+25, len(act))
		assets := make([]koios.Asset, 0, end-start)
		for _, s := range act[start:end] {
			assets = append(assets, s.token.Asset)
		}
		batch, err := t.client.AssetInfos(ctx, assets)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				t.log.Printf("Could not load token supply and logos: %v (trying again in %s)", err, t.interval)
			}
			return
		}
		infos = append(infos, batch...)
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	logos := 0
	for _, info := range infos {
		for _, s := range act {
			if s.token.Asset.PolicyID != info.PolicyID || s.token.Asset.AssetName != info.AssetName {
				continue
			}
			if n, ok := new(big.Int).SetString(info.TotalSupply, 10); ok {
				s.supply = n
			}
			s.info.Fingerprint = info.Fingerprint
			if info.CreationTime > 0 {
				created := time.Unix(info.CreationTime, 0).UTC()
				s.info.Created = &created
			}
			if reg := info.TokenRegistry; reg != nil {
				if reg.Ticker != "" {
					s.quote.Ticker = reg.Ticker
				}
				s.info.Description = strings.TrimSpace(reg.Description)
				s.info.Website = safeURL(reg.URL)
				if png, err := base64.StdEncoding.DecodeString(reg.Logo); err == nil && len(png) > 0 {
					s.logo = png
					logos++
				}
			}
		}
	}
	t.infoReady = true
	t.log.Printf("Token details loaded: total supply for %d tokens, %d logos", len(infos), logos)
}

func (t *Tracker) holdersLoop(ctx context.Context) {
	for {
		t.loadHolders(ctx)
		if err := sleepCtx(ctx, HoldersEvery); err != nil {
			return
		}
	}
}

// loadHolders counts the holders of each token in the current top list:
// wallets with a stake key plus addresses without one. One slow asset_summary
// request per token.
func (t *Tracker) loadHolders(ctx context.Context) {
	top := t.Quotes()
	t.log.Printf("Counting holders for %d tokens in the background (slow, can take an hour)…", len(top))
	for _, q := range top {
		if ctx.Err() != nil {
			return
		}
		var s *tokenState
		t.mu.RLock()
		for _, st := range t.states {
			if st.slug == q.ID {
				s = st
			}
		}
		t.mu.RUnlock()
		if s == nil {
			continue
		}
		start := time.Now()
		sum, err := t.client.AssetSummary(ctx, s.token.Asset)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				t.log.Printf("  %-8s could not count holders: %v (trying again in %s)", q.Ticker, err, HoldersEvery)
			}
			continue
		}
		n := sum.StakedWallets + sum.UnstakedAddresses
		t.mu.Lock()
		s.holders = n
		t.mu.Unlock()
		t.log.Printf("  %-8s %s holders (counted in %s)", q.Ticker, group(float64(n)), time.Since(start).Round(time.Second))
	}
}

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

// loadHistory first finds the current pool of every token (asset_utxos, 30
// tokens per request), so tokens that did not trade today are found too. Then
// it reads every Minswap V2 transaction of the last 24 hours once for the
// charts, changes and volume.
func (t *Tracker) loadHistory(ctx context.Context) {
	start, before := time.Now(), t.client.Requests()
	tip, err := t.client.Tip(ctx)
	if err != nil {
		t.failAll(err)
		return
	}

	t.log.Printf("Finding the Minswap pools of %d tokens…", len(t.states))
	current, err := t.findPools(ctx)
	if err != nil {
		t.failAll(err)
		return
	}
	t.log.Printf("Found %d Minswap pools for %d tokens (%d Koios requests, %s)",
		len(current), len(t.states), t.client.Requests()-before, time.Since(start).Round(time.Second))

	histStart, histBefore := time.Now(), t.client.Requests()
	from := tip.BlockHeight - int((HistorySpan+time.Hour)/time.Second)/secondsPerBlock
	t.log.Printf("Loading the last 24 hours of Minswap transactions from Koios…")
	txs, err := t.client.AddressTxs(ctx, []string{PoolAddress}, from, func(rows int) {
		t.log.Printf("  listed %s transactions", group(float64(rows)))
	})
	if err != nil {
		t.failAll(err)
		return
	}
	found, err := t.readPools(ctx, txs, func(done, total int) {
		t.log.Printf("  read %s of %s transactions", group(float64(done)), group(float64(total)))
	})
	if err != nil {
		t.failAll(err)
		return
	}

	t.mu.Lock()
	traded, listed := 0, 0
	for _, s := range t.states {
		s.points, s.trades, s.last, s.swaps, s.lastHeight = nil, nil, minswap.Pool{}, 0, 0
		t.record(s, found[s])
		if len(found[s]) > 0 {
			traded++
		}
		// The pool as it is now, if it is newer than the last swap in the history
		// (always the case for a token that did not trade in 24h).
		if cur, ok := current[s]; ok && cur.height > s.lastHeight {
			t.record(s, []poolState{cur})
		}
		if s.poolRef != "" && s.quote.ReserveADA >= MinPoolADA {
			listed++
		}
	}
	t.loaded = true
	t.lastCheck = tip.BlockHeight
	t.mu.Unlock()

	t.log.Printf("History loaded: %d tokens traded on Minswap in 24h, %d pools with at least ₳%s are listed (%s transactions, %d Koios requests, %s)",
		traded, listed, group(MinPoolADA), group(float64(len(txs))),
		t.client.Requests()-histBefore, time.Since(histStart).Round(time.Second))
}

// findPools asks Koios which UTxOs at the Minswap pool address hold each
// token's LP token: that UTxO is the token's pool, as it is now.
func (t *Tracker) findPools(ctx context.Context) (map[*tokenState]poolState, error) {
	byLP := map[string]*tokenState{}
	lps := make([]koios.Asset, 0, len(t.states))
	for _, s := range t.states {
		byLP[s.lp.AssetName] = s
		lps = append(lps, s.lp)
	}
	utxos, err := t.client.PoolUTxOs(ctx, PoolAddress, lps, func(done, total int) {
		if done == total || done%120 == 0 {
			t.log.Printf("  checked %d of %d tokens", done, total)
		}
	})
	if err != nil {
		return nil, err
	}
	out := map[*tokenState]poolState{}
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
		if s.address == "" {
			s.address = u.Address
		}
		out[s] = poolState{pool: p, height: u.BlockHeight}
	}
	return out, nil
}

// update checks the chain tip and, with utxo_info (50 pools per request), which
// pools were spent (a swap happened) since the previous check. Only then it
// reads Minswap's transactions since that check and takes the new pool states.
func (t *Tracker) update(ctx context.Context) {
	t.mu.RLock()
	known := t.tracked()
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
	for start := 0; start < len(known); start += 50 {
		end := min(start+50, len(known))
		refs := make([]string, 0, end-start)
		for _, s := range known[start:end] {
			refs = append(refs, s.poolRef)
		}
		batch, err := t.client.UTxOInfo(ctx, refs)
		if err != nil {
			t.failAll(err)
			return
		}
		utxos = append(utxos, batch...)
	}
	// A pool spent now was spent after the previous check, so its swap is in a
	// block from lastCheck on (minus a few blocks of safety margin).
	since := t.lastCheck - 3
	t.lastCheck = tip.BlockHeight
	unspent := map[string]bool{}
	for _, u := range utxos {
		if !u.IsSpent {
			unspent[u.Ref()] = true
		}
	}

	var spent []*tokenState
	addresses := map[string]bool{PoolAddress: true}
	fromHeight := 0
	for _, s := range known {
		if unspent[s.poolRef] {
			continue
		}
		spent = append(spent, s)
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

	found := map[*tokenState][]poolState{}
	if len(spent) > 0 {
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
		t.record(s, fresh)
		s.readFrom = 0
		q := s.quote
		t.mu.Unlock()
		t.log.Printf("  %-8s %-16s %-9s %d swap%s",
			q.Ticker, FormatADA(q.PriceADA), change(prev, q.PriceADA), len(fresh), plural(len(fresh)))
	}
	t.log.Printf("Checked %d pools: %d had new swaps", len(known), changed)
}

// readPools reads transactions with tx_info (50 per request) and returns, for
// each candidate token, the pool state after each of its swaps, oldest first.
func (t *Tracker) readPools(ctx context.Context, txs []koios.AddressTx, progress func(done, total int)) (map[*tokenState][]poolState, error) {
	byLP := map[string]*tokenState{}
	for _, s := range t.states {
		byLP[s.lp.AssetName] = s
	}

	var all []koios.Tx
	for start := 0; start < len(txs); start += 50 {
		end := min(start+50, len(txs))
		hashes := make([]string, 0, end-start)
		for _, tx := range txs[start:end] {
			hashes = append(hashes, tx.TxHash)
		}
		batch, err := t.client.TxInfo(ctx, hashes)
		if err != nil {
			return nil, err
		}
		all = append(all, batch...)
		if progress != nil && (end == len(txs) || (end/50)%10 == 0) {
			progress(end, len(txs))
		}
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].BlockHeight != all[j].BlockHeight {
			return all[i].BlockHeight < all[j].BlockHeight
		}
		return all[i].TxBlockIndex < all[j].TxBlockIndex
	})

	found := map[*tokenState][]poolState{}
	for _, tx := range all {
		for _, o := range tx.Outputs {
			if o.PaymentAddr.Cred != minswap.PoolCredential {
				continue
			}
			p, err := minswap.Parse(o.UTxO(tx))
			if err != nil {
				continue
			}
			s, ok := byLP[p.LPAsset.AssetName]
			if !ok || !p.IsADAPair(s.token.Asset) {
				continue
			}
			if s.address == "" {
				s.address = o.PaymentAddr.Bech32
			}
			found[s] = append(found[s], poolState{pool: p, height: tx.BlockHeight})
		}
	}
	return found, nil
}

// record adds pool states to the history and updates the quote. Call with t.mu held.
func (t *Tracker) record(s *tokenState, states []poolState) {
	if len(states) == 0 {
		return
	}
	for _, st := range states {
		p := st.pool
		vol := 0.0
		// A swap moves ADA in or out without changing the LP supply; adding or
		// removing liquidity changes the LP supply and is not a trade.
		if s.last.ReserveA != nil && p.Liquidity.Cmp(s.last.Liquidity) == 0 {
			diffA := new(big.Int).Sub(p.ReserveA, s.last.ReserveA)
			diffB := new(big.Int).Sub(p.ReserveB, s.last.ReserveB)
			vol = math.Abs(scaled(diffA, 6))
			if tr, ok := tradeFrom(p, diffA, diffB); ok {
				s.trades = append(s.trades, tr)
			}
		}
		s.points = append(s.points, Point{Time: p.LastSwap.Unix(), Price: p.PriceADA(), VolumeADA: vol})
		s.last = p
		s.lastHeight = st.height
		s.swaps++
	}
	s.points = trim(s.points, t.now())
	cut := t.now().Add(-HistorySpan)
	for len(s.trades) > 0 && s.trades[0].Time.Before(cut) {
		s.trades = s.trades[1:]
	}

	p := s.last
	lastSwap, now := p.LastSwap, t.now()
	s.poolRef = p.UTxO
	s.quote.Status, s.quote.Error = "live", ""
	s.quote.PriceADA = p.PriceADA()
	s.quote.ReserveADA = p.ReserveADA()
	s.quote.ReserveToken = p.ReserveToken()
	s.quote.Decimals = p.DecimalsB
	s.quote.PoolUTxO = p.UTxO
	s.quote.LastSwap = &lastSwap
	s.quote.UpdatedAt = &now
}

// tradeFrom turns the change in a pool's reserves into a trade. ADA going in
// while tokens come out is a buy; the other way round is a sell.
func tradeFrom(p minswap.Pool, diffADA, diffToken *big.Int) (Trade, bool) {
	if diffADA.Sign() == 0 || diffToken.Sign() == 0 || diffADA.Sign() == diffToken.Sign() {
		return Trade{}, false
	}
	ada := math.Abs(scaled(diffADA, 6))
	tok := math.Abs(scaled(diffToken, p.DecimalsB))
	if ada < MinTradeADA || tok == 0 {
		return Trade{}, false
	}
	side := "sell"
	if diffADA.Sign() > 0 {
		side = "buy"
	}
	hash, _, _ := strings.Cut(p.UTxO, "#")
	return Trade{Time: p.LastSwap, Side: side, PriceADA: ada / tok, AmountADA: ada, AmountToken: tok, TxHash: hash}, true
}

// safeURL keeps a registry website only if it is a plain http(s) link, so a
// token's metadata cannot put another kind of link on the page.
func safeURL(raw string) string {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return ""
	}
	return u.String()
}

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

func scaled(n *big.Int, decimals int) float64 {
	d := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	f, _ := new(big.Rat).SetFrac(n, d).Float64()
	return f
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// FormatADA shows a price with 4 significant digits, e.g. ₳0.002657.
func FormatADA(v float64) string {
	if v <= 0 {
		return "₳0"
	}
	if v >= 1 {
		return fmt.Sprintf("₳%.4f", v)
	}
	decimals := 3 - int(math.Floor(math.Log10(v)))
	return fmt.Sprintf("₳%.*f", decimals, v)
}

func change(prev, now float64) string {
	switch {
	case prev == 0:
		return ""
	case now > prev:
		return fmt.Sprintf("▲ %.2f%%", (now/prev-1)*100)
	case now < prev:
		return fmt.Sprintf("▼ %.2f%%", (1-now/prev)*100)
	default:
		return "= 0.00%"
	}
}

// group formats a number as 1,977,518.
func group(v float64) string {
	s := fmt.Sprintf("%.0f", v)
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	return b.String()
}
