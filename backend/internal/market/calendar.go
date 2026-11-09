package market

import "time"

// Calendar candles: a week candle starts on Monday at 00:00 UTC and a month
// candle on the 1st at 00:00 UTC, the way Birdeye, Binance and TradingView's
// crypto charts draw 1W and 1M. Their length varies (a month has 28 to 31
// days), so they cannot be cut in fixed steps like the other sizes.

// Calendar kinds of a CandleSize.
const (
	Week  = "week"
	Month = "month"
)

// weekStart is the Monday 00:00 UTC that starts t's week.
func weekStart(t time.Time) time.Time {
	d := t.UTC().Truncate(24 * time.Hour)
	back := (int(d.Weekday()) + 6) % 7 // days since Monday
	return d.AddDate(0, 0, -back)
}

// monthStart is the 1st of t's month, 00:00 UTC.
func monthStart(t time.Time) time.Time {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), 1, 0, 0, 0, 0, time.UTC)
}

// candleStarts lists the start times (Unix seconds) of the candles between
// from and to, oldest first.
//
// Every size starts at its first step at or after from: Fixed sizes at a
// multiple of their length, week and month candles at the first Monday or
// 1st. A candle that began before the history loaded would be missing its
// first part, so its open, high and low would be wrong. The last candle is
// the one still running, as on every chart.
func candleStarts(size CandleSize, from, to time.Time) []int64 {
	var starts []int64
	switch size.Calendar {
	case Week, Month:
		next := func(t time.Time) time.Time { return t.AddDate(0, 0, 7) }
		first := weekStart(from)
		if size.Calendar == Month {
			next = func(t time.Time) time.Time { return t.AddDate(0, 1, 0) }
			first = monthStart(from)
		}
		if first.Before(from) {
			first = next(first)
		}
		for t := first; !t.After(to); t = next(t) {
			starts = append(starts, t.Unix())
		}
	default:
		step := int64(size.Interval / time.Second)
		first := (from.Unix() + step - 1) / step * step // round up to a whole step
		for t := first; t <= to.Unix(); t += step {
			starts = append(starts, t)
		}
	}
	return starts
}
