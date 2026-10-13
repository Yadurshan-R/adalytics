package market

import (
	"fmt"
	"strings"
	"time"
)

// CandleSize is how long one candle lasts on the candlestick chart, and how
// far back the chart goes at that size. Like TradingView, the visitor picks
// the size and scrolls back through all the history loaded.
type CandleSize struct {
	Key      string
	Interval time.Duration // one candle
	Span     time.Duration // how far back the chart goes
}

// CandleSizes is every candle size, shortest first. 15 minute candles stop at
// 7 days so the chart stays a sensible size (about 670 candles).
var CandleSizes = []CandleSize{
	{Key: "15m", Interval: 15 * time.Minute, Span: 7 * 24 * time.Hour},
	{Key: "1h", Interval: time.Hour, Span: 30 * 24 * time.Hour},
	{Key: "4h", Interval: 4 * time.Hour, Span: 30 * 24 * time.Hour},
	{Key: "1D", Interval: 24 * time.Hour, Span: 30 * 24 * time.Hour},
}

// CandleSizeByKey finds a candle size by its key ("15m", "1h", "4h", "1D"),
// ignoring case.
func CandleSizeByKey(key string) (CandleSize, bool) {
	for _, c := range CandleSizes {
		if strings.EqualFold(c.Key, key) {
			return c, true
		}
	}
	return CandleSize{}, false
}

// Candles returns one token's candles at the given size, oldest first. While
// older history is still loading it returns what is loaded so far, so the
// chart grows instead of waiting for all 30 days.
func (t *Tracker) Candles(id, sizeKey string) (Quote, CandleSize, []Candle, error) {
	size, ok := CandleSizeByKey(sizeKey)
	if !ok {
		return Quote{}, CandleSize{}, nil, fmt.Errorf("unknown candle size %q: use 15m, 1h, 4h or 1D", sizeKey)
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, s := range t.states {
		if !strings.EqualFold(s.slug, id) {
			continue
		}
		if !t.loaded {
			return s.quote, size, nil, ErrRangeLoading
		}
		now := t.now()
		from := now.Add(-min(size.Span, t.span))
		if !t.historyDone && t.coveredFrom.After(from) {
			from = t.coveredFrom
		}
		return s.quote, size, candles(s.points, from, now, size.Interval), nil
	}
	return Quote{}, size, nil, fmt.Errorf("unknown token %q", id)
}
