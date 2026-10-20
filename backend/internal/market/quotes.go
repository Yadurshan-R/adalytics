package market

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

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
	Change7d     *float64   `json:"change_7d,omitempty"` // only once 7 days of history are loaded
	Volume24hADA float64    `json:"volume_24h_ada"`
	TotalSupply  float64    `json:"total_supply,omitempty"` // whole tokens
	FDVADA       float64    `json:"fdv_ada,omitempty"`      // price × total supply
	Holders      int        `json:"holders,omitempty"`      // wallets holding the token; 0 until loaded
	HasLogo      bool       `json:"has_logo"`
	Sparkline    []float64  `json:"sparkline"`              // prices over the last 24h (hourly), oldest first
	Sparkline7d  []float64  `json:"sparkline_7d,omitempty"` // prices over the last 7 days (every 2 hours), once loaded
	ReserveADA   float64    `json:"reserve_ada"`
	ReserveToken float64    `json:"reserve_token"`
	Decimals     int        `json:"decimals"`
	PoolUTxO     string     `json:"pool_utxo,omitempty"`
	LastSwap     *time.Time `json:"last_swap,omitempty"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`

	// The last 24 hours of trades (MinTradeADA and up).
	Trades24h     int      `json:"trades_24h"`
	Traders24h    int      `json:"traders_24h"` // different wallets that traded (wallets we could tell)
	Buys24h       int      `json:"buys_24h"`
	Sells24h      int      `json:"sells_24h"`
	BuyVol24hADA  float64  `json:"buy_volume_24h_ada"`
	SellVol24hADA float64  `json:"sell_volume_24h_ada"`
	Periods       []Period `json:"periods"` // 5m, 1h, 4h and 24h: change, volume, buys and sells

	// From the token registry and asset_info; empty if unknown.
	Fingerprint string     `json:"fingerprint,omitempty"`
	Created     *time.Time `json:"created,omitempty"`
	Description string     `json:"description,omitempty"`
	Website     string     `json:"website,omitempty"`
	PolicyState string     `json:"policy_state,omitempty"` // "locked", "open", "script" or "unknown" (policy.go)
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
	Fill        int       `json:"fill"`             // which order in the transaction: one wallet can have two in one batch
	Wallet      string    `json:"wallet,omitempty"` // stake address of the trader; "" if not known
}

// MaxTrades is the most recent trades the API returns for one token.
const MaxTrades = 50

// Quotes returns the top TopN tokens ranked by 24h trading volume. Only tokens
// with a Minswap pool holding at least MinPoolADA are ranked; tokens that did
// not trade in 24h follow, by pool size.
func (t *Tracker) Quotes() []Quote {
	t.mu.RLock()
	defer t.mu.RUnlock()
	now := t.now()
	cut := now.Add(-24 * time.Hour)
	usdPerADA := t.usdPerADA()
	week := t.rangeReady(RangeList[1])
	var out []Quote
	for _, s := range t.tracked() {
		q := s.quote
		q.PriceUSD = q.PriceADA * usdPerADA
		q.Fingerprint, q.Created, q.Description, q.Website = s.info.Fingerprint, s.info.Created, s.info.Description, s.info.Website
		traders := map[string]bool{}
		for i := len(s.trades) - 1; i >= 0; i-- {
			tr := s.trades[i]
			if tr.Time.Before(cut) {
				break
			}
			if tr.Wallet != "" {
				traders[tr.Wallet] = true
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
		q.Traders24h = len(traders)
		q.PolicyState = s.mint.state(now)
		q.Periods = periodsOf(s, now)
		q.HasLogo = len(s.logo) > 0
		q.Holders = s.holders
		if len(s.points) > 0 {
			q.Change1h = changePct(s.points, now.Add(-time.Hour).Unix())
			q.Change24h = changePct(s.points, now.Add(-24*time.Hour).Unix())
			q.Volume24hADA = volumeSince(s.points, now.Add(-24*time.Hour).Unix())
			q.Sparkline = sparkline(s.points, now.Add(-24*time.Hour), now, time.Hour)
			if week {
				c := changePct(s.points, now.Add(-7*24*time.Hour).Unix())
				q.Change7d = &c
				q.Sparkline7d = sparkline(s.points, now.Add(-7*24*time.Hour), now, 2*time.Hour)
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

// sparkline is the price at every step from start to end, for the list's small lines.
func sparkline(points []Point, start, end time.Time, step time.Duration) []float64 {
	spark := bucket(points, start, end, step)
	out := make([]float64, len(spark))
	for i, p := range spark {
		out[i] = p.Price
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

// ErrRangeLoading means the range's history is still loading.
var ErrRangeLoading = errors.New("this range is still loading")

// Chart returns one token's chart for a range: evenly spaced line points, or
// candles when candles is true.
func (t *Tracker) Chart(id, rangeKey string, candlesticks bool) (Quote, Range, []ChartPoint, []Candle, error) {
	r, ok := RangeByKey(rangeKey)
	if !ok {
		return Quote{}, Range{}, nil, nil, fmt.Errorf("unknown range %q: use 24H, 7D or 30D", rangeKey)
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, s := range t.states {
		if !strings.EqualFold(s.slug, id) {
			continue
		}
		if !t.rangeReady(r) {
			return s.quote, r, nil, nil, ErrRangeLoading
		}
		if len(s.points) == 0 {
			return s.quote, r, []ChartPoint{}, []Candle{}, nil // still loading
		}
		now := t.now()
		if candlesticks {
			return s.quote, r, nil, candles(s.points, now.Add(-r.Span), now, r.CandleInterval), nil
		}
		return s.quote, r, bucket(s.points, now.Add(-r.Span), now, r.Interval), nil, nil
	}
	return Quote{}, r, nil, nil, fmt.Errorf("unknown token %q", id)
}

// Trades returns a token's most recent trades of the last 24 hours, newest
// first, at most limit (and never more than MaxTrades).
func (t *Tracker) Trades(id string, limit int) (Quote, []Trade, error) {
	if limit <= 0 || limit > MaxTrades {
		limit = MaxTrades
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	cut := t.now().Add(-24 * time.Hour)
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
