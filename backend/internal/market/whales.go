package market

import (
	"slices"
	"sort"
	"strings"
	"time"
)

// WhaleTrade is one trade on the Whale Trades page: the trade, which token it
// was in, and how big it was against that token's pool.
type WhaleTrade struct {
	Trade
	ID        string  `json:"id"`         // the token's page id
	Ticker    string  `json:"ticker"`     // the token traded
	HasLogo   bool    `json:"has_logo"`   // whether /api/tokens/{id}/logo has a logo
	PoolShare float64 `json:"pool_share"` // the trade's ADA as % of the pool's ADA, e.g. 2.4
}

// WhaleFilter is what the Whale Trades page asks for. Empty fields mean all.
type WhaleFilter struct {
	Hours  int     // 1, 4, 24, 168 (7 days) or 720 (30 days); anything else is 24
	MinADA float64 // the size picked: 0 (all, fish), 10,000 (dolphin) or 50,000 (whale)
	MaxADA float64 // below this: 10,000 (fish), 50,000 (dolphin), or 0 for no top
	Token  string  // a token's page id
	Wallet string  // a stake1… or addr1… address
	Offset int     // how many matching trades to skip, for Show more
	Limit  int     // how many to send: PageWhales unless asked, at most MaxWhales
}

// WhaleHours are the time choices on the page, in hours.
var WhaleHours = []int{1, 4, 24, 168, 720}

// PageWhales is how many trades one page (and each Show more) gets.
const PageWhales = 200

// MaxWhales is the most trades one answer sends: a page refreshing after a
// few Show mores asks for all its rows again at once.
const MaxWhales = 1000

// BigTradeADA lets a trade in even when it is less than 1% of a deep pool:
// ₳10,000 and more is big for any token.
const BigTradeADA = 10_000

// WhaleTrades returns one page of the big trades that match the filter,
// newest first, and whether there are more. A big trade is at least 1% of
// its pool's ADA (WhaleADA) or ₳10,000 and more (BigTradeADA). Trades are
// kept for 30 days (the chart history), and this reads only the trades
// already in memory, so it costs no Koios requests.
func (t *Tracker) WhaleTrades(f WhaleFilter) ([]WhaleTrade, bool) {
	if !slices.Contains(WhaleHours, f.Hours) {
		f.Hours = 24
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	cut := t.now().Add(-time.Duration(f.Hours) * time.Hour)
	out := []WhaleTrade{}
	for _, s := range t.tracked() {
		if f.Token != "" && s.slug != f.Token {
			continue
		}
		whale := WhaleADA(s.quote)
		for i := len(s.trades) - 1; i >= 0 && !s.trades[i].Time.Before(cut); i-- {
			tr := s.trades[i]
			if whale <= 0 || (tr.AmountADA < whale && tr.AmountADA < BigTradeADA) || tr.AmountADA < f.MinADA || (f.MaxADA > 0 && tr.AmountADA >= f.MaxADA) {
				continue
			}
			if f.Wallet != "" && !strings.EqualFold(tr.Wallet, f.Wallet) {
				continue
			}
			share := tr.AmountADA / s.quote.ReserveADA * 100
			out = append(out, WhaleTrade{Trade: tr, ID: s.slug, Ticker: s.quote.Ticker, HasLogo: len(s.logo) > 0, PoolShare: share})
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Time.After(out[j].Time) })
	start := min(max(f.Offset, 0), len(out))
	if f.Limit <= 0 {
		f.Limit = PageWhales
	}
	end := min(start+min(f.Limit, MaxWhales), len(out))
	return out[start:end], end < len(out)
}

// WhaleToken is one token in the page's token search.
type WhaleToken struct {
	ID      string `json:"id"`
	Ticker  string `json:"ticker"`
	Name    string `json:"name"`
	HasLogo bool   `json:"has_logo"` // whether /api/tokens/{id}/logo has a logo
}

// WhaleTokens lists every listed token for the page's token search, the
// biggest pool first, so the first suggestions are the best-known tokens.
func (t *Tracker) WhaleTokens() []WhaleToken {
	t.mu.RLock()
	defer t.mu.RUnlock()
	states := t.tracked()
	sort.SliceStable(states, func(i, j int) bool { return states[i].quote.ReserveADA > states[j].quote.ReserveADA })
	out := []WhaleToken{}
	for _, s := range states {
		out = append(out, WhaleToken{ID: s.slug, Ticker: s.quote.Ticker, Name: s.quote.Name, HasLogo: len(s.logo) > 0})
	}
	return out
}
