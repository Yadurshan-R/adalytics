package market

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"
)

// TopTrader is what one wallet did in one token over a time range: a row of
// the Top Traders tab on the token page, like Birdeye's and DexScreener's.
type TopTrader struct {
	Wallet    string    `json:"wallet"`
	Buys      int       `json:"buys"`
	Sells     int       `json:"sells"`
	BuyADA    float64   `json:"buy_ada"`    // ADA spent buying
	SellADA   float64   `json:"sell_ada"`   // ADA received selling
	BuyToken  float64   `json:"buy_token"`  // whole tokens bought
	SellToken float64   `json:"sell_token"` // whole tokens sold
	Last      time.Time `json:"last"`       // its latest trade in this token
}

// add counts one more trade of the wallet.
func (w *TopTrader) add(tr Trade) {
	if tr.Side == "buy" {
		w.Buys++
		w.BuyADA += tr.AmountADA
		w.BuyToken += tr.AmountToken
	} else {
		w.Sells++
		w.SellADA += tr.AmountADA
		w.SellToken += tr.AmountToken
	}
	if tr.Time.After(w.Last) {
		w.Last = tr.Time
	}
}

// TraderHours are the time ranges the Top Traders tab offers: 24 hours, 7
// days and 30 days (the history the backend keeps). Others become 24.
var TraderHours = []int{24, 168, 720}

// MaxTopTraders is how many wallets the Top Traders tab gets.
const MaxTopTraders = 50

// MaxWalletTrades is how many of one wallet's trades the Trades tab gets.
const MaxWalletTrades = 200

// inRange calls fn for each of a token's trades in the last hours (24 if
// hours is not one of TraderHours), newest first. The caller holds t.mu.
func (t *Tracker) inRange(id string, hours int, fn func(Trade)) (Quote, error) {
	if !slices.Contains(TraderHours, hours) {
		hours = 24
	}
	cut := t.now().Add(-time.Duration(hours) * time.Hour)
	for _, s := range t.states {
		if !strings.EqualFold(s.slug, id) {
			continue
		}
		for i := len(s.trades) - 1; i >= 0 && !s.trades[i].Time.Before(cut); i-- {
			fn(s.trades[i])
		}
		return s.quote, nil
	}
	return Quote{}, fmt.Errorf("unknown token %q", id)
}

// TopTraders returns the wallets that traded a token most in the last hours,
// at most MaxTopTraders. sort "net" ranks by net buying (ADA spent buying
// minus ADA received selling, as the wallet card shows it); anything else
// ranks by volume (ADA bought plus sold). Trades whose wallet is not known
// are left out. It reads only the trades in memory: no Koios requests.
func (t *Tracker) TopTraders(id string, hours int, sortBy string) (Quote, []TopTrader, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	by := map[string]*TopTrader{}
	quote, err := t.inRange(id, hours, func(tr Trade) {
		if tr.Wallet == "" {
			return
		}
		if by[tr.Wallet] == nil {
			by[tr.Wallet] = &TopTrader{Wallet: tr.Wallet}
		}
		by[tr.Wallet].add(tr)
	})
	if err != nil {
		return Quote{}, nil, err
	}
	out := make([]TopTrader, 0, len(by))
	for _, w := range by {
		out = append(out, *w)
	}
	rank := func(w TopTrader) float64 { return w.BuyADA + w.SellADA }
	if sortBy == "net" {
		rank = func(w TopTrader) float64 { return w.BuyADA - w.SellADA }
	}
	sort.Slice(out, func(i, j int) bool {
		if a, b := rank(out[i]), rank(out[j]); a != b {
			return a > b
		}
		return out[i].Wallet < out[j].Wallet // the same order every time
	})
	return quote, out[:min(len(out), MaxTopTraders)], nil
}

// TraderTrades returns one wallet's trades in a token over the last hours,
// newest first (at most MaxWalletTrades), and its totals over all of them,
// so the wallet card matches its Top Traders row even when the list is cut.
func (t *Tracker) TraderTrades(id, wallet string, hours int) (Quote, TopTrader, []Trade, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	sum, out := TopTrader{Wallet: wallet}, []Trade{}
	quote, err := t.inRange(id, hours, func(tr Trade) {
		if wallet == "" || tr.Wallet != wallet {
			return
		}
		sum.add(tr)
		if len(out) < MaxWalletTrades {
			out = append(out, tr)
		}
	})
	return quote, sum, out, err
}
