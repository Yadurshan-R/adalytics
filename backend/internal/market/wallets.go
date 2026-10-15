package market

import (
	"sort"
	"time"
)

// WhaleShare marks a trade as a whale trade when it moves at least this part
// of the pool's ADA (1%). A share of the pool, not a fixed amount, fits both
// big and small tokens: either way such a trade moves the price about 2%.
const WhaleShare = 0.01

// WalletToken is what one wallet did in one token over the last 24 hours.
type WalletToken struct {
	ID      string    `json:"id"`
	Ticker  string    `json:"ticker"`
	Buys    int       `json:"buys"`
	Sells   int       `json:"sells"`
	BuyADA  float64   `json:"buy_ada"`  // ADA spent buying
	SellADA float64   `json:"sell_ada"` // ADA received selling
	Last    time.Time `json:"last"`     // its latest trade in this token
}

// WhaleADA is the smallest trade that counts as a whale trade for a token.
func WhaleADA(q Quote) float64 { return q.ReserveADA * WhaleShare }

// Wallet returns every token one wallet traded in the last 24 hours, its
// latest trade first. It reads only the trades already in memory.
func (t *Tracker) Wallet(wallet string) []WalletToken {
	t.mu.RLock()
	defer t.mu.RUnlock()
	cut := t.now().Add(-24 * time.Hour)
	out := []WalletToken{}
	for _, s := range t.states {
		w := WalletToken{ID: s.slug, Ticker: s.quote.Ticker}
		for i := len(s.trades) - 1; i >= 0 && !s.trades[i].Time.Before(cut); i-- {
			tr := s.trades[i]
			if tr.Wallet != wallet {
				continue
			}
			if tr.Side == "buy" {
				w.Buys++
				w.BuyADA += tr.AmountADA
			} else {
				w.Sells++
				w.SellADA += tr.AmountADA
			}
			if tr.Time.After(w.Last) {
				w.Last = tr.Time
			}
		}
		if w.Buys+w.Sells > 0 {
			out = append(out, w)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Last.After(out[j].Last) })
	return out
}
