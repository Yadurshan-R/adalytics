package market

import (
	"fmt"
	"strings"
	"time"
)

// Period is a token's trading over one recent period (5m, 1h, 4h or 24h), for
// the token list's columns and the activity box on the token page: how the price moved, how much was
// traded, and whether buyers or sellers were stronger.
type Period struct {
	Key       string  `json:"key"`
	ChangePct float64 `json:"change_pct"` // price change over the period
	VolumeADA float64 `json:"volume_ada"` // everything traded, small trades too
	Buys      int     `json:"buys"`
	Sells     int     `json:"sells"`
	BuyADA    float64 `json:"buy_ada"`  // ADA spent buying
	SellADA   float64 `json:"sell_ada"` // ADA received selling
}

// PeriodList is every period in the activity box, shortest first.
var PeriodList = []struct {
	Key  string
	Span time.Duration
}{
	{"5m", 5 * time.Minute},
	{"1h", time.Hour},
	{"4h", 4 * time.Hour},
	{"24h", 24 * time.Hour},
}

// Periods returns one token's trading over the last 5 minutes, hour, 4 hours
// and 24 hours, from the trades already in memory (no Koios request).
func (t *Tracker) Periods(id string) (Quote, []Period, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	now := t.now()
	for _, s := range t.states {
		if !strings.EqualFold(s.slug, id) {
			continue
		}
		return s.quote, periodsOf(s, now), nil
	}
	return Quote{}, nil, fmt.Errorf("unknown token %q", id)
}

// periodsOf works out a token's 5m, 1h, 4h and 24h trading from its points
// and trades. Call with t.mu held.
func periodsOf(s *tokenState, now time.Time) []Period {
	out := make([]Period, len(PeriodList))
	for i, pl := range PeriodList {
		cut := now.Add(-pl.Span)
		p := Period{Key: pl.Key, ChangePct: changePct(s.points, cut.Unix()), VolumeADA: volumeSince(s.points, cut.Unix())}
		for j := len(s.trades) - 1; j >= 0 && !s.trades[j].Time.Before(cut); j-- {
			if tr := s.trades[j]; tr.Side == "buy" {
				p.Buys++
				p.BuyADA += tr.AmountADA
			} else {
				p.Sells++
				p.SellADA += tr.AmountADA
			}
		}
		out[i] = p
	}
	return out
}
